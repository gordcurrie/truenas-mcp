package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

// newSynctestClient connects a Client to an in-memory WebSocket test server.
// It must be called from inside a synctest bubble so that the server, the
// client's read loop and PollJob's timer all run on the bubble's fake clock.
func newSynctestClient(t *testing.T, handlers map[string]methodHandler) *Client {
	t.Helper()
	srv := httptest.NewTestServer(t, wsHandler(t, handlers))
	transport, ok := srv.Client().Transport.(*http.Transport)
	if !ok || transport.DialContext == nil {
		t.Fatalf("unexpected test server transport %T", srv.Client().Transport)
	}

	c, err := NewClient("http://truenas.test", "test-api-key", false) //nolint:gosec // G101: test-api-key is a fake placeholder, not a real credential
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	c.dialContext = transport.DialContext
	if err := c.Connect(t.Context()); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

// jobSequence returns a core.get_jobs handler that reports each state in
// turn, repeating the last one, and counts how many times it was polled.
func jobSequence(polls *atomic.Int32, jobErr *string, states ...string) methodHandler {
	return func(json.RawMessage) (any, *rpcError) {
		n := int(polls.Add(1)) - 1
		state := states[min(n, len(states)-1)]
		job := Job{ID: 42, Method: "app.create", State: state}
		if state == "FAILED" || state == "ABORTED" {
			job.Error = jobErr
		}
		return []Job{job}, nil
	}
}

func TestPollJob(t *testing.T) {
	t.Run("returns job once it succeeds", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			var polls atomic.Int32
			c := newSynctestClient(t, map[string]methodHandler{
				"core.get_jobs": jobSequence(&polls, nil, "WAITING", "RUNNING", "RUNNING", "SUCCESS"),
			})

			start := time.Now()
			job, err := c.PollJob(t.Context(), 42)
			if err != nil {
				t.Fatalf("PollJob: %v", err)
			}
			if job.ID != 42 || job.State != "SUCCESS" {
				t.Errorf("job = %+v, want id 42 in SUCCESS", job)
			}
			if got := polls.Load(); got != 4 {
				t.Errorf("polled %d times, want 4", got)
			}
			if elapsed, want := time.Since(start), 3*jobPollInterval; elapsed != want {
				t.Errorf("elapsed %v, want %v (one interval between polls)", elapsed, want)
			}
		})
	})

	t.Run("succeeds immediately without waiting", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			var polls atomic.Int32
			c := newSynctestClient(t, map[string]methodHandler{
				"core.get_jobs": jobSequence(&polls, nil, "SUCCESS"),
			})

			start := time.Now()
			if _, err := c.PollJob(t.Context(), 42); err != nil {
				t.Fatalf("PollJob: %v", err)
			}
			if elapsed := time.Since(start); elapsed != 0 {
				t.Errorf("elapsed %v, want 0", elapsed)
			}
		})
	})

	for _, state := range []string{"FAILED", "ABORTED"} {
		t.Run(strings.ToLower(state)+" job returns remote error", func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				var polls atomic.Int32
				msg := "image pull failed"
				c := newSynctestClient(t, map[string]methodHandler{
					"core.get_jobs": jobSequence(&polls, &msg, "RUNNING", state),
				})

				_, err := c.PollJob(t.Context(), 42)
				want := "job 42 " + state + ": image pull failed"
				if err == nil || err.Error() != want {
					t.Errorf("err = %v, want %q", err, want)
				}
			})
		})
	}

	t.Run("failed job without message", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			var polls atomic.Int32
			c := newSynctestClient(t, map[string]methodHandler{
				"core.get_jobs": jobSequence(&polls, nil, "FAILED"),
			})

			_, err := c.PollJob(t.Context(), 42)
			if err == nil || err.Error() != "job 42 FAILED" {
				t.Errorf("err = %v, want %q", err, "job 42 FAILED")
			}
		})
	})

	t.Run("context cancelled while waiting", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			var polls atomic.Int32
			c := newSynctestClient(t, map[string]methodHandler{
				"core.get_jobs": jobSequence(&polls, nil, "RUNNING"),
			})

			ctx, cancel := context.WithTimeout(t.Context(), jobPollInterval*5/2)
			defer cancel()

			_, err := c.PollJob(ctx, 42)
			if !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("err = %v, want context.DeadlineExceeded", err)
			}
			if !strings.Contains(err.Error(), "polling job 42") {
				t.Errorf("err = %q, want it to mention polling job 42", err)
			}
			// Polls at t=0, 1x and 2x the interval; the deadline lands before the third tick.
			if got := polls.Load(); got != 3 {
				t.Errorf("polled %d times, want 3", got)
			}
		})
	})

	t.Run("job not found", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			c := newSynctestClient(t, map[string]methodHandler{
				"core.get_jobs": func(json.RawMessage) (any, *rpcError) { return []Job{}, nil },
			})

			_, err := c.PollJob(t.Context(), 42)
			if err == nil || !strings.Contains(err.Error(), "job 42 not found") {
				t.Errorf("err = %v, want job 42 not found", err)
			}
		})
	})
}
