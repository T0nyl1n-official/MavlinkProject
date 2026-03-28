package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	// 测试本地服务器
	fmt.Println("测试本地服务器...")
	
	// 等待服务器启动
	time.Sleep(3 * time.Second)
	
	// 测试根路径
	resp, err := http.Get("http://localhost:8080/")
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
		return
	}
	
	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应内容: %s\n", string(body))
}