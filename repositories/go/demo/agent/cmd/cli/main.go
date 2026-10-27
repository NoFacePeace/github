package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"

	"github.com/NoFacePeace/github/repositories/go/demo/agent/event"
	"github.com/NoFacePeace/github/repositories/go/demo/agent/service"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, flag.ErrHelp) {
			return
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, input io.Reader, output, errorOutput io.Writer) error {
	flags := flag.NewFlagSet("agent-cli", flag.ContinueOnError)
	flags.SetOutput(errorOutput)
	configPath := flags.String("config", "config.yaml", "Path to the YAML configuration")
	var req service.Request
	flags.StringVar(&req.Provider, "provider", "", "Provider override")
	flags.StringVar(&req.Protocol, "protocol", "", "Protocol override")
	flags.StringVar(&req.Model, "model", "", "Model override")
	if err := flags.Parse(args); err != nil {
		return err
	}
	svc, err := buildService(*configPath)
	if err != nil {
		return err
	}
	return interactive(ctx, svc, req, input, output, errorOutput)
}

func interactive(ctx context.Context, svc service.Service, req service.Request, input io.Reader, output, errorOutput io.Writer) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	type lineResult struct {
		text string
		err  error
	}
	lines := make(chan lineResult)
	// 读取与执行分开，等待输入时也能响应取消；CLI 退出时进程释放 stdin。
	go func() {
		defer close(lines)
		scanner := bufio.NewScanner(input)
		scanner.Buffer(make([]byte, 4096), 1024*1024)
		for scanner.Scan() {
			select {
			case lines <- lineResult{text: scanner.Text()}:
			case <-ctx.Done():
				return
			}
		}
		if err := scanner.Err(); err != nil {
			select {
			case lines <- lineResult{err: err}:
			case <-ctx.Done():
			}
		}
	}()
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, err := fmt.Fprint(output, "> "); err != nil {
			return err
		}
		var line lineResult
		select {
		case <-ctx.Done():
			return ctx.Err()
		case next, ok := <-lines:
			if !ok {
				return nil
			}
			line = next
		}
		if line.err != nil {
			return line.err
		}
		req.Input = strings.TrimSpace(line.text)
		if req.Input == "" {
			continue
		}
		if req.Input == "exit" || req.Input == "quit" {
			return nil
		}
		requestCtx, stop := context.WithCancel(ctx)
		var writeErr error
		err := svc.Run(requestCtx, req, func(e event.Event) {
			if writeErr != nil {
				return
			}
			switch e := e.(type) {
			case event.TextDeltaEvent:
				_, writeErr = fmt.Fprint(output, e.Text)
			case event.ToolCallEvent:
				_, writeErr = fmt.Fprintln(errorOutput, "Tool:", e.Call.Name)
			case event.CompletedEvent:
				if e.Reason == event.StopMaxSteps {
					_, writeErr = fmt.Fprintln(errorOutput, "Stopped: max_steps")
				}
			}
			if writeErr != nil {
				stop()
			}
		})
		stop()
		if writeErr != nil {
			return writeErr
		}
		if _, writeErr := fmt.Fprintln(output); writeErr != nil {
			return writeErr
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err != nil {
			if _, writeErr := fmt.Fprintln(errorOutput, err); writeErr != nil {
				return writeErr
			}
		}
	}
}
