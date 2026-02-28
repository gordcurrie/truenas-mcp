package truenas

import (
	"context"
	"fmt"
	"time"
)

const jobPollInterval = 2 * time.Second

// JobProgress holds the current progress of a running TrueNAS job.
type JobProgress struct {
	Percent     float64 `json:"percent"`
	Description string  `json:"description"`
}

// Job represents an async job returned by TrueNAS SCALE long-running operations.
// The job ID is returned immediately by lifecycle operations; use PollJob to wait
// for completion.
type Job struct {
	ID       int         `json:"id"`
	Method   string      `json:"method"`
	State    string      `json:"state"` // WAITING, RUNNING, SUCCESS, FAILED, ABORTED
	Progress JobProgress `json:"progress"`
	Error    *string     `json:"error"`
}

// getJobs fetches the job with the given ID from /core/get_jobs.
func (c *Client) getJobs(ctx context.Context, id int) (*Job, error) {
	path := fmt.Sprintf("/core/get_jobs?id=%d", id)
	var jobs []Job
	if err := c.get(ctx, path, &jobs); err != nil {
		return nil, fmt.Errorf("fetching job %d: %w", id, err)
	}
	if len(jobs) == 0 {
		return nil, fmt.Errorf("job %d not found", id)
	}
	return &jobs[0], nil
}

// PollJob polls the TrueNAS job API until the job reaches a terminal state
// (SUCCESS, FAILED, or ABORTED) or the context is cancelled.
// It returns the completed Job or an error if the job failed.
func (c *Client) PollJob(ctx context.Context, id int) (*Job, error) {
	for {
		job, err := c.getJobs(ctx, id)
		if err != nil {
			return nil, err
		}
		switch job.State {
		case "SUCCESS":
			return job, nil
		case "FAILED", "ABORTED":
			msg := fmt.Sprintf("job %d %s", id, job.State)
			if job.Error != nil {
				msg += ": " + *job.Error
			}
			return nil, fmt.Errorf("%s", msg) //nolint:goerr113 // dynamic error is intentional: surfaces the remote job failure message
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("polling job %d: %w", id, ctx.Err())
		case <-time.After(jobPollInterval):
		}
	}
}
