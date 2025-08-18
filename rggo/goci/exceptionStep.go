package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

type exceptionStep struct {
	step
}

func newExceptionStep(name, exe, message, proj string, args []string) exceptionStep {
	return exceptionStep{
		step: newStep(name, exe, message, proj, args),
	}
}

func (s exceptionStep) execute() (string, error) {
	cmd := exec.Command(s.exe, s.args...)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Dir = s.proj

	if err := cmd.Run(); err != nil {
		return "", &stepErr{
			step:  s.name,
			msg:   "failed to execute",
			cause: nil,
		}
	}
	var errMessage string
	isGoFmtSuccess := out.Len() == 0
	isGoLintSuccess := out.String() == "0 issues.\n"
	isError := !isGoFmtSuccess && !isGoLintSuccess
	if !isGoFmtSuccess {
		errMessage = "invalid format"
	} else {
		errMessage = "error linter"
	}

	if isError {
		return "", &stepErr{
			step:  s.name,
			msg:   fmt.Sprintf("%s: %s", errMessage, out.String()),
			cause: nil,
		}
	}

	return s.message, nil
}
