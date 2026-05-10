package AI

import (
	"encoding/json"
	"testing"
	"time"

	Models "MavlinkProject/Models"
)

// ==================== AlertHub Broadcast 测试 ====================

func TestAlertHubBroadcast(t *testing.T) {
	hub := &AlertHub{
		clients:      make(map[*AlertClient]bool),
		broadcast:    make(chan *Models.AlertJSON, 256),
		register:     make(chan *AlertClient),
		unregister:   make(chan *AlertClient),
		sseListeners: make([]*SSEListener, 0),
	}
	go hub.run()

	// 创建模拟客户端
	client := &AlertClient{
		hub:  hub,
		send: make(chan []byte, 256),
	}

	// 注册客户端
	hub.register <- client
	time.Sleep(50 * time.Millisecond)

	if hub.ClientCount() != 1 {
		t.Errorf("注册后客户端数量应该为 1, 实际 %d", hub.ClientCount())
	}

	// 广播告警
	alert := &Models.AlertJSON{
		AlertID:     "test_broadcast_alert",
		AlertType:   "anomaly",
		Severity:    Models.SeverityHigh,
		AnomalyType: Models.AnomalyFire,
		Source:      Models.SourceSensor,
		Timestamp:   time.Now().Unix(),
		Confidence:  0.95,
	}

	hub.Broadcast(alert)

	// 等待接收广播消息
	select {
	case data := <-client.send:
		var received Models.AlertJSON
		if err := json.Unmarshal(data, &received); err != nil {
			t.Fatalf("反序列化广播数据失败: %v", err)
		}
		if received.AlertID != "test_broadcast_alert" {
			t.Errorf("AlertID 不匹配: 期望 test_broadcast_alert, 实际 %s", received.AlertID)
		}
		if received.Severity != Models.SeverityHigh {
			t.Errorf("Severity 不匹配: 期望 %s, 实际 %s", Models.SeverityHigh, received.Severity)
		}
	case <-time.After(2 * time.Second):
		t.Error("等待广播消息超时")
	}
}

func TestAlertHubBroadcastToMultipleClients(t *testing.T) {
	hub := &AlertHub{
		clients:      make(map[*AlertClient]bool),
		broadcast:    make(chan *Models.AlertJSON, 256),
		register:     make(chan *AlertClient),
		unregister:   make(chan *AlertClient),
		sseListeners: make([]*SSEListener, 0),
	}
	go hub.run()

	// 创建多个模拟客户端
	client1 := &AlertClient{hub: hub, send: make(chan []byte, 256)}
	client2 := &AlertClient{hub: hub, send: make(chan []byte, 256)}
	client3 := &AlertClient{hub: hub, send: make(chan []byte, 256)}

	hub.register <- client1
	hub.register <- client2
	hub.register <- client3
	time.Sleep(50 * time.Millisecond)

	if hub.ClientCount() != 3 {
		t.Errorf("注册 3 个客户端后数量应该为 3, 实际 %d", hub.ClientCount())
	}

	alert := &Models.AlertJSON{
		AlertID:   "multi_broadcast",
		Severity:  Models.SeverityCritical,
		Timestamp: time.Now().Unix(),
	}

	hub.Broadcast(alert)

	// 验证所有客户端都收到消息
	for i, client := range []*AlertClient{client1, client2, client3} {
		select {
		case data := <-client.send:
			var received Models.AlertJSON
			json.Unmarshal(data, &received)
			if received.AlertID != "multi_broadcast" {
				t.Errorf("客户端 %d: AlertID 不匹配", i)
			}
		case <-time.After(2 * time.Second):
			t.Errorf("客户端 %d: 等待广播消息超时", i)
		}
	}
}

func TestAlertHubUnregisterClient(t *testing.T) {
	hub := &AlertHub{
		clients:      make(map[*AlertClient]bool),
		broadcast:    make(chan *Models.AlertJSON, 256),
		register:     make(chan *AlertClient),
		unregister:   make(chan *AlertClient),
		sseListeners: make([]*SSEListener, 0),
	}
	go hub.run()

	client := &AlertClient{hub: hub, send: make(chan []byte, 256)}
	hub.register <- client
	time.Sleep(50 * time.Millisecond)

	if hub.ClientCount() != 1 {
		t.Errorf("注册后客户端数量应该为 1, 实际 %d", hub.ClientCount())
	}

	// 注销客户端
	hub.unregister <- client
	time.Sleep(50 * time.Millisecond)

	if hub.ClientCount() != 0 {
		t.Errorf("注销后客户端数量应该为 0, 实际 %d", hub.ClientCount())
	}
}

// ==================== AlertHub BroadcastSSE 测试 ====================

func TestAlertHubBroadcastSSE(t *testing.T) {
	hub := &AlertHub{
		clients:      make(map[*AlertClient]bool),
		broadcast:    make(chan *Models.AlertJSON, 256),
		register:     make(chan *AlertClient),
		unregister:   make(chan *AlertClient),
		sseListeners: make([]*SSEListener, 0),
	}

	// 添加 SSE 监听器
	listener := &SSEListener{
		ch:      make(chan *Models.AlertJSON, 64),
		cleanup: make(chan struct{}),
	}
	hub.sseListeners = append(hub.sseListeners, listener)

	alert := &Models.AlertJSON{
		AlertID:     "sse_test_alert",
		AlertType:   "anomaly",
		Severity:    Models.SeverityCritical,
		AnomalyType: Models.AnomalyGas,
		Source:      Models.SourceDrone,
		Timestamp:   time.Now().Unix(),
		Confidence:  0.88,
	}

	hub.BroadcastSSE(alert)

	select {
	case received := <-listener.ch:
		if received.AlertID != "sse_test_alert" {
			t.Errorf("SSE AlertID 不匹配: 期望 sse_test_alert, 实际 %s", received.AlertID)
		}
		if received.Severity != Models.SeverityCritical {
			t.Errorf("SSE Severity 不匹配: 期望 %s, 实际 %s", Models.SeverityCritical, received.Severity)
		}
		if received.AnomalyType != Models.AnomalyGas {
			t.Errorf("SSE AnomalyType 不匹配: 期望 %s, 实际 %s", Models.AnomalyGas, received.AnomalyType)
		}
	case <-time.After(2 * time.Second):
		t.Error("等待 SSE 消息超时")
	}
}

func TestAlertHubBroadcastSSEMultipleListeners(t *testing.T) {
	hub := &AlertHub{
		clients:      make(map[*AlertClient]bool),
		broadcast:    make(chan *Models.AlertJSON, 256),
		register:     make(chan *AlertClient),
		unregister:   make(chan *AlertClient),
		sseListeners: make([]*SSEListener, 0),
	}

	listener1 := &SSEListener{ch: make(chan *Models.AlertJSON, 64), cleanup: make(chan struct{})}
	listener2 := &SSEListener{ch: make(chan *Models.AlertJSON, 64), cleanup: make(chan struct{})}
	hub.sseListeners = append(hub.sseListeners, listener1, listener2)

	alert := &Models.AlertJSON{
		AlertID:   "multi_sse_alert",
		Severity:  Models.SeverityMedium,
		Timestamp: time.Now().Unix(),
	}

	hub.BroadcastSSE(alert)

	for i, listener := range []*SSEListener{listener1, listener2} {
		select {
		case received := <-listener.ch:
			if received.AlertID != "multi_sse_alert" {
				t.Errorf("SSE 监听器 %d: AlertID 不匹配", i)
			}
		case <-time.After(2 * time.Second):
			t.Errorf("SSE 监听器 %d: 等待消息超时", i)
		}
	}
}

func TestAlertHubBroadcastSSENoListeners(t *testing.T) {
	hub := &AlertHub{
		clients:      make(map[*AlertClient]bool),
		broadcast:    make(chan *Models.AlertJSON, 256),
		register:     make(chan *AlertClient),
		unregister:   make(chan *AlertClient),
		sseListeners: make([]*SSEListener, 0),
	}

	alert := &Models.AlertJSON{
		AlertID:   "no_listener_alert",
		Severity:  Models.SeverityInfo,
		Timestamp: time.Now().Unix(),
	}

	// 没有监听器时不应该 panic
	hub.BroadcastSSE(alert)
}

// ==================== SSEListener 测试 ====================

func TestSSEListenerChannel(t *testing.T) {
	listener := &SSEListener{
		ch:      make(chan *Models.AlertJSON, 64),
		cleanup: make(chan struct{}),
	}

	alert := &Models.AlertJSON{
		AlertID:   "listener_test",
		Severity:  Models.SeverityLow,
		Timestamp: time.Now().Unix(),
	}

	listener.ch <- alert

	select {
	case received := <-listener.ch:
		if received.AlertID != "listener_test" {
			t.Errorf("AlertID 不匹配: 期望 listener_test, 实际 %s", received.AlertID)
		}
	default:
		t.Error("应该能从 SSEListener 通道接收到消息")
	}
}

func TestSSEListenerCleanup(t *testing.T) {
	listener := &SSEListener{
		ch:      make(chan *Models.AlertJSON, 64),
		cleanup: make(chan struct{}),
	}

	// 关闭 cleanup 通道
	close(listener.cleanup)

	select {
	case <-listener.cleanup:
		// 成功接收到关闭信号
	default:
		t.Error("应该能从 cleanup 通道接收到关闭信号")
	}
}

// ==================== AlertHub ClientCount 测试 ====================

func TestAlertHubClientCount(t *testing.T) {
	hub := &AlertHub{
		clients:      make(map[*AlertClient]bool),
		broadcast:    make(chan *Models.AlertJSON, 256),
		register:     make(chan *AlertClient),
		unregister:   make(chan *AlertClient),
		sseListeners: make([]*SSEListener, 0),
	}
	go hub.run()

	if hub.ClientCount() != 0 {
		t.Errorf("初始客户端数量应该为 0, 实际 %d", hub.ClientCount())
	}

	client1 := &AlertClient{hub: hub, send: make(chan []byte, 256)}
	hub.register <- client1
	time.Sleep(50 * time.Millisecond)

	if hub.ClientCount() != 1 {
		t.Errorf("注册 1 个客户端后数量应该为 1, 实际 %d", hub.ClientCount())
	}

	client2 := &AlertClient{hub: hub, send: make(chan []byte, 256)}
	hub.register <- client2
	time.Sleep(50 * time.Millisecond)

	if hub.ClientCount() != 2 {
		t.Errorf("注册 2 个客户端后数量应该为 2, 实际 %d", hub.ClientCount())
	}
}

// ==================== Broadcast 包含 SSE 推送测试 ====================

func TestBroadcastIncludesSSE(t *testing.T) {
	hub := &AlertHub{
		clients:      make(map[*AlertClient]bool),
		broadcast:    make(chan *Models.AlertJSON, 256),
		register:     make(chan *AlertClient),
		unregister:   make(chan *AlertClient),
		sseListeners: make([]*SSEListener, 0),
	}
	go hub.run()

	// 添加 SSE 监听器
	listener := &SSEListener{
		ch:      make(chan *Models.AlertJSON, 64),
		cleanup: make(chan struct{}),
	}
	hub.mu.Lock()
	hub.sseListeners = append(hub.sseListeners, listener)
	hub.mu.Unlock()

	alert := &Models.AlertJSON{
		AlertID:   "broadcast_sse_test",
		Severity:  Models.SeverityHigh,
		Timestamp: time.Now().Unix(),
	}

	hub.Broadcast(alert)

	// Broadcast 应该同时推送到 SSE 监听器
	select {
	case received := <-listener.ch:
		if received.AlertID != "broadcast_sse_test" {
			t.Errorf("SSE AlertID 不匹配: 期望 broadcast_sse_test, 实际 %s", received.AlertID)
		}
	case <-time.After(2 * time.Second):
		t.Error("Broadcast 应该同时推送到 SSE 监听器, 等待超时")
	}
}
