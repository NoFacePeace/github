package executor

import (
	"context"
	"io"
	"maps"
)

// SteerSession 保存初始化后的输出流和后续命令入口。
type SteerSession struct {
	Stdout io.ReadCloser
	// Steer 发送一条命令，并在 Context 取消时结束等待。
	Steer func(context.Context, SteerCommand) (map[string]any, error)
}

// consumeSteerCommands 顺序处理追加指令，并为重发的命令复用首次回执。
func consumeSteerCommands(ctx context.Context, commands <-chan SteerCommand, steer func(context.Context, SteerCommand) (map[string]any, error), publish func(EventType, map[string]any) error) error {
	results := make(map[string]map[string]any)
	for {
		select {
		case <-ctx.Done():
			return nil
		case command, ok := <-commands:
			if !ok || ctx.Err() != nil {
				return nil
			}
			payload, exists := results[command.CommandID]
			if !exists {
				response, err := steer(ctx, command)
				payload = make(map[string]any, len(response)+1)
				maps.Copy(payload, response)
				payload["commandId"] = command.CommandID
				if err != nil {
					payload["error"] = err.Error()
				}
				results[command.CommandID] = payload
			}
			if err := publish(EventTypeSteerResult, payload); err != nil {
				return err
			}
		}
	}
}
