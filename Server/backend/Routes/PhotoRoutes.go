package Routes

import (
    "fmt"
    "net/http"
    "os"
    "path/filepath"
    "time"

    "github.com/gin-gonic/gin"
)

// InitPhotoRoutes 初始化照片数据上传的路由
func InitPhotoRoutes(r *gin.Engine) {
    // 定义上传文件的物理保存目录
    uploadDir := "./output/photos"
    
    // 确保目录存在
    if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
        os.MkdirAll(uploadDir, 0755)
    }

    uploadGroup := r.Group("/api/upload")
    {
        uploadGroup.POST("/photo", func(c *gin.Context) {
            droneID := c.PostForm("drone_id")
            if droneID == "" {
                droneID = "unknown_drone"
            }
            
            // 接收上传的文件，"photo"是前端/Central发来的字段名
            file, err := c.FormFile("photo")
            if err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "未接收到照片文件: " + err.Error()})
                return
            }

            // 保存到后端本地目录 (带上时间戳防重名)
            timestamp := time.Now().Format("20060102_150405")
            filename := fmt.Sprintf("%s_%s_%s", droneID, timestamp, filepath.Base(file.Filename))
            savePath := filepath.Join(uploadDir, filename)
            
            if err := c.SaveUploadedFile(file, savePath); err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "保存文件失败"})
                return
            }

            // 返回静态文件的访问URL
            c.JSON(http.StatusOK, gin.H{
                "code":    0,
                "msg":     "照片上传成功", 
                "data":    gin.H{"url": fmt.Sprintf("/photos/%s", filename)},
            })
        })
    }
}