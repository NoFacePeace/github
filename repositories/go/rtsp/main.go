package main

import (
	"fmt"
	"log"
	"time"

	"gocv.io/x/gocv"
)

func main() {
	// RTSP 流地址
	// rtspURL := "rtsp://admin:admin12345@61.76.112.92:554/ISAPI/Streaming/channels/101"
	// rtspURL := "rtsp://admin:admin@61.39.27.104/cam/realmonitor?channel=1&subtype=1"
	// 打开 RTSP 视频流
	rtspURL := "rtsp://admin:admin12345@118.200.158.63:554/ISAPI/Streaming/channels/101"

	webcam, err := gocv.OpenVideoCapture(rtspURL)
	if err != nil {
		log.Fatalf("Error opening video capture: %v", err)
	}
	defer webcam.Close()

	// 创建窗口
	window := gocv.NewWindow("RTSP Stream")
	defer window.Close()

	// 创建 Mat 用于存储帧数据
	img := gocv.NewMat()
	defer img.Close()

	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Println("Cannot read from RTSP stream")
			break
		}
		if img.Empty() {
			continue
		}

		// 显示帧
		window.IMShow(img)
		if window.WaitKey(1) == 27 { // 按 ESC 退出
			break
		}

		time.Sleep(10 * time.Millisecond) // 稍微延迟，避免 CPU 占用过高
	}
}
