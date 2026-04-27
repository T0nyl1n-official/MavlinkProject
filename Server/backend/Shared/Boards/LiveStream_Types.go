package boards

import (
	"sync"
	"time"
)

type StreamStatus string

const (
	StreamStatus_Connected    StreamStatus = "connected"
	StreamStatus_Disconnected StreamStatus = "disconnected"
	StreamStatus_Buffering     StreamStatus = "buffering"
	StreamStatus_Error         StreamStatus = "error"
)

type VideoCodec string

const (
	VideoCodec_H264  VideoCodec = "h264"
	VideoCodec_H265  VideoCodec = "h265"
	VideoCodec_MJPEG VideoCodec = "mjpeg"
)

type AudioCodec string

const (
	AudioCodec_AAC AudioCodec = "aac"
	AudioCodec_PCM AudioCodec = "pcm"
)

type LiveStreamInfo struct {
	StreamID        string      `json:"stream_id"`
	TaskCode        string      `json:"task_code"`
	CentralID       string      `json:"central_id"`
	DroneID         string      `json:"drone_id"`
	StreamStatus    StreamStatus `json:"status"`
	VideoCodec      VideoCodec   `json:"video_codec"`
	AudioCodec      AudioCodec   `json:"audio_codec"`
	Resolution      string      `json:"resolution"`
	FPS             int         `json:"fps"`
	Bitrate         int64       `json:"bitrate"`
	Duration        int64       `json:"duration"`
	StartTime       time.Time   `json:"start_time"`
	LastUpdateTime  time.Time   `json:"last_update_time"`
	ViewerCount     int         `json:"viewer_count"`
}

type LiveStreamRequest struct {
	MessageID   string                 `json:"message_id"`
	MessageTime int64                  `json:"message_time"`
	Message     LiveStreamMessageData  `json:"message"`
	FromID      string                 `json:"from_id"`
	FromType    string                 `json:"from_type"`
	ToID        string                 `json:"to_id"`
	ToType      string                 `json:"to_type"`
}

type LiveStreamMessageData struct {
	MessageType string                 `json:"message_type"`
	Attribute   string                 `json:"attribute"`
	Connection  string                 `json:"connection"`
	Command     string                 `json:"command"`
	Data        map[string]interface{} `json:"data"`
}

type LiveStreamMetadata struct {
	TaskCode   string    `json:"task_code"`
	CentralID  string    `json:"central_id"`
	DroneID    string    `json:"drone_id"`
	VideoCodec string    `json:"video_codec"`
	AudioCodec string    `json:"audio_codec"`
	Resolution string    `json:"resolution"`
	FPS        int       `json:"fps"`
	Timestamp  time.Time `json:"timestamp"`
}

type LiveStreamResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message,omitempty"`
	Error   string          `json:"error,omitempty"`
	Data    *LiveStreamInfo `json:"data,omitempty"`
}

type LiveStreamListResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message,omitempty"`
	Error   string            `json:"error,omitempty"`
	Data    []*LiveStreamInfo `json:"data,omitempty"`
}

type LiveStreamManager struct {
	mu           sync.RWMutex
	activeStreams map[string]*ActiveStream
	streamIndex  map[string]*ActiveStream
}

type ActiveStream struct {
	mu           sync.RWMutex
	Info         *LiveStreamInfo
	buffer       []byte
	viewers      map[string]chan []byte
	controlChan  chan int
	lastFrameTime time.Time
}

func NewLiveStreamManager() *LiveStreamManager {
	return &LiveStreamManager{
		activeStreams: make(map[string]*ActiveStream),
		streamIndex:   make(map[string]*ActiveStream),
	}
}

func (lm *LiveStreamManager) CreateStream(info *LiveStreamInfo) *ActiveStream {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	stream := &ActiveStream{
		Info:        info,
		buffer:      make([]byte, 0),
		viewers:     make(map[string]chan []byte),
		controlChan: make(chan int, 10),
	}

	lm.activeStreams[info.StreamID] = stream
	lm.streamIndex[info.TaskCode] = stream

	return stream
}

func (lm *LiveStreamManager) GetStream(streamID string) (*ActiveStream, bool) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	stream, ok := lm.activeStreams[streamID]
	return stream, ok
}

func (lm *LiveStreamManager) GetStreamByTask(taskCode string) (*ActiveStream, bool) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	stream, ok := lm.streamIndex[taskCode]
	return stream, ok
}

func (lm *LiveStreamManager) RemoveStream(streamID string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if stream, ok := lm.activeStreams[streamID]; ok {
		delete(lm.streamIndex, stream.Info.TaskCode)
		close(stream.controlChan)
		for _, viewer := range stream.viewers {
			close(viewer)
		}
		delete(lm.activeStreams, streamID)
	}
}

func (lm *LiveStreamManager) ListStreams() []*LiveStreamInfo {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	list := make([]*LiveStreamInfo, 0, len(lm.activeStreams))
	for _, stream := range lm.activeStreams {
		list = append(list, stream.Info)
	}
	return list
}

func (s *ActiveStream) WriteData(data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.buffer = append(s.buffer, data...)
	s.lastFrameTime = time.Now()

	const maxBufferSize = 10 * 1024 * 1024
	if len(s.buffer) > maxBufferSize {
		s.buffer = s.buffer[len(s.buffer)-maxBufferSize:]
	}

	for _, viewer := range s.viewers {
		select {
		case viewer <- data:
		default:
		}
	}
}

func (s *ActiveStream) AddViewer(viewerID string) chan []byte {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan []byte, 100)
	s.viewers[viewerID] = ch
	s.Info.ViewerCount = len(s.viewers)
	return ch
}

func (s *ActiveStream) RemoveViewer(viewerID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ch, ok := s.viewers[viewerID]; ok {
		close(ch)
		delete(s.viewers, viewerID)
	}
	s.Info.ViewerCount = len(s.viewers)
}

func (s *ActiveStream) GetBuffer() []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.buffer
}
