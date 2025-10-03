//go:build !integration

package notify

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	testCases := []struct {
		severity Severity
	}{
		{SeverityLow},
		{SeverityNormal},
		{SeverityUrgent},
	}

	for _, tc := range testCases {
		t.Run(tc.severity.String(), func(t *testing.T) {
			expTitle := "My title"
			expMessage := "my message"
			n := New(expTitle, expMessage, tc.severity)

			if n.title != expTitle {
				t.Errorf("expect title %s, got %s", expTitle, n.title)
			}
			if n.message != expMessage {
				t.Errorf("expect message %s, got %s", expMessage, n.message)
			}
			if n.severity != tc.severity {
				t.Errorf("expect severity %d, got %d", tc.severity, n.severity)
			}
		})
	}
}

func TestSeverityString(t *testing.T) {
	testCases := []struct {
		os           string
		severity     Severity
		expSevString string
	}{
		{"linux", SeverityLow, "low"},
		{"linux", SeverityNormal, "normal"},
		{"linux", SeverityUrgent, "critical"},
		{"darwin", SeverityLow, "Low"},
		{"darwin", SeverityNormal, "Normal"},
		{"darwin", SeverityUrgent, "Critical"},
		{"windows", SeverityLow, "Info"},
		{"windows", SeverityNormal, "Warning"},
		{"windows", SeverityUrgent, "Error"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s:%s", tc.os, tc.severity), func(t *testing.T) {
			if runtime.GOOS != tc.os {
				t.Skip()
			}
			n := New("my title", "my message", tc.severity)
			if n.severity.String() != tc.expSevString {
				t.Errorf("expect severity string %s, got %s", tc.expSevString, n.severity.String())
			}
		})
	}
}

func mockCmd(exe string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess"}
	cs = append(cs, exe)
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	cmdName := ""
	switch runtime.GOOS {
	case "linux":
		cmdName = "notify-send"
	case "darwin":
		cmdName = "terminal-notifier"
	case "windows":
		cmdName = "powershell"
	}
	time.Sleep(5 * time.Second)
	if strings.Contains(os.Args[2], cmdName) {
		os.Exit(0)
	}

	os.Exit(1)
}

func TestSend(t *testing.T) {
	command = mockCmd
	n := New("title", "message", SeverityLow)
	if err := n.Send(); err != nil {
		t.Error(err)
	}
}
