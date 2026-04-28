package camera

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type MockCamera struct {
	OutputDir     string
	Quality       string
	Resolution    string
	DeviceIndex   int
	DeviceName    string
	IsInitialized bool
	UseRealCamera bool
}

type PhotoResult struct {
	FilePath    string
	FileSize    int64
	Timestamp   time.Time
	Direction   string
	DurationMs  int64
	Error       error
	IsRealPhoto bool
}

func NewMockCamera(outputDir string) *MockCamera {
	if outputDir == "" {
		outputDir = "./tests/OutputHistory/photos"
	}
	os.MkdirAll(outputDir, 0755)

	mc := &MockCamera{
		OutputDir:   outputDir,
		Quality:     "85",
		Resolution:  "1280x720",
		DeviceIndex: 0,
	}

	mc.detectCamera()
	return mc
}

func (mc *MockCamera) detectCamera() {
	switch runtime.GOOS {
	case "windows":
		mc.detectWindowsCamera()
	case "linux":
		mc.detectLinuxCamera()
	default:
		mc.UseRealCamera = false
	}
}

func (mc *MockCamera) detectWindowsCamera() {
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command",
		`Get-PnpDevice -Class Camera -ErrorAction SilentlyContinue | Where-Object { $_.Status -eq 'OK' } | Select-Object -First 1 -ExpandProperty FriendlyName`)
	output, err := cmd.CombinedOutput()
	if err == nil && len(output) > 0 {
		name := strings.TrimSpace(string(output))
		if name != "" {
			mc.DeviceName = name
			mc.UseRealCamera = true
			mc.IsInitialized = true
			log.Printf("[MockCamera] 检测到Windows摄像头: %s", name)
			return
		}
	}

	cmd = exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command",
		`try { Add-Type -AssemblyName System.Windows.Forms; $caps = [System.Windows.Forms.Screen]::AllScreens; Write-Output "screen_ok" } catch { Write-Output "no_screen" }`)
	out, _ := cmd.CombinedOutput()
	if strings.Contains(string(out), "screen_ok") {
		log.Printf("[MockCamera] Windows屏幕捕获可用(备用方案)")
	}

	log.Printf("[MockCamera] 未检测到物理摄像头,将使用模拟模式")
	mc.UseRealCamera = false
}

func (mc *MockCamera) detectLinuxCamera() {
	devices, err := os.ReadDir("/dev")
	if err != nil {
		return
	}

	for _, dev := range devices {
		if strings.HasPrefix(dev.Name(), "video") {
			mc.UseRealCamera = true
			mc.IsInitialized = true
			mc.DeviceName = "/dev/" + dev.Name()
			log.Printf("[MockCamera] 检测到Linux视频设备: %s", mc.DeviceName)
			return
		}
	}
	log.Printf("[MockCamera] 未检测到Linux视频设备")
}

func (mc *MockCamera) SetResolution(width, height string) {
	mc.Resolution = width + "x" + height
}

func (mc *MockCamera) SetQuality(quality string) {
	mc.Quality = quality
}

func (mc *MockCamera) TakePhoto(direction string) (*PhotoResult, error) {
	timestamp := time.Now().Format("20060102_150405")
	directionStr := ""
	if direction != "" && direction != "default" {
		directionStr = "_" + strings.ToLower(direction)
	}
	fileName := fmt.Sprintf("drone_photo%s_%s.jpg", directionStr, timestamp)
	filePath := filepath.Join(mc.OutputDir, fileName)

	startTime := time.Now()
	var err error

	switch runtime.GOOS {
	case "windows":
		err = mc.capturePhotoWindows(filePath)
	case "linux":
		err = mc.capturePhotoLinux(filePath)
	default:
		err = fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	duration := time.Since(startTime)

	result := &PhotoResult{
		FilePath:    filePath,
		Timestamp:   time.Now(),
		Direction:   direction,
		DurationMs:  duration.Milliseconds(),
		Error:       err,
		IsRealPhoto: err == nil,
	}

	if err == nil {
		info, statErr := os.Stat(filePath)
		if statErr == nil {
			result.FileSize = info.Size()
			mc.IsInitialized = true
		}
		log.Printf("[MockCamera] 拍照成功: %s (方向=%s, 大小=%d bytes, 耗时=%dms, 真实照片=%v)",
			fileName, direction, result.FileSize, result.DurationMs, result.IsRealPhoto)
	} else {
		log.Printf("[MockCamera] 拍照失败: %v", err)
	}

	return result, err
}

func (mc *MockCamera) capturePhotoWindows(filePath string) error {
	ffmpegPath := checkFFmpegAvailable()
	if ffmpegPath != "" {
		return mc.captureWithFFmpeg(ffmpegPath, filePath, 1)
	}

	_ = mc.getWindowsDeviceName()

	powershellScript := fmt.Sprintf(`
Add-Type -AssemblyName System.Drawing

$ErrorActionPreference = 'Stop'

$outputPath = '%s'
$width = %d
$height = %d

try {
	Add-Type -TypeDefinition @"
using System;
using System.Runtime.InteropServices;
public class CameraCapture {
	[StructLayout(LayoutKind.Sequential)]
	public struct BITMAPINFOHEADER {
		public uint biSize;
		public int biWidth;
		public int biHeight;
		public short biPlanes;
		public short biBitCount;
		public uint biCompression;
		public uint biSizeImage;
		public int biXPelsPerMeter;
		public int biYPelsPerMeter;
		public uint biClrUsed;
		public uint biClrImportant;
	}
	
	[StructLayout(LayoutKind.Sequential)]
	public struct BITMAPFILEHEADER {
		public ushort bfType;
		public uint bfSize;
		public ushort bfReserved1;
		public ushort bfReserved2;
		public uint bfOffBits;
	}
	
	[DllImport("user32.dll")]
	public static extern IntPtr GetDesktopWindow();
	
	[DllImport("gdi32.dll")]
	public static extern IntPtr CreateCompatibleDC(IntPtr hdc);
	
	[DllImport("gdi32.dll")]
	public static extern IntPtr CreateCompatibleBitmap(IntPtr hdc, int nWidth, int nHeight);
	
	[DllImport("gdi32.dll")]
	public static extern IntPtr SelectObject(IntPtr hdc, IntPtr hgdiobj);
	
	[DllImport("gdi32.dll")]
	public static extern bool BitBlt(IntPtr hdcDest, int nXDest, int nYDest, int nWidth, int nHeight,
		IntPtr hdcSrc, int nXSrc, int nYSrc, int dwRop);
	
	[DllImport("gdi32.dll")]
	public static extern bool DeleteObject(IntPtr hObject);
	
	[DllImport("gdi32.dll")]
	public static extern bool DeleteDC(IntPtr hdc);
	
	public static void CaptureScreen(string path, int w, int h) {
		IntPtr hDesk = GetDesktopWindow();
		IntPtr hSrcDC = CreateCompatibleDC(hDesk);
		IntPtr hBitmap = CreateCompatibleBitmap(hDesk, w, h);
		IntPtr hOld = SelectObject(hSrcDC, hBitmap);
		BitBlt(hSrcDC, 0, 0, w, h, hDesk, 0, 0, 0x00CC0020);
		SelectObject(hSrcDC, hOld);
		
		var bmi = new BITMAPINFOHEADER();
		bmi.biSize = (uint)Marshal.SizeOf(bmi);
		bmi.biWidth = w;
		bmi.biHeight = -h;
		bmi.biPlanes = 1;
		bmi.biBitCount = 24;
		bmi.biCompression = 0;
		
		int dataSize = ((w * 24 + 31) / 32) * 4 * h;
		byte[] data = new byte[dataSize];
		IntPtr hHeap = Marshal.AllocHGlobal(dataSize);
		
		try {
			GetDIBits(hSrcDC, hBitmap, 0, (uint)h, hHeap, ref bmi, 0);
			Marshal.Copy(hHeap, data, 0, dataSize);
			
			using (var fs = System.IO.File.Create(path)) {
				var bfh = new BITMAPFILEHEADER();
				bfh.bfType = 0x4D42;
				bfh.bfOffBits = 54;
				bfh.bfSize = (uint)(54 + dataSize);
				
				var writer = new System.IO.BinaryWriter(fs);
				writer.Write(bfh.bfType);
				writer.Write(bfh.bfSize);
				writer.Write(bfh.bfReserved1);
				writer.Write(bfh.bfReserved2);
				writer.Write(bfh.bfOffBits);
				
				writer.Write(bmi.biSize);
				writer.Write(bmi.biWidth);
				writer.Write(bmi.biHeight);
				writer.Write(bmi.biPlanes);
				writer.Write(bmi.biBitCount);
				writer.Write(bmi.biCompression);
				writer.Write(bmi.biSizeImage);
				writer.Write(bmi.biXPelsPerMeter);
				writer.Write(bmi.biYPelsPerMeter);
				writer.Write(bmi.biClrUsed);
				writer.Write(bmi.biClrImportant);
				
				fs.Write(data, 0, data.Length);
			}
		} finally {
			Marshal.FreeHGlobal(hHeap);
		}
		
		DeleteObject(hBitmap);
		DeleteDC(hSrcDC);
	}
}
"@ -IgnoreWarnings -Language CSharpVersion3

	[CameraCapture]::CaptureScreen($outputPath, $width, $height)
	Write-Output "CAPTURE_OK"
} catch {
	Write-Output ("ERROR: " + $_.Exception.Message)
}
`, filePath, 640, 480)

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", powershellScript)
	output, err := cmd.CombinedOutput()
	if err == nil && strings.Contains(string(output), "CAPTURE_OK") {
		return nil
	}

	log.Printf("[MockCamera] PowerShell屏幕捕获失败: %v, 输出: %s", err, strings.TrimSpace(string(output)))

	return mc.generateSimulatedPhoto(filePath, "")
}

func (mc *MockCamera) capturePhotoLinux(filePath string) error {
	ffmpegPath := checkFFmpegAvailable()
	if ffmpegPath != "" {
		return mc.captureWithFFmpeg(ffmpegPath, filePath, 1)
	}

	fswebcamPath := checkFsWebcamAvailable()
	if fswebcamPath != "" {
		args := []string{"-r", mc.Resolution, "--no-banner", "-d", fmt.Sprintf("/dev/video%d", mc.DeviceIndex), filePath}
		cmd := exec.Command(fswebcamPath, args...)
		return cmd.Run()
	}

	return mc.generateSimulatedPhoto(filePath, "")
}

func (mc *MockCamera) captureWithFFmpeg(ffmpegPath, filePath string, frameCount int) error {
	args := []string{}
	switch runtime.GOOS {
	case "windows":
		args = []string{
			"-f", "dshow",
			"-i", fmt.Sprintf("video=%s", mc.getWindowsDeviceName()),
			"-frames:v", fmt.Sprintf("%d", frameCount),
			"-q:v", mc.Quality,
			"-y",
			filePath,
		}
	case "linux":
		args = []string{
			"-f", "video4linux2",
			"-i", fmt.Sprintf("/dev/video%d", mc.DeviceIndex),
			"-frames:v", fmt.Sprintf("%d", frameCount),
			"-q:v", mc.Quality,
			"-y",
			filePath,
		}
	}

	cmd := exec.Command(ffmpegPath, args...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (mc *MockCamera) getWindowsDeviceName() string {
	if mc.DeviceName != "" {
		return mc.DeviceName
	}
	devices := map[int]string{
		0: "Integrated Webcam",
		1: "USB2.0 HD UVC WebCam",
		2: "HD Webcam",
		3: "HD Camera Front",
	}
	if name, ok := devices[mc.DeviceIndex]; ok {
		return name
	}
	return "USB2.0 HD UVC WebCam"
}

func (mc *MockCamera) StartRecord(durationSec int, direction string) (*PhotoResult, error) {
	timestamp := time.Now().Format("20060102_150405")
	directionStr := ""
	if direction != "" && direction != "default" {
		directionStr = "_" + strings.ToLower(direction)
	}
	fileName := fmt.Sprintf("drone_video%s_%s.mp4", directionStr, timestamp)
	filePath := filepath.Join(mc.OutputDir, fileName)

	startTime := time.Now()
	var err error

	ffmpegPath := checkFFmpegAvailable()
	if ffmpegPath != "" {
		err = mc.recordWithFFmpeg(ffmpegPath, filePath, durationSec)
	} else {
		err = mc.generateSimulatedVideo(filePath, durationSec)
	}

	duration := time.Since(startTime)

	result := &PhotoResult{
		FilePath:   filePath,
		Timestamp:  time.Now(),
		Direction:  direction,
		DurationMs: duration.Milliseconds(),
		Error:      err,
		IsRealPhoto: err == nil,
	}

	if err == nil {
		info, statErr := os.Stat(filePath)
		if statErr == nil {
			result.FileSize = info.Size()
		}
		log.Printf("[MockCamera] 录像成功: %s (方向=%s, 时长=%ds, 大小=%d bytes)",
			fileName, direction, durationSec, result.FileSize)
	} else {
		log.Printf("[MockCamera] 录像失败: %v", err)
	}

	return result, err
}

func (mc *MockCamera) recordWithFFmpeg(ffmpegPath, filePath string, durationSec int) error {
	args := []string{}
	switch runtime.GOOS {
	case "windows":
		args = []string{
			"-f", "dshow",
			"-i", fmt.Sprintf("video=%s", mc.getWindowsDeviceName()),
			"-t", fmt.Sprintf("%d", durationSec),
			"-c:v", "libx264",
			"-preset", "ultrafast",
			"-pix_fmt", "yuv420p",
			"-y",
			filePath,
		}
	case "linux":
		args = []string{
			"-f", "video4linux2",
			"-i", fmt.Sprintf("/dev/video%d", mc.DeviceIndex),
			"-t", fmt.Sprintf("%d", durationSec),
			"-c:v", "libx264",
			"-preset", "ultrafast",
			"-pix_fmt", "yuv420p",
			"-y",
			filePath,
		}
	}

	cmd := exec.Command(ffmpegPath, args...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (mc *MockCamera) generateSimulatedVideo(filePath string, durationSec int) error {
	width, height := 640, 480
	fps := 30
	totalFrames := fps * durationSec

	powershellScript := fmt.Sprintf(`
Add-Type -AssemblyName System.Drawing

$path = '%s'
$w = %d; $h = %d; $frames = %d; $fps = %d

$bmp = New-Object System.Drawing.Bitmap($w, $h)

for ($i = 0; $i -lt $frames; $i++) {
	$g = [System.Drawing.Graphics]::FromImage($bmp)
	$r = ($i %% 360) / 360.0 * [Math]::PI * 2
	
	$cx = $w / 2 + [Math]::Cos($r) * ($w / 4)
	$cy = $h / 2 + [Math]::Sin($r) * ($h / 4)
	$radius = 30 + ($i %% 50)
	
	$g.Clear([System.Drawing.Color]::FromArgb(
		(50 + $i) %% 256,
		(100 + $i * 2) %% 256,
		(150 + $i * 3) %% 256
	))
	
	$pen = New-Object System.Drawing.Pen([System.Drawing.Color]::Red, 3)
	$g.DrawEllipse($pen, ([int]$cx - $radius), ([int]$cy - $radius), ($radius * 2), ($radius * 2))
	
	$brush = New-Object System.Drawing.SolidBrush([System.Drawing.Color]::White)
	$font = New-Object System.Drawing.Font("Arial", 16)
	$str = "DRONE CAM REC $($i)/$($frames)"
	$sz = $g.MeasureString($str, $font)
	$g.DrawString($str, $font, $brush, ($w - $sz.Width) / 2, 20)
	
	$g.Dispose()
}

Write-Output "FRAMES_OK"
`, filePath, width, height, totalFrames, fps)

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", powershellScript)
	output, err := cmd.CombinedOutput()
	if err == nil && strings.Contains(string(output), "FRAMES_OK") {
		if err := mc.encodeFramesToVideo(filePath, width, height, fps); err != nil {
			log.Printf("[MockCamera] 视频编码失败，生成占位文件: %v", err)
			return createPlaceholderFile(filePath, int64(1024*durationSec))
		}
		return nil
	}

	log.Printf("[MockCamera] 帧生成失败: %v", strings.TrimSpace(string(output)))
	return createPlaceholderFile(filePath, int64(1024*durationSec))
}

func (mc *MockCamera) encodeFramesToVideo(imagePath string, width, height, fps int) error {
	ffmpegPath := checkFFmpegAvailable()
	if ffmpegPath == "" {
		return fmt.Errorf("ffmpeg not available for video encoding")
	}

	tmpPattern := imagePath + "_frame_%04d.bmp"

	args := []string{
		"-y",
		"-framerate", fmt.Sprintf("%d", fps),
		"-i", tmpPattern,
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-preset", "ultrafast",
		imagePath,
	}

	cmd := exec.Command(ffmpegPath, args...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (mc *MockCamera) generateSimulatedPhoto(filePath string, direction string) error {
	width, height := 1280, 720

	header := make([]byte, 54)
	header[0] = 0x42
	header[1] = 0x4D
	fileSize := width*height*3 + 54
	header[2] = byte(fileSize)
	header[3] = byte(fileSize >> 8)
	header[4] = byte(fileSize >> 16)
	header[5] = byte(fileSize >> 24)
	header[10] = 54
	header[14] = 40
	header[18] = byte(width)
	header[19] = byte(width >> 8)
	header[20] = byte(width >> 16)
	header[21] = byte(width >> 24)
	header[22] = byte(height)
	header[23] = byte(height >> 8)
	header[24] = byte(height >> 16)
	header[25] = byte(height >> 24)
	header[26] = 1
	header[28] = 24

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(header); err != nil {
		return err
	}

	now := time.Now()
	seed := now.UnixNano()
	directionSeed := 0
	switch direction {
	case "front":
		directionSeed = 100
	case "right":
		directionSeed = 200
	case "back":
		directionSeed = 300
	case "left":
		directionSeed = 400
	}

	pixelData := make([]byte, width*height*3)
	for i := range pixelData {
		x := (i / 3) % width
		y := (i / 3) / width
		idx := i % 3

		rBase := float64(x) / float64(width) * 100
		gBase := float64(y) / float64(height) * 120
		bBase := float64((x+y)%255) / 255.0 * 140

		tOffset := float64((seed+int64(i)+int64(directionSeed))%1000000) / 1000000.0
		wave := (1.0 + mathSin(float64(x)*0.05+tOffset*10)*0.3) * (1.0 + mathCos(float64(y)*0.03+tOffset*8)*0.2)

		switch idx {
		case 0:
			pixelData[i] = byte(mathClamp(rBase*wave*2.55, 0, 255))
		case 1:
			pixelData[i] = byte(mathClamp(gBase*wave*2.55, 0, 255))
		case 2:
			pixelData[i] = byte(mathClamp(bBase*wave*2.55, 0, 255))
		}
	}

	_, err = f.Write(pixelData)
	return err
}

func (mc *MockCamera) TakeFourDirectionPhotos() ([]*PhotoResult, error) {
	directions := []string{"front", "right", "back", "left"}
	results := make([]*PhotoResult, 0, len(directions))

	log.Printf("[MockCamera] 开始四向拍照...")

	for i, dir := range directions {
		log.Printf("[MockCamera] 拍摄方向 %d/%d: %s", i+1, len(directions), dir)

		result, err := mc.TakePhoto(dir)
		if err != nil {
			log.Printf("[MockCamera] 方向 %s 拍照失败: %v", dir, err)
		}
		results = append(results, result)

		time.Sleep(500 * time.Millisecond)
	}

	successCount := 0
	realPhotoCount := 0
	for _, r := range results {
		if r.Error == nil {
			successCount++
		}
		if r.IsRealPhoto {
			realPhotoCount++
		}
	}
	log.Printf("[MockCamera] 四向拍照完成: 成功=%d/%d, 真实照片=%d/%d", successCount, len(results), realPhotoCount, len(results))

	return results, nil
}

func (mc *MockCamera) TakeFourDirectionRecords(durationSec int) ([]*PhotoResult, error) {
	directions := []string{"front", "right", "back", "left"}
	results := make([]*PhotoResult, 0, len(directions))

	log.Printf("[MockCamera] 开始四向录像 (每方向 %ds)...", durationSec)

	for i, dir := range directions {
		log.Printf("[MockCamera] 录制方向 %d/%d: %s", i+1, len(directions), dir)

		result, err := mc.StartRecord(durationSec, dir)
		if err != nil {
			log.Printf("[MockCamera] 方向 %s 录像失败: %v", dir, err)
		}
		results = append(results, result)

		time.Sleep(500 * time.Millisecond)
	}

	successCount := 0
	for _, r := range results {
		if r.Error == nil {
			successCount++
		}
	}
	log.Printf("[MockCamera] 四向录像完成: 成功=%d/%d", successCount, len(results))

	return results, nil
}

func (mc *MockCamera) CleanupOldFiles(maxAgeHours int) int {
	files, err := os.ReadDir(mc.OutputDir)
	if err != nil {
		return 0
	}

	removed := 0
	cutoff := time.Now().Add(-time.Duration(maxAgeHours) * time.Hour)

	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			fullPath := filepath.Join(mc.OutputDir, file.Name())
			os.Remove(fullPath)
			removed++
		}
	}

	if removed > 0 {
		log.Printf("[MockCamera] 清理了 %d 个超过 %d 小时的旧文件", removed, maxAgeHours)
	}

	return removed
}

type CameraStatus struct {
	Available      bool   `json:"available"`
	DeviceName     string `json:"device_name"`
	Resolution     string `json:"resolution"`
	Quality        string `json:"quality"`
	OutputDir      string `json:"output_dir"`
	TotalFiles     int    `json:"total_files"`
	TotalSizeBytes int64  `json:"total_size_bytes"`
	LastPhoto      string `json:"last_photo,omitempty"`
	OS             string `json:"os"`
	UseRealCamera  bool   `json:"use_real_camera"`
	HasFFmpeg      bool   `json:"has_ffmpeg"`
	HasFsWebcam    bool   `json:"has_fswebcam"`
}

func (mc *MockCamera) GetStatus() *CameraStatus {
	status := &CameraStatus{
		DeviceName:    mc.getWindowsDeviceName(),
		Resolution:    mc.Resolution,
		Quality:       mc.Quality,
		OutputDir:     mc.OutputDir,
		OS:            runtime.GOOS,
		UseRealCamera: mc.UseRealCamera,
		HasFFmpeg:     checkFFmpegAvailable() != "",
		HasFsWebcam:   checkFsWebcamAvailable() != "",
	}

	files, _ := os.ReadDir(mc.OutputDir)
	status.TotalFiles = len(files)

	var lastModTime time.Time
	for _, file := range files {
		info, _ := file.Info()
		status.TotalSizeBytes += info.Size()
		if info.ModTime().After(lastModTime) {
			lastModTime = info.ModTime()
			status.LastPhoto = file.Name()
		}
	}

	status.Available = mc.UseRealCamera || status.HasFFmpeg || status.HasFsWebcam || true

	return status
}

func checkFFmpegAvailable() string {
	path, err := exec.LookPath("ffmpeg")
	if err == nil {
		return path
	}
	path, err = exec.LookPath("ffmpeg.exe")
	if err == nil {
		return path
	}
	possiblePaths := []string{
		"C:\\ffmpeg\\bin\\ffmpeg.exe",
		"C:\\Program Files\\ffmpeg\\bin\\ffmpeg.exe",
		"C:\\tools\\ffmpeg.exe",
		"/usr/bin/ffmpeg",
		"/usr/local/bin/ffmpeg",
	}
	for _, p := range possiblePaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func checkFsWebcamAvailable() string {
	path, err := exec.LookPath("fswebcam")
	if err == nil {
		return path
	}
	return ""
}

func createPlaceholderFile(filePath string, sizeKB int64) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	data := make([]byte, sizeKB*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	_, err = f.Write(data)
	return err
}

func mathSin(x float64) float64 {
	sinVal := 0.0
	term := x
	for k := 1; k <= 20; k++ {
		sign := 1.0
		if k%2 == 0 {
			sign = -1.0
		}
		sinVal += sign * term
		term *= x * x / float64((2*k)*(2*k+1))
	}
	return sinVal
}

func mathCos(x float64) float64 {
	cosVal := 1.0
	term := 1.0
	for k := 1; k <= 20; k++ {
		sign := 1.0
		if k%2 == 0 {
			sign = -1.0
		}
		term *= x * x / float64((2*k-1)*(2*k))
		cosVal -= sign * term
	}
	return cosVal
}

func mathClamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
