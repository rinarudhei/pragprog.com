package main_test

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

var (
	binName  = "todo"
	fileName = ".test.json"
)

func TestMain(m *testing.M) {
	fmt.Println("building tool ...")

	if runtime.GOOS == "windows" {
		binName += ".exe"
	}

	build := exec.Command("go", "build", "-o", binName)

	if err := build.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "cannot build tool %s: %s", binName, err)
	}

	fmt.Println("running tests ...")

	code := m.Run()

	fmt.Println("cleaning up ...")
	os.Remove(binName)
	os.Remove(fileName)
	os.Unsetenv("TODO_FILE_NAME_ENV")
	os.Exit(code)
}

func TestTodoCLI(t *testing.T) {
	t.Setenv("TODO_FILE_NAME_ENV", fileName)

	task := "test task number 1"
	task2 := "test task number 2"

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("error get directory: %s", err)
	}

	cmdPath := filepath.Join(dir, binName)

	t.Run("add new task", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-add", task)

		if err := cmd.Run(); err != nil {
			t.Fatalf("error running cmd: %s", err)
		}
	})
	t.Run("add new task from os.Stdin", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-add")
		cmdStdin, err := cmd.StdinPipe()
		if err != nil {
			t.Fatal(err)
		}

		io.WriteString(cmdStdin, task2)
		cmdStdin.Close()

		if err := cmd.Run(); err != nil {
			t.Fatalf("error running cmd: %s", err)
		}
	})

	t.Run("complete a task", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-complete", "2")

		if err := cmd.Run(); err != nil {
			t.Fatalf("error running cmd: %s", err)
		}
	})

	t.Run("list tasks with a completed task", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-list")

		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatal(err)
		}

		expected := fmt.Sprintf("  1: %s\nX 2: %s\n", task, task2)
		if string(out) != expected {
			t.Errorf("expect %s, instead got %s", expected, string(out))
		}

	})

	t.Run("delete a task", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-del", "1")

		if err := cmd.Run(); err != nil {
			t.Fatalf("error runing cmd: %s", err)
		}
	})

	t.Run("list task after delete a task", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-list")

		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatal(err)
		}

		expected := fmt.Sprintf("X 1: %s\n", task2)
		if string(out) != expected {
			t.Errorf("expect %s, instead got %s", expected, string(out))
		}

	})


}
