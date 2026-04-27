package boards

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	Board "MavlinkProject/Server/backend/Shared/Boards"
)

const (
	H264StartCode1 = 0x00
	H264StartCode2 = 0x00
	H264StartCode3 = 0x01
	H264StartCode4 = 0x01
)

type RTMPTranslator struct {
	mu            sync.RWMutex
	listener      net.Listener
	streamManager *Board.LiveStreamManager
	running       bool
	stopChan      chan struct{}
	connections   map[string]net.Conn
}

type H264Frame struct {
	Timestamp uint32
	Data      []byte
	NALUType  uint8
}

func NewRTMPTranslator(streamManager *Board.LiveStreamManager) *RTMPTranslator {
	return &RTMPTranslator{
		streamManager: streamManager,
		running:       false,
		stopChan:      make(chan struct{}),
		connections:   make(map[string]net.Conn),
	}
}

func (rt *RTMPTranslator) Start(listenAddr string) error {
	rt.mu.Lock()
	if rt.running {
		rt.mu.Unlock()
		return fmt.Errorf("RTMPTranslator already running")
	}
	rt.running = true
	rt.mu.Unlock()

	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %v", listenAddr, err)
	}
	rt.listener = listener

	go rt.acceptLoop()

	log.Printf("[RTMPTranslator] Started listening on %s", listenAddr)
	return nil
}

func (rt *RTMPTranslator) Stop() {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if !rt.running {
		return
	}

	rt.running = false
	close(rt.stopChan)

	if rt.listener != nil {
		rt.listener.Close()
	}

	for _, conn := range rt.connections {
		conn.Close()
	}
	rt.connections = make(map[string]net.Conn)

	log.Printf("[RTMPTranslator] Stopped")
}

func (rt *RTMPTranslator) acceptLoop() {
	for {
		select {
		case <-rt.stopChan:
			return
		default:
		}

		rt.listener.(*net.TCPListener).SetDeadline(time.Now().Add(1 * time.Second))

		conn, err := rt.listener.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			log.Printf("[RTMPTranslator] Accept error: %v", err)
			continue
		}

		go rt.handleConnection(conn)
	}
}

func (rt *RTMPTranslator) handleConnection(conn net.Conn) {
	connID := fmt.Sprintf("%s_%d", conn.RemoteAddr().String(), time.Now().UnixNano())

	rt.mu.Lock()
	rt.connections[connID] = conn
	rt.mu.Unlock()

	defer func() {
		conn.Close()
		rt.mu.Lock()
		delete(rt.connections, connID)
		rt.mu.Unlock()
	}()

	log.Printf("[RTMPTranslator] New connection from %s (ID: %s)", conn.RemoteAddr(), connID)

	reader := bufio.NewReader(conn)

	for {
		select {
		case <-rt.stopChan:
			return
		default:
		}

		conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		frame, err := rt.readH264Frame(reader)
		if err != nil {
			if err == io.EOF {
				log.Printf("[RTMPTranslator] Connection %s closed by client", connID)
			} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			} else {
				log.Printf("[RTMPTranslator] Read error from %s: %v", connID, err)
			}
			return
		}

		if frame == nil {
			continue
		}

		rt.streamManager.WriteFrameToAll(frame.Data)
	}
}

func (rt *RTMPTranslator) readH264Frame(reader *bufio.Reader) (*H264Frame, error) {
	header := make([]byte, 4)
	_, err := io.ReadFull(reader, header)
	if err != nil {
		return nil, err
	}

	var frameSize uint32
	if header[0] == H264StartCode1 && header[1] == H264StartCode2 && header[2] == H264StartCode3 {
		frameSize = binary.BigEndian.Uint32(header[3:4])
		header = header[:3]
	} else if header[0] == H264StartCode1 && header[1] == H264StartCode2 && header[2] == H264StartCode3 && header[3] == H264StartCode4 {
		frameSize = binary.BigEndian.Uint32(header[4:8])
		header = header[:4]
	} else if header[0]&0x1F != 0 && header[0]&0x1F != 1 && header[0]&0x1F != 5 && header[0]&0x1F != 6 && header[0]&0x1F != 7 && header[0]&0x1F != 8 {
		return nil, fmt.Errorf("invalid H.264 start code")
	}

	if frameSize > 2*1024*1024 {
		return nil, fmt.Errorf("frame too large: %d bytes", frameSize)
	}

	data := make([]byte, frameSize)
	_, err = io.ReadFull(reader, data)
	if err != nil {
		return nil, err
	}

	frameData := append(header, data...)
	naluType := frameData[len(frameData)-1] & 0x1F

	return &H264Frame{
		Timestamp: uint32(time.Now().UnixMilli()),
		Data:      frameData,
		NALUType:  naluType,
	}, nil
}

func (rt *RTMPTranslator) GetStatus() map[string]interface{} {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	return map[string]interface{}{
		"running":      rt.running,
		"connections":  len(rt.connections),
		"listenerAddr": rt.listener.Addr().String(),
	}
}

type FFmpegProcess struct {
	mu        sync.RWMutex
	cmd       string
	args      []string
	taskCode  string
	centralID string
	running   bool
	stopChan  chan struct{}
}

func NewFFmpegProcess(taskCode, centralID string) *FFmpegProcess {
	return &FFmpegProcess{
		taskCode:  taskCode,
		centralID: centralID,
		stopChan:  make(chan struct{}),
	}
}

type FFmpegConfig struct {
	ListenAddr   string
	RTMPURL      string
	VideoCodec   string
	OutputFormat string
	LowLatency   bool
	BufferSize   int
}

func DefaultFFmpegConfig() *FFmpegConfig {
	return &FFmpegConfig{
		ListenAddr:   "127.0.0.1:8554",
		VideoCodec:   "libx264",
		OutputFormat: "h264",
		LowLatency:   true,
		BufferSize:   8192,
	}
}

func (f *FFmpegProcess) Start(config *FFmpegConfig) error {
	if config == nil {
		config = DefaultFFmpegConfig()
	}

	args := []string{
		"-fflags", "nobuffer",
		"-flags", "low_delay",
		"-i", config.RTMPURL,
		"-c:v", config.VideoCodec,
		"-preset", "ultrafast",
		"-tune", "zerolatency",
		"-b:v", "1000k",
		"-max_delay", "500000",
		"-bufsize", "1000k",
		"-f", config.OutputFormat,
		"-listen", "1",
		"tcp://" + config.ListenAddr,
	}

	f.cmd = "ffmpeg"
	f.args = args

	log.Printf("[FFmpegProcess] Starting: ffmpeg %v", args)

	go f.run()

	return nil
}

func (f *FFmpegProcess) run() {
	f.mu.Lock()
	f.running = true
	f.mu.Unlock()

	defer func() {
		f.mu.Lock()
		f.running = false
		f.mu.Unlock()
		close(f.stopChan)
	}()

	for {
		select {
		case <-f.stopChan:
			log.Printf("[FFmpegProcess] Stopped for task %s", f.taskCode)
			return
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (f *FFmpegProcess) Stop() {
	f.mu.RLock()
	if !f.running {
		f.mu.RUnlock()
		return
	}
	f.mu.RUnlock()

	close(f.stopChan)
	log.Printf("[FFmpegProcess] Stopping for task %s", f.taskCode)
}

func (f *FFmpegProcess) IsRunning() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.running
}
