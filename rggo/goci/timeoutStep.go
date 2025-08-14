package main

import (
	"bytes"
	"context"
	"os/exec"
	"time"
)

type timeoutStep struct {
	step
	timeout time.Duration
}

func newTimeoutStep(name, exe, proj, message string, args []string, timeout time.Duration) timeoutStep {
	s := timeoutStep{
		step:    newStep(name, exe, message, proj, args),
		timeout: timeout,
	}

	if s.timeout == 0 {
		s.timeout = 30 * time.Second
	}

	return s
}

func (s timeoutStep) execute() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, s.exe, s.args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Dir = s.proj

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", &stepErr{
				step:  s.name,
				msg:   "failed timeout",
				cause: context.DeadlineExceeded,
			}
		}

		return "", &stepErr{
			step:  s.name,
			msg:   "failed to execute",
			cause: err,
		}
	}

	return s.message, nil
}
