package AI

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	Models "MavlinkProject/Models"
)

var (
	thermalUploadDir  = "./output/thermal_photos"
	generatedImageDir = "./output/yolo_Generated"
)

func init() {
	os.MkdirAll(thermalUploadDir, 0755)
	os.MkdirAll(generatedImageDir, 0755)
}

type DronePhotoResponse struct {
	Code              int                           `json:"code"`
	Message           string                        `json:"message"`
	Alert             *Models.AlertJSON             `json:"alert,omitempty"`
	RawResult         *Models.ThermalDetectResponse `json:"raw_result,omitempty"`
	PhotoPath         string                        `json:"photo_path,omitempty"`
	AnnotatedImageURL string                        `json:"annotated_image_url,omitempty"`
}

func HandleDronePhotoUpload(c *gin.Context) {
	droneID := c.PostForm("drone_id")
	if droneID == "" {
		droneID = "unknown_drone"
	}

	file, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, DronePhotoResponse{
			Code:    1,
			Message: "未接收到照片文件: " + err.Error(),
		})
		return
	}

	latStr := c.DefaultPostForm("latitude", "0")
	lonStr := c.DefaultPostForm("longitude", "0")
	lat, _ := strconv.ParseFloat(latStr, 64)
	lon, _ := strconv.ParseFloat(lonStr, 64)

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s_%s", droneID, timestamp, filepath.Base(file.Filename))
	savePath := filepath.Join(thermalUploadDir, filename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, DronePhotoResponse{
			Code:    2,
			Message: "保存照片文件失败: " + err.Error(),
		})
		return
	}

	log.Printf("[ThermalHandler] 收到无人机照片: drone=%s, file=%s, lat=%.6f, lon=%.6f",
		droneID, filename, lat, lon)

	service := GetAnalysisService()

	alert, rawResult, err := service.ProcessDronePhoto(droneID, savePath, lat, lon)
	if err != nil {
		log.Printf("[ThermalHandler] 热源检测失败: drone=%s, err=%v", droneID, err)
		c.JSON(http.StatusInternalServerError, DronePhotoResponse{
			Code:      3,
			Message:   "热源检测失败: " + err.Error(),
			PhotoPath: savePath,
		})
		return
	}

	alertHistory.Add(*alert)

	var annotatedURL string
	if rawResult != nil && rawResult.ImageAnnotated != "" {
		annotatedURL = fmt.Sprintf("/api/ai/drone/photo/generated/%s", rawResult.ImageAnnotated)
	}

	c.JSON(http.StatusOK, DronePhotoResponse{
		Code:              0,
		Message:           "照片接收并检测完成",
		Alert:             alert,
		RawResult:         rawResult,
		PhotoPath:         savePath,
		AnnotatedImageURL: annotatedURL,
	})
}

func HandleDownloadGeneratedImage(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" || strings.Contains(filename, "..") || strings.Contains(filename, "\\") {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid filename"})
		return
	}

	filePath := filepath.Join(generatedImageDir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"code": 2, "message": "file not found"})
		return
	}

	c.File(filePath)
}
