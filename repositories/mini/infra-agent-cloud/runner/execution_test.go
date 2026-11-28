package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"infra-agent-cloud/executor"
)

func TestRunnerExecutionFlow(t *testing.T) {
	for _, tc := range []struct {
		name        string
		final       string
		exitCode    int
		reject      bool
		failFinal   bool
		failFailure bool
		want        []string
	}{
		{
			name:  "completed",
			final: `{"schemaVersion":"1","executionId":"execution","seq":3,"type":"completed","payload":{"result":{"exitCode":0,"text":"done"}}}`,
			want:  []string{"heartbeat", "claim", "renew", "secrets", "start", "ack", "started", "output", "destroy", "completed"},
		},
		{
			name:     "missing terminal",
			exitCode: 1,
			want:     []string{"heartbeat", "claim", "renew", "secrets", "start", "ack", "started", "output", "destroy", "failed"},
		},
		{
			name:      "terminal delivery failed",
			final:     `{"schemaVersion":"1","executionId":"execution","seq":3,"type":"completed","payload":{"result":{"exitCode":0,"text":"done"}}}`,
			failFinal: true,
			want:      []string{"heartbeat", "claim", "renew", "secrets", "start", "ack", "started", "output", "destroy", "completed"},
		},
		{
			name:   "lease rejected",
			reject: true,
			want:   []string{"heartbeat", "claim", "renew", "failed", "ack"},
		},
		{
			name:        "lease rejection event delivery failed",
			reject:      true,
			failFailure: true,
			want:        []string{"heartbeat", "claim", "renew", "failed", "ack"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			var calls []string
			record := func(call string) { calls = append(calls, call) }
			job := &Job{
				MessageID: "execution", TaskID: "task", SessionID: "session", RunnerID: "runner",
				LeaseID: "lease", LeaseExpiresAt: time.Now().Add(time.Minute),
				Engine: "codex", Prompt: "hello", ExecutorSessionID: "previous",
			}
			acknowledged := make(chan struct{})
			store := &executionTestStore{
				jobs: []*Job{job}, record: record, failFinal: tc.failFinal, failFailure: tc.failFailure,
				onAck: func(string) {
					close(acknowledged)
					if tc.reject {
						cancel()
					}
				},
				onTerminal: func() {
					if !tc.reject {
						cancel()
					}
				},
			}
			client := &executionTestClient{record: record, reject: tc.reject}
			process := &executionTestRuntime{
				record: record,
				start: func(spec *executor.Spec, secrets *executor.Secrets) (*ExecutionHandle, error) {
					if spec.ExecutionID != job.MessageID || spec.ResumeSessionID != job.ExecutorSessionID ||
						spec.Prompt != job.Prompt || secrets.ExecutionID != spec.ExecutionID {
						return nil, errors.New("incorrect spec or secrets")
					}
					return &ExecutionHandle{
						Stdout: io.MultiReader(waitForTestSignal{ctx, acknowledged}, strings.NewReader(`{"schemaVersion":"1","executionId":"execution","seq":1,"type":"started"}
{"schemaVersion":"1","executionId":"execution","seq":2,"type":"channel_output"}
`+tc.final)),
						Stderr: strings.NewReader("diagnostic"),
						Wait:   func(context.Context) (int, error) { return tc.exitCode, nil },
					}, nil
				},
			}

			node := NewRunner(Heartbeat{RunnerID: job.RunnerID, Capacity: 1}, client, store, process)
			if err := node.Run(ctx); err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(calls, tc.want) {
				t.Fatalf("calls = %v, want %v", calls, tc.want)
			}
		})
	}
}

type waitForTestSignal struct {
	ctx    context.Context
	signal <-chan struct{}
}

func (r waitForTestSignal) Read([]byte) (int, error) {
	select {
	case <-r.signal:
		return 0, io.EOF
	case <-r.ctx.Done():
		return 0, r.ctx.Err()
	}
}

func TestRunnerStartsNextJobBeforeFirstExecutionExits(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	release := make(chan struct{})
	first := &Job{
		MessageID: "first", TaskID: "task1", SessionID: "session1", RunnerID: "runner",
		LeaseID: "lease1", LeaseExpiresAt: time.Now().Add(time.Minute),
	}
	second := &Job{
		MessageID: "second", TaskID: "task2", SessionID: "session2", RunnerID: "runner",
		LeaseID: "lease2", LeaseExpiresAt: first.LeaseExpiresAt,
	}
	var mu sync.Mutex
	var calls []string
	var completed int
	record := func(call string) {
		mu.Lock()
		defer mu.Unlock()
		calls = append(calls, call)
		if call == "completed" {
			completed++
			if completed == 2 {
				cancel()
			}
		}
	}
	store := &executionTestStore{
		jobs: []*Job{first, first, second}, record: record,
		onAck: func(entryID string) {
			if entryID == "entry-"+second.MessageID {
				close(release)
			}
		},
	}
	process := &executionTestRuntime{
		record: record,
		start: func(spec *executor.Spec, _ *executor.Secrets) (*ExecutionHandle, error) {
			wait := func(ctx context.Context) (int, error) {
				select {
				case <-release:
					return 0, nil
				case <-ctx.Done():
					return -1, ctx.Err()
				}
			}
			final := fmt.Sprintf(`{"schemaVersion":"1","executionId":%q,"seq":1,"type":"completed","payload":{"result":{"text":"done"}}}`, spec.ExecutionID)
			return &ExecutionHandle{
				Stdout: io.MultiReader(waitForTestSignal{ctx, release}, strings.NewReader(final)),
				Stderr: strings.NewReader(""), Wait: wait, WaitForExit: wait,
			}, nil
		},
		stop: func(context.Context) error { return nil },
	}
	node := NewRunner(Heartbeat{RunnerID: "runner", Capacity: 2}, &executionTestClient{record: record}, store, process)
	if err := node.Run(ctx); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	defer mu.Unlock()
	if want := []string{"heartbeat", "claim", "renew", "secrets", "start", "ack", "claim", "ack", "claim", "renew", "secrets", "start", "ack"}; !slices.Equal(calls[:min(len(calls), len(want))], want) {
		t.Fatalf("calls = %v, want prefix %v", calls, want)
	}
	if completed != 2 {
		t.Fatalf("completed = %d, want 2", completed)
	}
}

func TestStopCommandWaitsForExecutionCleanup(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var calls []string
	record := func(call string) { calls = append(calls, call) }
	job := &Job{
		MessageID: "execution", TaskID: "task", SessionID: "session", RunnerID: "runner",
		LeaseID: "lease", LeaseExpiresAt: time.Now().Add(time.Minute),
	}
	output, writer := io.Pipe()
	exited := make(chan struct{})
	wait := func(ctx context.Context) (int, error) {
		select {
		case <-ctx.Done():
			return -1, ctx.Err()
		case <-exited:
			return 130, nil
		}
	}
	process := &executionTestRuntime{
		record: record,
		start: func(*executor.Spec, *executor.Secrets) (*ExecutionHandle, error) {
			return &ExecutionHandle{Stdout: output, Stderr: strings.NewReader(""), Wait: wait, WaitForExit: wait}, nil
		},
		stop: func(ctx context.Context) error {
			record("stop")
			_, err := io.WriteString(writer, `{"schemaVersion":"1","executionId":"execution","seq":1,"type":"stopped","payload":{"result":{}}}`+"\n")
			_ = writer.Close()
			close(exited)
			return err
		},
	}
	t.Cleanup(func() { _ = output.Close(); _ = writer.Close() })
	node := NewRunner(Heartbeat{RunnerID: job.RunnerID}, &executionTestClient{record: record}, &executionTestStore{record: record}, process)
	if err := node.handleJob(ctx, job); err != nil {
		t.Fatal(err)
	}
	err := node.handleCommand(ctx, &Command{
		Type: CommandTypeStop, RunnerID: job.RunnerID, TaskID: job.TaskID, SessionID: job.SessionID,
		ExecutionToken: job.LeaseID, ExpiresAt: job.LeaseExpiresAt,
	})
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"renew", "secrets", "start", "stop", "destroy", "stopped"}; !slices.Equal(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

type executionTestStore struct {
	Store
	jobs        []*Job
	record      func(string)
	failFinal   bool
	failFailure bool
	onAck       func(string)
	onTerminal  func()
}

func (s *executionTestStore) PublishHeartbeat(context.Context, Heartbeat) error {
	s.record("heartbeat")
	return nil
}

func (s *executionTestStore) ClaimJob(context.Context) (*Job, string, error) {
	if len(s.jobs) == 0 {
		return nil, "", nil
	}
	s.record("claim")
	job := s.jobs[0]
	s.jobs = s.jobs[1:]
	return job, "entry-" + job.MessageID, nil
}

func (s *executionTestStore) ClaimCommand(context.Context) (*Command, error) {
	return nil, nil
}

func (s *executionTestStore) PublishEvent(_ context.Context, _ *Job, event Event) error {
	s.record(string(event.Type))
	if s.onTerminal != nil && (event.Type == EventTypeCompleted || event.Type == EventTypeFailed) {
		s.onTerminal()
	}
	if event.Type == EventTypeCompleted &&
		(s.failFinal || event.Result == nil || event.Result.Text == nil || *event.Result.Text != "done") {
		return errors.New("event delivery failed")
	}
	if event.Type == EventTypeFailed && s.failFailure {
		return errors.New("event delivery failed")
	}
	return nil
}

func (s *executionTestStore) AckJob(_ context.Context, entryID string) error {
	s.record("ack")
	if s.onAck != nil {
		s.onAck(entryID)
	}
	return nil
}

type executionTestClient struct {
	record func(string)
	reject bool
}

func (c *executionTestClient) RenewExecution(context.Context, *Job) (time.Time, error) {
	c.record("renew")
	if c.reject {
		return time.Time{}, ErrExecutionLeaseRejected
	}
	return time.Now().Add(time.Minute), nil
}

func (c *executionTestClient) ClaimSecrets(_ context.Context, job *Job) (*executor.Secrets, error) {
	c.record("secrets")
	return &executor.Secrets{ExecutionID: job.MessageID}, nil
}

type executionTestRuntime struct {
	record func(string)
	start  func(*executor.Spec, *executor.Secrets) (*ExecutionHandle, error)
	stop   func(context.Context) error
}

func (r *executionTestRuntime) Stop(ctx context.Context, _ *ExecutionHandle) error {
	return r.stop(ctx)
}

func (r *executionTestRuntime) Start(_ context.Context, spec *executor.Spec, secrets *executor.Secrets) (*ExecutionHandle, error) {
	r.record("start")
	return r.start(spec, secrets)
}

func (r *executionTestRuntime) Destroy(context.Context, *ExecutionHandle) error {
	r.record("destroy")
	return nil
}
