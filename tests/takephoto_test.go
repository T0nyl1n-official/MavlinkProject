package tests

import (
    "log"
    "testing"
)

func TestTakePhotoChain(t *testing.T) {
    // 构建一个只有 TakePhoto 任务的任务链
    chain := ProgressChain{
        ChainID: "test_takephoto_001",
        Tasks: []Task{
            {
                TaskID:  "task_takephoto",
                Command: "TakePhoto", // 树莓派收到此指令将调用本地 fswebcam 拍照并上传
                Data: map[string]interface{}{
                    "delay": 5.0, // 给予树莓派充足的调用摄像头和网络上传时间
                },
                Status: "pending",
            },
        },
        Status: "pending",
    }

    log.Printf("==== 开始发送拍照任务链: %s ====", chain.ChainID)
    // 利用你原本已有的 SendProgressChainToCentral 函数发送 TCP 报文到通过 FRP 转发的树莓派服务上
    if err := SendProgressChainToCentral(chain); err != nil {
        t.Fatalf("发送拍照任务失败: %v", err)
    }
    
    log.Printf("拍照任务已成功下发！")
    log.Printf("请转去查看本地后端的终端日志（确保你本地的 Backend Server 已经运行）。")
    log.Printf("若执行成功，后端日志应显示接收到上传的照片，且图片保存在 ./OutputLogs/photos/ 目录下。")
}