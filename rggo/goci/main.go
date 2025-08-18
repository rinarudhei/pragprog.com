package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	proj := flag.String("p", "", "project directory path")
	flag.Parse()

	if err := run(*proj, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type executer interface {
	execute() (string, error)
}

func run(proj string, out io.Writer) error {
	if proj == "" {
		return fmt.Errorf("project directory is required: %w", ErrValidation)
	}
	pipeline := make([]executer, 6)
	pipeline[0] = newStep(
		"go build",
		"go",
		"Go Build: SUCCESS",
		proj,
		[]string{"build", ".", "errors"},
	)

	pipeline[1] = newStep(
		"go test",
		"go",
		"Go Test: SUCCESS",
		proj,
		[]string{"test", "-v"},
	)

	pipeline[2] = newExceptionStep(
		"go fmt",
		"gofmt",
		"Gofmt: SUCCESS",
		proj,
		[]string{"-l", "."},
	)

	pipeline[3] = newExceptionStep(
		"go lint",
		"golangci-lint",
		"Golint: SUCCESS",
		proj,
		[]string{"run", "."},
	)

	pipeline[4] = newExceptionStep(
		"go cyclo",
		"golangci-lint",
		"Gocyclo: SUCCESS",
		proj,
		[]string{"run", ".", "--enable-only", "gocyclo"},
	)

	pipeline[5] = newTimeoutStep(
		"git push",
		"git",
		"Git push: SUCCESS",
		proj,
		[]string{"push", "origin", "master"},
		10*time.Second,
	)

	sig := make(chan os.Signal, 1)
	errCh := make(chan error)
	doneCh := make(chan struct{})

	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for _, s := range pipeline {
			msg, err := s.execute()
			if err != nil {
				errCh <- err
				return
			}

			_, err = fmt.Fprintln(out, msg)
			if err != nil {
				errCh <- err
				return
			}
		}

		close(doneCh)
	}()

	for {
		select {
		case rec := <-sig:
			signal.Stop(sig)
			return fmt.Errorf("%s: Exciting: %w", rec, ErrSignal)
		case err := <-errCh:
			return err
		case <-doneCh:
			return nil
		}
	}
}
