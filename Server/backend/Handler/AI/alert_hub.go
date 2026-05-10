package AI

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	Models "MavlinkProject/Models"
)

type AlertHub struct {
	clients      map[*AlertClient]bool
	broadcast    chan *Models.AlertJSON
	register     chan *AlertClient
	unregister   chan *AlertClient
	sseListeners []*SSEListener
	mu           sync.RWMutex
}

type AlertClient struct {
	hub  *AlertHub
	conn *websocket.Conn
	send chan []byte
}

type SSEListener struct {
	ch      chan *Models.AlertJSON
	cleanup chan struct{}
}

var (
	globalHub *AlertHub
	hubOnce   sync.Once
)

func GetAlertHub() *AlertHub {
	hubOnce.Do(func() {
		globalHub = &AlertHub{
			clients:      make(map[*AlertClient]bool),
			broadcast:    make(chan *Models.AlertJSON, 256),
			register:     make(chan *AlertClient),
			unregister:   make(chan *AlertClient),
			sseListeners: make([]*SSEListener, 0),
		}
		go globalHub.run()
	})
	return globalHub
}

func (h *AlertHub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("[AlertHub] 客户端连接: 当前连接数=%d", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("[AlertHub] 客户端断开: 当前连接数=%d", len(h.clients))

		case alert := <-h.broadcast:
			data, err := json.Marshal(alert)
			if err != nil {
				log.Printf("[AlertHub] 告警序列化失败: %v", err)
				continue
			}

			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- data:
				default:
					h.mu.RUnlock()
					h.mu.Lock()
					delete(h.clients, client)
					close(client.send)
					h.mu.Unlock()
					h.mu.RLock()
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *AlertHub) Broadcast(alert *Models.AlertJSON) {
	select {
	case h.broadcast <- alert:
	default:
		log.Printf("[AlertHub] 广播通道已满，丢弃告警: %s", alert.AlertID)
	}
	h.BroadcastSSE(alert)
}

func (h *AlertHub) BroadcastSSE(alert *Models.AlertJSON) {
	h.mu.RLock()
	for _, listener := range h.sseListeners {
		select {
		case listener.ch <- alert:
		default:
		}
	}
	h.mu.RUnlock()
}

func (h *AlertHub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

var alertUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == "" ||
			origin == "https://www.deeppluse.dpdns.org" ||
			origin == "http://www.deeppluse.dpdns.org" ||
			origin == "https://deeppluse.dpdns.org" ||
			origin == "http://localhost:3000" ||
			origin == "http://localhost:8080" ||
			origin == "https://localhost:3000"
	},
}

func HandleAlertWebSocket(c *gin.Context) {
	conn, err := alertUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[AlertHub] WebSocket 升级失败: %v", err)
		return
	}

	hub := GetAlertHub()
	client := &AlertClient{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
	}

	hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *AlertClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *AlertClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func HandleAlertSSE(c *gin.Context) {
	hub := GetAlertHub()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	alertChan := make(chan *Models.AlertJSON, 64)
	cleanup := make(chan struct{})

	listener := &SSEListener{
		ch:      alertChan,
		cleanup: cleanup,
	}

	hub.mu.Lock()
	hub.sseListeners = append(hub.sseListeners, listener)
	hub.mu.Unlock()

	defer func() {
		close(cleanup)
		hub.mu.Lock()
		for i, l := range hub.sseListeners {
			if l == listener {
				hub.sseListeners = append(hub.sseListeners[:i], hub.sseListeners[i+1:]...)
				break
			}
		}
		hub.mu.Unlock()
	}()

	c.Stream(func(w io.Writer) bool {
		select {
		case alert := <-alertChan:
			data, _ := json.Marshal(alert)
			c.SSEvent("alert", string(data))
			return true
		case <-c.Request.Context().Done():
			return false
		case <-time.After(30 * time.Second):
			c.SSEvent("heartbeat", gin.H{"timestamp": time.Now().Unix()})
			return true
		}
	})
}
