package Boards

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	Board "MavlinkProject/Server/backend/Shared/Boards"
)

var (
	globalLiveStreamManager *Board.LiveStreamManager
	liveStreamOnce          sync.Once
)

func GetLiveStreamManager() *Board.LiveStreamManager {
	liveStreamOnce.Do(func() {
		globalLiveStreamManager = Board.NewLiveStreamManager()
	})
	return globalLiveStreamManager
}

func generateStreamID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return "stream_" + hex.EncodeToString(b)
}

func generateViewerID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "viewer_" + hex.EncodeToString(b)
}

type LiveStreamHandler struct{}

func NewLiveStreamHandler() *LiveStreamHandler {
	return &LiveStreamHandler{}
}

// HandleCentralUpload 处理 Central 上传的视频流
//
// Central 通过此接口上传视频流，请求格式：
//
// Content-Type: multipart/form-data 或 application/octet-stream
//
// Headers (必须):
//   - X-Task-Code: 任务代码（与原任务链相同）
//   - X-Central-ID: Central 设备 ID
//   - X-Drone-ID: 无人机 ID（可选）
//   - X-Video-Codec: 视频编码格式 (h264/h265/mjpeg)
//   - X-Audio-Codec: 音频编码格式 (aac/pcm)
//   - X-Resolution: 分辨率 (如 "1920x1080")
//   - X-FPS: 帧率 (如 "30")
//
// Body: 视频二进制数据流（H.264/H.265 NALU 单元或 FLV 容器格式）
//
// 响应:
//   - 200: 接收成功，返回 StreamID
//   - 400: 参数错误
//   - 500: 服务器内部错误
func (h *LiveStreamHandler) HandleCentralUpload(c *gin.Context) {
	taskCode := c.GetHeader("X-Task-Code")
	centralID := c.GetHeader("X-Central-ID")
	droneID := c.GetHeader("X-Drone-ID")
	videoCodec := c.GetHeader("X-Video-Codec")
	audioCodec := c.GetHeader("X-Audio-Codec")
	resolution := c.GetHeader("X-Resolution")
	fpsStr := c.GetHeader("X-FPS")

	if taskCode == "" || centralID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "缺少必要参数: X-Task-Code 和 X-Central-ID 必填",
			"code":    "MISSING_REQUIRED_PARAMS",
		})
		return
	}

	if videoCodec == "" {
		videoCodec = "h264"
	}
	if resolution == "" {
		resolution = "1920x1080"
	}
	fps := 30
	if fpsStr != "" {
		if v, err := strconv.Atoi(fpsStr); err == nil && v > 0 {
			fps = v
		}
	}

	streamID := generateStreamID()
	now := time.Now()

	streamInfo := &Board.LiveStreamInfo{
		StreamID:       streamID,
		TaskCode:       taskCode,
		CentralID:      centralID,
		DroneID:        droneID,
		StreamStatus:   Board.StreamStatus_Connected,
		VideoCodec:     Board.VideoCodec(videoCodec),
		AudioCodec:     Board.AudioCodec(audioCodec),
		Resolution:     resolution,
		FPS:            fps,
		Bitrate:        0,
		Duration:       0,
		StartTime:      now,
		LastUpdateTime: now,
		ViewerCount:    0,
	}

	manager := GetLiveStreamManager()

	existingStream, exists := manager.GetStreamByTask(taskCode)
	if exists {
		streamID = existingStream.Info.StreamID
		existingStream.mu.Lock()
		existingStream.Info.LastUpdateTime = now
		existingStream.Info.StreamStatus = Board.StreamStatus_Connected
		existingStream.mu.Unlock()
	} else {
		manager.CreateStream(streamInfo)
	}

	reader := c.Request.Body
	bufferSize := 4096
	buf := make([]byte, bufferSize)

	totalBytes := int64(0)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			data := buf[:n]
			totalBytes += int64(n)

			if stream, ok := manager.GetStream(streamID); ok {
				stream.WriteData(data)
				stream.mu.Lock()
				stream.Info.Bitrate = totalBytes
				stream.Info.Duration = int64(time.Since(now).Seconds())
				stream.Info.LastUpdateTime = time.Now()
				stream.mu.Unlock()
			}
		}

		if err != nil {
			if err != io.EOF {
				log.Printf("[LiveStream] 读取流数据错误: %v", err)
			}
			break
		}
	}

	if stream, ok := manager.GetStream(streamID); ok {
		stream.mu.Lock()
		stream.Info.StreamStatus = Board.StreamStatus_Disconnected
		stream.Info.LastUpdateTime = time.Now()
		stream.mu.Unlock()
	}

	log.Printf("[LiveStream] 流接收完成: stream_id=%s, task_code=%s, total_bytes=%d",
		streamID, taskCode, totalBytes)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"stream_id": streamID,
			"task_code": taskCode,
			"bytes_received": totalBytes,
			"message": "视频流接收成功",
		},
	})
}

// HandleCentralUploadWithMetadata 使用 BoardMessage 格式上传视频流
//
// Central 可以通过 JSON 元数据 + 二进制流的混合方式上传
//
// 请求格式:
// Content-Type: multipart/form-data
//
// Parts:
//   - metadata: JSON 字符串 (LiveStreamRequest 结构体)
//   - stream_data: 视频二进制数据
//
// metadata 示例:
//
//	{
//	  "message_id": "msg_live_001",
//	  "message_time": 1714234567,
//	  "message": {
//	    "message_type": "Request",
//	    "attribute": "Mission",
//	    "connection": "HTTPS",
//	    "command": "VideoStream",
//	    "data": {
//	      "task_code": "TASK_20260427_001",
//	      "video_codec": "h264",
//	      "resolution": "1920x1080",
//	      "fps": 30
//	    }
//	  },
//	  "from_id": "central_001",
//	  "from_type": "Central",
//	  "to_id": "backend_001",
//	  "to_type": "Backend"
//	}
func (h *LiveStreamHandler) HandleCentralUploadWithMetadata(c *gin.Context) {
	metadataForm, err := c.FormFile("metadata")
	if err != nil {
		h.HandleCentralUpload(c)
		return
	}

	metadataFile, err := metadataForm.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无法读取元数据",
			"code":    "METADATA_READ_ERROR",
		})
		return
	}
	defer metadataFile.Close()

	var req Board.LiveStreamRequest
	if err := json.NewDecoder(metadataFile).Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "元数据 JSON 解析失败: " + err.Error(),
			"code":    "METADATA_PARSE_ERROR",
		})
		return
	}

	taskCode, _ := req.Message.Data["task_code"].(string)
	videoCodec, _ := req.Message.Data["video_codec"].(string)
	audioCodec, _ := req.Message.Data["audio_codec"].(string)
	resolution, _ := req.Message.Data["resolution"].(string)
	droneID, _ := req.Message.Data["drone_id"].(string)

	fps := 30
	if fpsVal, ok := req.Message.Data["fps"].(float64); ok {
		fps = int(fpsVal)
	}

	if taskCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "metadata 中缺少 task_code",
			"code":    "MISSING_TASK_CODE",
		})
		return
	}

	c.Request.Header.Set("X-Task-Code", taskCode)
	c.Request.Header.Set("X-Central-ID", req.FromID)
	c.Request.Header.Set("X-Drone-ID", droneID)
	c.Request.Header.Set("X-Video-Codec", videoCodec)
	c.Request.Header.Set("X-Audio-Codec", audioCodec)
	c.Request.Header.Set("X-Resolution", resolution)
	c.Request.Header.Set("X-FPS", strconv.Itoa(fps))

	h.HandleCentralUpload(c)
}

// HandleFrontendGetStream 为前端提供视频流获取接口
//
// 前端通过此接口获取实时视频流
//
// 请求:
//   GET /api/backend/live?stream_id=xxx&task_code=xxx
//
// Query Parameters:
//   - stream_id: 流 ID（可选，优先使用）
//   - task_code: 任务代码（可选，作为备选）
//   - format: 输出格式 (raw/mjpeg/flv) 默认 raw
//
// 响应:
//   - 200: 视频二进制流 (Content-Type: video/mp4 或 image/jpeg)
//   - 404: 流不存在
//   - 406: 不支持的格式
func (h *LiveStreamHandler) HandleFrontendGetStream(c *gin.Context) {
	streamID := c.Query("stream_id")
	taskCode := c.Query("task_code")
	format := c.DefaultQuery("format", "mjpeg")

	manager := GetLiveStreamManager()
	var stream *Board.ActiveStream
	var exists bool

	if streamID != "" {
		stream, exists = manager.GetStream(streamID)
	} else if taskCode != "" {
		stream, exists = manager.GetStreamByTask(taskCode)
	}

	if !exists || stream == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "视频流不存在或已结束",
			"code":    "STREAM_NOT_FOUND",
		})
		return
	}

	viewerID := generateViewerID()
	ch := stream.AddViewer(viewerID)
	defer stream.RemoveViewer(viewerID)

	switch format {
	case "mjpeg":
		c.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")
		c.Header().Set("Cache-Control", "no-cache")
		c.Header().Set("Connection", "keep-alive")

		writer := c.Writer
		for data := range ch {
			_, err := writer.Write([]byte("--frame\r\n"))
			if err != nil {
				break
			}
			_, err = writer.Write([]byte("Content-Type: image/jpeg\r\n\r\n"))
			if err != nil {
				break
			}
			_, err = writer.Write(data)
			if err != nil {
				break
			}
			_, err = writer.Write([]byte("\r\n"))
			if err != nil {
				break
			}
			writer.Flush()
		}

	case "raw":
		c.Header().Set("Content-Type", "application/octet-stream")
		c.Header().Set("Cache-Control", "no-cache")

		writer := c.Writer
		for data := range ch {
			_, err := writer.Write(data)
			if err != nil {
				break
			}
			writer.Flush()
		}

	case "flv":
		c.Header().Set("Content-Type", "video/x-flv")
		c.Header().Set("Cache-Control", "no-cache")

		writer := c.Writer
		writer.Write([]byte("FLV\x01\x01\x00\x00\x00\x09"))

		for data := range ch {
			_, err := writer.Write(data)
			if err != nil {
				break
			}
			writer.Flush()
		}

	default:
		c.JSON(http.StatusNotAcceptable, gin.H{
			"success": false,
			"error":   "不支持的格式: " + format,
			"code":    "UNSUPPORTED_FORMAT",
		})
		return
	}
}

// HandleFrontendWebSocket WebSocket 方式提供视频流
//
// 前端通过 WebSocket 连接接收视频帧
//
// 连接:
//   WS /api/backend/live/ws?stream_id=xxx
//
// 消息格式 (服务端→前端):
//   type: binary (视频帧数据) 或 text (控制消息)
//
// 控制消息示例:
//   {"type":"info","data":{"stream_id":"...","status":"connected"}}
//   {"type":"error","message":"..."}
func (h *LiveStreamHandler) HandleFrontendWebSocket(c *gin.Context) {
	streamID := c.Query("stream_id")
	taskCode := c.Query("task_code")

	manager := GetLiveStreamManager()
	var stream *Board.ActiveStream
	var exists bool

	if streamID != "" {
		stream, exists = manager.GetStream(streamID)
	} else if taskCode != "" {
		stream, exists = manager.GetStreamByTask(taskCode)
	}

	if !exists || stream == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "视频流不存在",
			"code":    "STREAM_NOT_FOUND",
		})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "WebSocket 升级失败",
			"code":    "WS_UPGRADE_ERROR",
		})
		return
	}
	defer conn.Close()

	viewerID := generateViewerID()
	ch := stream.AddViewer(viewerID)
	defer stream.RemoveViewer(viewerID)

	infoMsg, _ := json.Marshal(map[string]interface{}{
		"type": "info",
		"data": stream.Info,
	})
	conn.WriteMessage(websocket.TextMessage, infoMsg)

	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()

	for data := range ch {
		err := conn.WriteMessage(websocket.BinaryMessage, data)
		if err != nil {
			break
		}
	}
}

// HandleListStreams 获取当前活跃的视频流列表
//
// 请求:
//   GET /api/backend/live/list
//
// 响应:
//   - 200: 流列表
func (h *LiveStreamHandler) HandleListStreams(c *gin.Context) {
	manager := GetLiveStreamManager()
	streams := manager.ListStreams()

	c.JSON(http.StatusOK, Board.LiveStreamListResponse{
		Success: true,
		Message: fmt.Sprintf("共 %d 个活跃流", len(streams)),
		Data:    streams,
	})
}

// HandleGetStreamInfo 获取指定流的详细信息
//
// 请求:
//   GET /api/backend/live/info/:stream_id
//
// 响应:
//   - 200: 流信息
//   - 404: 流不存在
func (h *LiveStreamHandler) HandleGetStreamInfo(c *gin.Context) {
	streamID := c.Param("stream_id")

	manager := GetLiveStreamManager()
	stream, exists := manager.GetStream(streamID)

	if !exists || stream == nil {
		c.JSON(http.StatusNotFound, Board.LiveStreamResponse{
			Success: false,
			Error:   "视频流不存在",
		})
		return
	}

	c.JSON(http.StatusOK, Board.LiveStreamResponse{
		Success: true,
		Message: "获取成功",
		Data:    stream.Info,
	})
}

// HandleStopStream 停止指定的视频流
//
// 请求:
//   DELETE /api/backend/live/:stream_id
//
// 权限: 需要管理员权限
func (h *LiveStreamHandler) HandleStopStream(c *gin.Context) {
	streamID := c.Param("stream_id")

	manager := GetLiveStreamManager()
	stream, exists := manager.GetStream(streamId)

	if !exists || stream == nil {
		c.JSON(http.StatusNotFound, Board.LiveStreamResponse{
			Success: false,
			Error:   "视频流不存在",
		})
		return
	}

	manager.RemoveStream(streamID)

	c.JSON(http.StatusOK, Board.LiveStreamResponse{
		Success: true,
		Message: "视频流已停止",
	})
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 8192,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return strings.HasSuffix(origin, ".deeppluse.dpdns.org") ||
			strings.HasPrefix(origin, "http://localhost") ||
			strings.HasPrefix(origin, "https://localhost")
	},
}
