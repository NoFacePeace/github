package main

import (
	"fmt"
	"time"
)

func main() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	go func() {
		for t := range ticker.C {
			fmt.Println(time.Now().Format("2006-01-02 15:04:05"))
			fmt.Println("任务开始:", t)
			time.Sleep(5 * time.Second) // 模拟任务耗时
			fmt.Println("任务结束:", t)
		}
	}()

	time.Sleep(50 * time.Second) // 主程序运行 10 秒
}
