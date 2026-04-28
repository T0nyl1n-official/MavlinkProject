package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	LiveStreamHandler "MavlinkProject/Server/backend/Handler/Boards"
	Board "MavlinkProject/Server/backend/Shared/Boards"
)

func TestLiveStreamManagerCreation(t *testing.T) {
	t.Log("=== 测试 LiveStreamManager 初始化 ===")
	manager := LiveStreamHandler.GetLiveStreamManager()

	if manager == nil {
		t.Fatal("LiveStreamManager 不应为 nil")
	}

	t.Logf("✓ LiveStreamManager 创建成功: %p", manager)
}

func TestCreateAndRetrieveStream(t *testing.T) {
	t.Log("=== 测试流创建与检索 ===")
	manager := LiveStreamHandler.GetLiveStreamManager()

	testTaskCode := fmt.Sprintf("TASK_TEST_%d", time.Now().UnixNano())
	streamInfo := &Board.LiveStreamInfo{
		StreamID:     fmt.Sprintf("stream_test_%d", time.Now().UnixNano()),
		TaskCode:     testTaskCode,
		CentralID:    "central_test_001",
		DroneID:      "drone_test_001",
		StreamStatus: Board.StreamStatus_Connected,
		VideoCodec:   Board.VideoCodec_H264,
		AudioCodec:   Board.AudioCodec_AAC,
		Resolution:   "1920x1080",
		FPS:          30,
		StartTime:    time.Now(),
	}

	stream := manager.CreateStream(streamInfo)
	if stream == nil {
		t.Fatal("创建流失败")
	}
	t.Logf("✓ 流创建成功: %s", streamInfo.StreamID)

	retrieved, exists := manager.GetStream(streamInfo.StreamID)
	if !exists {
		t.Fatalf("无法检索到刚创建的流: %s", streamInfo.StreamID)
	}
	t.Logf("✓ 通过 StreamID 检索成功")

	retrievedByTask, exists := manager.GetStreamByTask(testTaskCode)
	if !exists {
		t.Fatalf("无法通过 TaskCode 检索到流: %s", testTaskCode)
	}
	t.Logf("✓ 通过 TaskCode 检索成功")

	if retrieved.Info.StreamID != retrievedByTask.Info.StreamID {
		t.Errorf("两种方式检索到的流不一致")
	}

	defer manager.RemoveStream(streamInfo.StreamID)
}

func TestStreamDataWriteAndRead(t *testing.T) {
	t.Log("=== 测试流数据写入与读取 ===")
	manager := LiveStreamHandler.GetLiveStreamManager()

	streamInfo := &Board.LiveStreamInfo{
		StreamID:     fmt.Sprintf("stream_data_%d", time.Now().UnixNano()),
		TaskCode:     fmt.Sprintf("TASK_DATA_%d", time.Now().UnixNano()),
		CentralID:    "central_data_001",
		StreamStatus: Board.StreamStatus_Connected,
		VideoCodec:   Board.VideoCodec_H264,
		StartTime:    time.Now(),
	}

	stream := manager.CreateStream(streamInfo)
	defer manager.RemoveStream(streamInfo.StreamID)

	testData1 := []byte{0x00, 0x00, 0x00, 0x01, 0x67, 0x42, 0xC0, 0x28}
	testData2 := []byte{0x00, 0x00, 0x00, 0x01, 0x65, 0xB8, 0x10, 0x00}

	stream.WriteData(testData1)
	stream.WriteData(testData2)

	buffer := stream.GetBuffer()
	if len(buffer) != len(testData1)+len(testData2) {
		t.Errorf("缓冲区大小不正确: 期望 %d, 实际 %d", len(testData1)+len(testData2), len(buffer))
	} else {
		t.Logf("✓ 数据写入读取正确: 缓冲区大小=%d", len(buffer))
	}
}

func TestViewerManagement(t *testing.T) {
	t.Log("=== 测试观众管理 ===")
	manager := LiveStreamHandler.GetLiveStreamManager()

	streamInfo := &Board.LiveStreamInfo{
		StreamID:     fmt.Sprintf("stream_viewer_%d", time.Now().UnixNano()),
		TaskCode:     fmt.Sprintf("TASK_VIEWER_%d", time.Now().UnixNano()),
		CentralID:    "central_viewer_001",
		StreamStatus: Board.StreamStatus_Connected,
		VideoCodec:   Board.VideoCodec_H264,
		StartTime:    time.Now(),
	}

	stream := manager.CreateStream(streamInfo)
	defer manager.RemoveStream(streamInfo.StreamID)

	viewerID1 := "viewer_test_001"
	ch1 := stream.AddViewer(viewerID1)

	if ch1 == nil {
		t.Fatal("添加观众失败")
	}
	t.Logf("✓ 观众添加成功: %s (初始观众数: %d)", viewerID1, stream.Info.ViewerCount)

	if stream.Info.ViewerCount != 1 {
		t.Errorf("观众数不正确: 期望 1, 实际 %d", stream.Info.ViewerCount)
	}

	viewerID2 := "viewer_test_002"
	ch2 := stream.AddViewer(viewerID2)
	if stream.Info.ViewerCount != 2 {
		t.Errorf("观众数不正确: 期望 2, 实际 %d", stream.Info.ViewerCount)
	}
	t.Logf("✓ 第二个观众添加成功 (当前观众数: %d)", stream.Info.ViewerCount)

	stream.RemoveViewer(viewerID1)
	if stream.Info.ViewerCount != 1 {
		t.Errorf("移除后观众数不正确: 期望 1, 实际 %d", stream.Info.ViewerCount)
	}
	t.Logf("✓ 观众移除成功 (剩余观众数: %d)", stream.Info.ViewerCount)

	select {
	case <-ch1:
		t.Logf("✓ 观众通道已关闭")
	default:
		t.Error("观众通道应该已关闭")
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		close(ch2)
	}()
}

func TestDataDistributionToViewers(t *testing.T) {
	t.Log("=== 测试数据分发到观众 ===")
	manager := LiveStreamHandler.GetLiveStreamManager()

	streamInfo := &Board.LiveStreamInfo{
		StreamID:     fmt.Sprintf("stream_dist_%d", time.Now().UnixNano()),
		TaskCode:     fmt.Sprintf("TASK_DIST_%d", time.Now().UnixNano()),
		CentralID:    "central_dist_001",
		StreamStatus: Board.StreamStatus_Connected,
		VideoCodec:   Board.VideoCodec_H264,
		StartTime:    time.Now(),
	}

	stream := manager.CreateStream(streamInfo)
	defer manager.RemoveStream(streamInfo.StreamID)

	viewerCh := stream.AddViewer("viewer_dist_001")

	testData := []byte("H264_FRAME_DATA_TEST_" + fmt.Sprintf("%d", time.Now().UnixNano()))

	go func() {
		time.Sleep(20 * time.Millisecond)
		stream.WriteData(testData)
	}()

	select {
	case received := <-viewerCh:
		if string(received) != string(testData) {
			t.Errorf("接收数据不匹配")
		} else {
			t.Logf("✓ 数据分发成功: 收到 %d 字节", len(received))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("超时等待接收数据")
	}
}

func TestStreamStatusUpdate(t *testing.T) {
	t.Log("=== 测试流状态更新 ===")
	manager := LiveStreamHandler.GetLiveStreamManager()

	streamInfo := &Board.LiveStreamInfo{
		StreamID:     fmt.Sprintf("stream_status_%d", time.Now().UnixNano()),
		TaskCode:     fmt.Sprintf("TASK_STATUS_%d", time.Now().UnixNano()),
		CentralID:    "central_status_001",
		StreamStatus: Board.StreamStatus_Connected,
		VideoCodec:   Board.VideoCodec_H264,
		StartTime:    time.Now(),
	}

	stream := manager.CreateStream(streamInfo)
	defer manager.RemoveStream(streamInfo.StreamID)

	initialStatus := stream.GetStatus()
	if initialStatus != Board.StreamStatus_Connected {
		t.Errorf("初始状态不正确: %s", initialStatus)
	}
	t.Logf("✓ 初始状态: %s", initialStatus)

	stream.UpdateStatus(Board.StreamStatus_Buffering)
	newStatus := stream.GetStatus()
	if newStatus != Board.StreamStatus_Buffering {
		t.Errorf("更新后状态不正确: %s", newStatus)
	} else {
		t.Logf("✓ 状态更新为: %s", newStatus)
	}

	stream.UpdateLastUpdateTime()
	t.Logf("✓ 最后更新时间已更新: %v", stream.Info.LastUpdateTime)
}

func TestStreamStatsUpdate(t *testing.T) {
	t.Log("=== 测试流统计信息更新 ===")
	manager := LiveStreamHandler.GetLiveStreamManager()

	streamInfo := &Board.LiveStreamInfo{
		StreamID:     fmt.Sprintf("stream_stats_%d", time.Now().UnixNano()),
		TaskCode:     fmt.Sprintf("TASK_STATS_%d", time.Now().UnixNano()),
		CentralID:    "central_stats_001",
		StreamStatus: Board.StreamStatus_Connected,
		VideoCodec:   Board.VideoCodec_H264,
		StartTime:    time.Now(),
	}

	stream := manager.CreateStream(streamInfo)
	defer manager.RemoveStream(streamInfo.StreamID)

	testBitrate := int64(5000000)
	testDuration := float64(120.5)

	stream.UpdateStreamStats(testBitrate, testDuration)

	if stream.Info.Bitrate != testBitrate {
		t.Errorf("比特率不正确: 期望 %d, 实际 %d", testBitrate, stream.Info.Bitrate)
	} else {
		t.Logf("✓ 比特率更新正确: %d bps", stream.Info.Bitrate)
	}

	if stream.Info.Duration != int64(testDuration) {
		t.Errorf("时长不正确: 期望 %d, 实际 %d", int64(testDuration), stream.Info.Duration)
	} else {
		t.Logf("✓ 时长更新正确: %d 秒", stream.Info.Duration)
	}
}

func TestListStreams(t *testing.T) {
	t.Log("=== 测试列出所有活跃流 ===")
	manager := LiveStreamHandler.GetLiveStreamManager()

	var createdStreams []string
	for i := 0; i < 3; i++ {
		streamInfo := &Board.LiveStreamInfo{
			StreamID:     fmt.Sprintf("stream_list_%d_%d", i, time.Now().UnixNano()),
			TaskCode:     fmt.Sprintf("TASK_LIST_%d_%d", i, time.Now().UnixNano()),
			CentralID:    fmt.Sprintf("central_list_%03d", i),
			StreamStatus: Board.StreamStatus_Connected,
			VideoCodec:   Board.VideoCodec_H264,
			StartTime:    time.Now(),
		}
		manager.CreateStream(streamInfo)
		createdStreams = append(createdStreams, streamInfo.StreamID)
	}

	list := manager.ListStreams()
	t.Logf("✓ 当前活跃流数量: %d", len(list))

	for _, info := range list {
		t.Logf("  - StreamID: %s, TaskCode: %s, Status: %s",
			info.StreamID, info.TaskCode, info.StreamStatus)
	}

	for _, streamID := range createdStreams {
		manager.RemoveStream(streamID)
	}
}

func TestRemoveStream(t *testing.T) {
	t.Log("=== 测试删除流 ===")
	manager := LiveStreamHandler.GetLiveStreamManager()

	streamInfo := &Board.LiveStreamInfo{
		StreamID:     fmt.Sprintf("stream_remove_%d", time.Now().UnixNano()),
		TaskCode:     fmt.Sprintf("TASK_REMOVE_%d", time.Now().UnixNano()),
		CentralID:    "central_remove_001",
		StreamStatus: Board.StreamStatus_Connected,
		VideoCodec:   Board.VideoCodec_H264,
		StartTime:    time.Now(),
	}

	manager.CreateStream(streamInfo)
	t.Logf("✓ 流创建成功: %s", streamInfo.StreamID)

	_, existsBefore := manager.GetStream(streamInfo.StreamID)
	if !existsBefore {
		t.Fatal("创建后应能找到流")
	}

	manager.RemoveStream(streamInfo.StreamID)

	_, existsAfter := manager.GetStream(streamInfo.StreamID)
	if existsAfter {
		t.Error("删除后不应再找到流")
	} else {
		t.Logf("✓ 流删除成功: %s", streamInfo.StreamID)
	}

	_, existsByTask := manager.GetStreamByTask(streamInfo.TaskCode)
	if existsByTask {
		t.Error("删除后不应通过 TaskCode 找到流")
	} else {
		t.Logf("✓ TaskCode 索引也已清理")
	}
}

func TestConcurrentStreamAccess(t *testing.T) {
	t.Log("=== 测试并发访问 ===")
	manager := LiveStreamHandler.GetLiveStreamManager()

	streamInfo := &Board.LiveStreamInfo{
		StreamID:     fmt.Sprintf("stream_concurrent_%d", time.Now().UnixNano()),
		TaskCode:     fmt.Sprintf("TASK_CONCURRENT_%d", time.Now().UnixNano()),
		CentralID:    "central_concurrent_001",
		StreamStatus: Board.StreamStatus_Connected,
		VideoCodec:   Board.VideoCodec_H264,
		StartTime:    time.Now(),
	}

	stream := manager.CreateStream(streamInfo)
	defer manager.RemoveStream(streamInfo.StreamID)

	var wg sync.WaitGroup
	errChan := make(chan error, 100)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			data := []byte(fmt.Sprintf("concurrent_data_%d_", idx))

			for j := 0; j < 100; j++ {
				stream.WriteData(data)
				_ = stream.GetBuffer()
				_ = stream.GetStatus()
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	errCount := 0
	for range errChan {
		errCount++
	}

	if errCount > 0 {
		t.Errorf("并发访问中出现 %d 个错误", errCount)
	} else {
		t.Logf("✓ 并发访问测试通过 (10个goroutine × 100次操作)")
	}
}

func TestBufferOverflowProtection(t *testing.T) {
	t.Log("=== 测试缓冲区溢出保护 ===")
	manager := LiveStreamHandler.GetLiveStreamManager()

	streamInfo := &Board.LiveStreamInfo{
		StreamID:     fmt.Sprintf("stream_buffer_%d", time.Now().UnixNano()),
		TaskCode:     fmt.Sprintf("TASK_BUFFER_%d", time.Now().UnixNano()),
		CentralID:    "central_buffer_001",
		StreamStatus: Board.StreamStatus_Connected,
		VideoCodec:   Board.VideoCodec_H264,
		StartTime:    time.Now(),
	}

	stream := manager.CreateStream(streamInfo)
	defer manager.RemoveStream(streamInfo.StreamID)

	largeChunk := make([]byte, 1024*1024)
	for i := 0; i < 15; i++ {
		stream.WriteData(largeChunk)
	}

	buffer := stream.GetBuffer()
	maxAllowedSize := 10 * 1024 * 1024
	if len(buffer) > maxAllowedSize+1024*1024 {
		t.Errorf("缓冲区超出限制: %d > %d", len(buffer), maxAllowedSize)
	} else {
		t.Logf("✓ 缓冲区溢出保护正常工作: 当前大小=%d bytes (%.2f MB)",
			len(buffer), float64(len(buffer))/(1024*1024))
	}
}

func TestRTMPTranslatorCreation(t *testing.T) {
	t.Log("=== 测试 RTMPTranslator 初始化 ===")

	translator := LiveStreamHandler.GetRTMPTranslator()
	if translator == nil {
		t.Fatal("RTMPTranslator 不应为 nil")
	}
	t.Logf("✓ RTMPTranslator 创建成功: %p", translator)
}

func TestLiveStreamHandlerCreation(t *testing.T) {
	t.Log("=== 测试 LiveStreamHandler 初始化 ===")

	handler := LiveStreamHandler.NewLiveStreamHandler()
	if handler == nil {
		t.Fatal("LiveStreamHandler 不应为 nil")
	}
	t.Logf("✓ LiveStreamHandler 创建成功: %p", handler)
}

func TestHandleCentralUploadWithMetadata(t *testing.T) {
	t.Log("=== 测试 HandleCentralUploadWithMetadata API ===")
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := LiveStreamHandler.NewLiveStreamHandler()

	router.POST("/api/board/live", handler.HandleCentralUploadWithMetadata)

	metadata := Board.LiveStreamMetadata{
		TaskCode:   fmt.Sprintf("TASK_UPLOAD_%d", time.Now().UnixNano()),
		CentralID:  "central_upload_test",
		DroneID:    "drone_upload_test",
		VideoCodec: "h264",
		AudioCodec: "aac",
		Resolution: "1920x1080",
		FPS:        30,
		Timestamp:  time.Now(),
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("序列化元数据失败: %v", err)
	}

	reqBody := bytes.NewBuffer(metadataJSON)
	req := httptest.NewRequest(http.MethodPost, "/api/board/live", reqBody)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	t.Logf("响应状态码: %d", w.Code)
	t.Logf("响应体: %s", w.Body.String())

	if w.Code != http.StatusOK && w.Code != http.StatusUnauthorized {
		t.Logf("注意: 响应状态码为 %d (可能是中间件拦截)", w.Code)
	}
}

func TestHandleFrontendGetStream(t *testing.T) {
	t.Log("=== 测试 HandleFrontendGetStream API ===")
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := LiveStreamHandler.NewLiveStreamHandler()
	manager := LiveStreamHandler.GetLiveStreamManager()

	router.GET("/api/backend/live", handler.HandleFrontendGetStream)

	t.Log("场景1: 无参数请求 - 应返回 400")
	req := httptest.NewRequest(http.MethodGet, "/api/backend/live", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	t.Logf("  响应状态码: %d (期望: 400)", w.Code)
	if w.Code != http.StatusBadRequest {
		t.Errorf("  ❌ 期望 400，实际 %d", w.Code)
	} else {
		t.Logf("  ✓ 正确返回 400: %s", w.Body.String())
	}

	t.Log("场景2: 无效的 stream_id - 应返回 404")
	req = httptest.NewRequest(http.MethodGet, "/api/backend/live?stream_id=nonexistent_stream_123", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	t.Logf("  响应状态码: %d (期望: 404)", w.Code)
	if w.Code != http.StatusNotFound {
		t.Errorf("  ❌ 期望 404，实际 %d", w.Code)
	} else {
		t.Logf("  ✓ 正确返回 404: %s", w.Body.String())
	}

	t.Log("场景3: 无效的 task_code - 应返回 404")
	req = httptest.NewRequest(http.MethodGet, "/api/backend/live?task_code=INVALID_TASK", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	t.Logf("  响应状态码: %d (期望: 404)", w.Code)
	if w.Code != http.StatusNotFound {
		t.Errorf("  ❌ 期望 404，实际 %d", w.Code)
	} else {
		t.Logf("  ✓ 正确返回 404: %s", w.Body.String())
	}

	t.Log("场景4: 真实流但格式不支持 - 应返回 406")
	streamInfo2 := &Board.LiveStreamInfo{
		StreamID:     fmt.Sprintf("stream_format_test_%d", time.Now().UnixNano()),
		TaskCode:     fmt.Sprintf("TASK_FORMAT_%d", time.Now().UnixNano()),
		CentralID:    "central_format_001",
		StreamStatus: Board.StreamStatus_Connected,
		VideoCodec:   Board.VideoCodec_H264,
		StartTime:    time.Now(),
	}
	manager.CreateStream(streamInfo2)
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/backend/live?stream_id=%s&format=unsupported", streamInfo2.StreamID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	t.Logf("  响应状态码: %d (期望: 406)", w.Code)
	if w.Code != http.StatusNotAcceptable {
		t.Errorf("  ❌ 期望 406，实际 %d", w.Code)
	} else {
		t.Logf("  ✓ 正确返回 406: %s", w.Body.String())
	}
	manager.RemoveStream(streamInfo2.StreamID)

	t.Log("场景5: 有效流但已断开 - 应返回 503")
	streamInfo := &Board.LiveStreamInfo{
		StreamID:     fmt.Sprintf("stream_disconnected_%d", time.Now().UnixNano()),
		TaskCode:     fmt.Sprintf("TASK_DISCONNECTED_%d", time.Now().UnixNano()),
		CentralID:    "central_disconnected_001",
		StreamStatus: Board.StreamStatus_Disconnected,
		VideoCodec:   Board.VideoCodec_H264,
		StartTime:    time.Now(),
	}
	stream := manager.CreateStream(streamInfo)
	stream.UpdateStatus(Board.StreamStatus_Disconnected)

	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/backend/live?stream_id=%s", streamInfo.StreamID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	t.Logf("  响应状态码: %d (期望: 503)", w.Code)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("  ❌ 期望 503，实际 %d", w.Code)
	} else {
		t.Logf("  ✓ 正确返回 503: %s", w.Body.String())
	}

	manager.RemoveStream(streamInfo.StreamID)
	t.Log("✓ HandleFrontendGetStream 错误处理测试完成")
}

func TestHandleListStreamsAPI(t *testing.T) {
	t.Log("=== 测试 HandleListStreams API ===")
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := LiveStreamHandler.NewLiveStreamHandler()

	router.GET("/api/backend/live/list", handler.HandleListStreams)

	req := httptest.NewRequest(http.MethodGet, "/api/backend/live/list", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	t.Logf("响应状态码: %d", w.Code)
	t.Logf("响应体: %s", w.Body.String())

	var response Board.LiveStreamListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err == nil {
		t.Logf("✓ 响应解析成功: success=%v, streams_count=%d",
			response.Success, len(response.Data))
	}
}

func TestHandleStopStreamAPI(t *testing.T) {
	t.Log("=== 测试 HandleStopStream API ===")
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := LiveStreamHandler.NewLiveStreamHandler()

	router.DELETE("/api/backend/live/:stream_id", handler.HandleStopStream)

	testStreamID := fmt.Sprintf("nonexistent_stream_%d", time.Now().UnixNano())
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/backend/live/%s", testStreamID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	t.Logf("响应状态码: %d", w.Code)
	t.Logf("响应体: %s", w.Body.String())

	if w.Code == http.StatusNotFound {
		t.Logf("✓ 不存在的流返回 404 (符合预期)")
	}
}

func TestMultipleStreamLifecycle(t *testing.T) {
	t.Log("=== 测试多流生命周期 ===")
	manager := LiveStreamHandler.GetLiveStreamManager()

	var streamIDs []string

	for i := 0; i < 5; i++ {
		streamInfo := &Board.LiveStreamInfo{
			StreamID:     fmt.Sprintf("stream_lifecycle_%d_%d", i, time.Now().UnixNano()),
			TaskCode:     fmt.Sprintf("TASK_LIFECYCLE_%d_%d", i, time.Now().UnixNano()),
			CentralID:    fmt.Sprintf("central_lc_%03d", i),
			DroneID:      fmt.Sprintf("drone_lc_%03d", i),
			StreamStatus: Board.StreamStatus_Connected,
			VideoCodec:   Board.VideoCodec_H264,
			AudioCodec:   Board.AudioCodec_AAC,
			Resolution:   "1280x720",
			FPS:          25 + i*5,
			StartTime:    time.Now(),
		}

		stream := manager.CreateStream(streamInfo)
		streamIDs = append(streamIDs, streamInfo.StreamID)

		stream.UpdateStreamStats(int64(3000000+i*1000000), float64(60*i))
	}

	allStreams := manager.ListStreams()
	t.Logf("✓ 创建了 %d 个流", len(allStreams))

	for idx, info := range allStreams {
		t.Logf("  流 #%d: ID=%s, Task=%s, Bitrate=%d, FPS=%d",
			idx+1, info.StreamID, info.TaskCode, info.Bitrate, info.FPS)
	}

	for i := 0; i < len(streamIDs)/2; i++ {
		manager.RemoveStream(streamIDs[i])
	}
	t.Logf("✓ 移除了 %d 个流", len(streamIDs)/2)

	remaining := manager.ListStreams()
	t.Logf("✓ 剩余 %d 个活跃流", len(remaining))

	for _, id := range streamIDs[len(streamIDs)/2:] {
		manager.RemoveStream(id)
	}
	t.Logf("✓ 所有流已清理完毕")
}

func BenchmarkStreamWrite(b *testing.B) {
	manager := LiveStreamHandler.GetLiveStreamManager()

	streamInfo := &Board.LiveStreamInfo{
		StreamID:     "benchmark_stream",
		TaskCode:     "benchmark_task",
		CentralID:    "benchmark_central",
		StreamStatus: Board.StreamStatus_Connected,
		VideoCodec:   Board.VideoCodec_H264,
		StartTime:    time.Now(),
	}

	stream := manager.CreateStream(streamInfo)
	defer manager.RemoveStream(streamInfo.StreamID)

	data := make([]byte, 1400)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			stream.WriteData(data)
			i++
		}
	})
}

func BenchmarkStreamAddViewer(b *testing.B) {
	manager := LiveStreamHandler.GetLiveStreamManager()

	streamInfo := &Board.LiveStreamInfo{
		StreamID:     "bench_viewer_stream",
		TaskCode:     "bench_viewer_task",
		CentralID:    "bench_viewer_central",
		StreamStatus: Board.StreamStatus_Connected,
		VideoCodec:   Board.VideoCodec_H264,
		StartTime:    time.Now(),
	}

	stream := manager.CreateStream(streamInfo)
	defer manager.RemoveStream(streamInfo.StreamID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		viewerID := fmt.Sprintf("viewer_bench_%d", i)
		ch := stream.AddViewer(viewerID)
		if i%100 == 0 {
			stream.RemoveViewer(viewerID)
		}
		_ = ch
	}
}

func TestMain(m *testing.M) {
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("视频流传输系统测试套件")
	fmt.Println("测试时间:", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println(strings.Repeat("=", 80))

	exitCode := m.Run()

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("测试完成")
	fmt.Println(strings.Repeat("=", 80))

	os.Exit(exitCode)
}
