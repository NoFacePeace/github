package executor

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"
)

const (
	maxInputBytes      = 4 * 1024 * 1024
	secretInputTimeout = 5 * time.Second
	commandBufferSize  = 16
)

// ControlInput 管理 steer 输入，首帧为敏感配置，后续帧为控制命令。
// 输入错误取消 Context 并写入 Errors；正常 EOF 只结束命令流，不取消执行。
type ControlInput struct {
	Secrets  *Secrets
	Commands <-chan SteerCommand
	Errors   <-chan error
	Context  context.Context

	cancel context.CancelCauseFunc
	done   <-chan struct{}
}

// Close 取消输入读取并等待后台读取退出，可重复调用。
func (c *ControlInput) Close() {
	c.cancel(context.Canceled)
	<-c.done
}

// ReadSecrets 读取普通模式输入到 EOF；接管 input，并在结束或取消时关闭。
// input 的 Close 必须能够解除阻塞中的 Read。
func ReadSecrets(ctx context.Context, input io.ReadCloser) (*Secrets, error) {
	readCtx, cancel := context.WithTimeout(ctx, secretInputTimeout)
	defer cancel()
	stopClose := context.AfterFunc(readCtx, func() { _ = input.Close() })
	defer stopClose()
	defer input.Close()

	data, err := io.ReadAll(io.LimitReader(input, maxInputBytes+1))
	if cause := context.Cause(readCtx); cause != nil {
		return nil, cause
	}
	if err != nil {
		return nil, errors.New("failed to read execution secrets")
	}
	if len(data) > maxInputBytes {
		return nil, errors.New("execution secrets exceed input limit")
	}
	return parseSecrets(data)
}

// OpenControlInput 接管 input，同步读取首帧后返回，并持续读取后续命令。
// 所有帧必须以换行结束；input 的 Close 必须能够解除阻塞中的 Read。
func OpenControlInput(ctx context.Context, input io.ReadCloser, executionID string) (*ControlInput, error) {
	inputCtx, cancel := context.WithCancelCause(ctx)
	stopClose := context.AfterFunc(inputCtx, func() { _ = input.Close() })
	timer := time.AfterFunc(secretInputTimeout, func() {
		cancel(errors.New("execution secrets input timed out"))
	})
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 4096), maxInputBytes+1)
	scanner.Split(splitControlFrame)
	var secrets *Secrets
	var err error
	if scanner.Scan() {
		secrets, err = parseSecrets(scanner.Bytes())
		if err == nil && secrets.ExecutionID != executionID {
			err = errors.New("execution secrets have the wrong execution ID")
		}
	} else if scanner.Err() != nil {
		err = errors.New("failed to read execution secrets frame")
	} else {
		err = errors.New("execution secrets input ended before the first frame")
	}
	timer.Stop()
	if cause := context.Cause(inputCtx); cause != nil {
		err = cause
	}
	if err != nil {
		cancel(err)
		stopClose()
		_ = input.Close()
		return nil, err
	}

	commands := make(chan SteerCommand, commandBufferSize)
	errs := make(chan error, 1)
	done := make(chan struct{})
	control := &ControlInput{
		Secrets: secrets, Commands: commands, Errors: errs, Context: inputCtx,
		cancel: cancel, done: done,
	}
	go func() {
		defer close(done)
		defer close(commands)
		defer close(errs)
		defer stopClose()
		defer input.Close()
		fail := func(err error) {
			if inputCtx.Err() == nil {
				errs <- err
				cancel(err)
			}
		}
		// 复用首帧的 Scanner，保留其已预读的后续数据。
		for scanner.Scan() {
			var command *SteerCommand
			if err := json.Unmarshal(scanner.Bytes(), &command); err != nil || command == nil {
				fail(errors.New("invalid steer command JSON"))
				return
			}
			if command.ExecutionID != executionID {
				fail(errors.New("steer command has the wrong execution ID"))
				return
			}
			if command.CommandID == "" || strings.TrimSpace(command.Prompt) == "" {
				fail(errors.New("steer command requires a command ID and prompt"))
				return
			}
			select {
			case <-inputCtx.Done():
				return
			case commands <- *command:
			}
		}
		if scanner.Err() != nil {
			fail(errors.New("failed to read steer command frame"))
		}
	}()
	return control, nil
}

func parseSecrets(data []byte) (*Secrets, error) {
	var secrets *Secrets
	if err := json.Unmarshal(data, &secrets); err != nil || secrets == nil {
		return nil, errors.New("invalid execution secrets JSON")
	}
	return secrets, nil
}

func splitControlFrame(data []byte, atEOF bool) (int, []byte, error) {
	if end := bytes.IndexByte(data, '\n'); end >= 0 {
		if end > maxInputBytes {
			return 0, nil, errors.New("control input frame exceeds input limit")
		}
		return end + 1, data[:end], nil
	}
	if len(data) > maxInputBytes {
		return 0, nil, errors.New("control input frame exceeds input limit")
	}
	if atEOF && len(data) > 0 {
		return 0, nil, errors.New("control input ended with an incomplete frame")
	}
	return 0, nil, nil
}
