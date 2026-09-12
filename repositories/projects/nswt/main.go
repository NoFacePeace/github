package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	action := flag.String("action", "flow", "action to run: flow, ciphercode, or asyncrush")
	packetPath := flag.String("packet", "", "path to the packet JSON file")
	configPath := flag.String("config", "config.json", "path to the flow config JSON file")
	timeout := flag.Duration("timeout", 30*time.Second, "total flow timeout")
	flag.Parse()

	if *action != "flow" && *action != "ciphercode" && *action != "asyncrush" {
		exitf("unknown action %q: expected flow, ciphercode, or asyncrush", *action)
	}
	if *packetPath == "" {
		if *action == "asyncrush" {
			*packetPath = "asyncrush.json"
		} else {
			*packetPath = "ciphercode.json"
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	params, err := loadPacketRequest(*packetPath)
	if err != nil {
		exitf("load packet: %v", err)
	}

	switch *action {
	case "ciphercode":
		result, err := ciphercode(ctx, http.DefaultClient, params)
		if err != nil {
			exitf("request ciphercode: %v", err)
		}
		fmt.Println(formatCiphercodeResponse(result))
	case "asyncrush":
		result, err := asyncRush(ctx, http.DefaultClient, params)
		if err != nil {
			exitf("request async rush: %v", err)
		}
		fmt.Println(formatAsyncRushResponse(result))
	case "flow":
		options, err := loadRushFlowOptions(*configPath)
		if err != nil {
			exitf("load config: %v", err)
		}
		result, err := runRushFlow(ctx, http.DefaultClient, params, options)
		if err != nil {
			exitf("run rush flow: %v", err)
		}
		fmt.Println(formatCiphercodeResponse(result.Ciphercode))
		fmt.Println(formatAsyncRushResponse(result.AsyncRush))
	}
}

func formatCiphercodeResponse(result ciphercodeResponse) string {
	if result.Code != 0 {
		return fmt.Sprintf("获取口令失败: code=%d, msg=%s", result.Code, result.Msg)
	}
	if result.Data == nil {
		return fmt.Sprintf("获取口令成功: %s", result.Msg)
	}

	return fmt.Sprintf(
		"获取口令成功: %s\nciphercode: %s\nsafesalt: %s",
		result.Msg,
		result.Data.Ciphercode,
		result.Data.Safesalt,
	)
}

func formatAsyncRushResponse(result asyncRushResponse) string {
	if result.Code != 0 {
		return fmt.Sprintf("抢券失败: code=%d, msg=%s", result.Code, result.Msg)
	}
	return fmt.Sprintf("抢券成功: %s", result.Msg)
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
