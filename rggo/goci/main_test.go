package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func mockCmdContext(ctx context.Context, exe string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess"}
	cs = append(cs, exe)
	cs = append(cs, args...)

	cmd := exec.CommandContext(ctx, os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func mockCmdTimeout(ctx context.Context, exe string, args ...string) *exec.Cmd {
	cmd := mockCmdContext(ctx, exe, args...)

	cmd.Env = append(cmd.Env, "GO_HELPER_TIMEOUT=1")
	return cmd
}

func TestHelperProcess(t *testing.T) {
	t.Helper()

	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	if os.Getenv("GO_HELPER_TIMEOUT") == "1" {
		time.Sleep(15 * time.Second)
	}

	if os.Args[2] == "git" {
		fmt.Fprintln(os.Stdout, "Everything up-to-date")
		os.Exit(0)
	}
}

func TestRun(t *testing.T) {
	testCases := []struct {
		name    string
		proj    string
		out     string
		expErr  error
		mockCmd func(ctx context.Context, name string, args ...string) *exec.Cmd
	}{
		{name: "success", proj: "./testdata/tool/", out: "Go Build: SUCCESS\nGo Test: SUCCESS\nGofmt: SUCCESS\nGolint: SUCCESS\nGocyclo: SUCCESS\nGit push: SUCCESS\n", expErr: nil, mockCmd: mockCmdContext},
		{name: "failed timeout", proj: "./testdata/tool/", out: "", expErr: &stepErr{step: "git push"}, mockCmd: mockCmdTimeout},
		{name: "failed", proj: "./testdata/toolErr/", out: "", expErr: &stepErr{step: "go build"}},
		{name: "failed gofmt", proj: "./testdata/toolFmtErr/", out: "", expErr: &stepErr{step: "go fmt"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mockCmd != nil {
				command = tc.mockCmd
			}
			var out bytes.Buffer

			err := run(tc.proj, &out)
			if tc.expErr != nil {
				if err == nil {
					t.Errorf("Expected error: %q. Got 'nil' instead.", tc.expErr)
				}
				if !errors.Is(err, tc.expErr) {
					t.Errorf("Expected error: %q. Got %q.", tc.expErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %q", err)
			}

			if out.String() != tc.out {
				t.Errorf("Expected output: %q. Got %q", tc.out, out.String())
			}
		})
	}
}

func TestRunKill(t *testing.T) {
	testCases := []struct {
		name   string
		proj   string
		sig    syscall.Signal
		expErr error
	}{
		{"SIGNINT", "./testdata/tool", syscall.SIGINT, ErrSignal},
		{"SIGTERM", "./testdata/tool", syscall.SIGTERM, ErrSignal},
		{"SIGQUIT", "./testdata/tool", syscall.SIGQUIT, nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			command = mockCmdTimeout
			errCh := make(chan error)
			ignSigCh := make(chan os.Signal, 1)
			expSigCh := make(chan os.Signal, 1)

			signal.Notify(ignSigCh, syscall.SIGQUIT)
			defer signal.Stop(ignSigCh)
			signal.Notify(expSigCh, tc.sig)
			defer signal.Stop(expSigCh)

			go func() {
				errCh <- run(tc.proj, io.Discard)
			}()
			go func() {
				time.Sleep(2 * time.Second)
				syscall.Kill(syscall.Getpid(), tc.sig)
			}()

			select {
			case err := <-errCh:
				if err == nil {
					t.Errorf("Expected signal %q, got %q", tc.expErr, err)
				}

				if !errors.Is(err, tc.expErr) {
					t.Errorf("Expected error: %q. Got %q", tc.expErr, err)
				}

				select {
				case rec := <-expSigCh:
					if rec != tc.sig {
						t.Errorf("expected %q, got %q", tc.sig, rec)
					}
				default:
					t.Error("signal not received")
				}

			case <-ignSigCh:
			}
		})
	}
}
