package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"

	"pragprog.com/rggo/cobra/pScan/scan"
)

func setup(t *testing.T, hosts []string, initList bool) (string, func()) {
	tf, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	tf.Close()

	if initList {
		hl := &scan.HostList{}
		for _, h := range hosts {
			if err := hl.Add(h); err != nil {
				t.Fatal(err)
			}
		}

		if err := hl.Save(tf.Name()); err != nil {
			t.Fatal(err)
		}
	}

	return tf.Name(), func() {
		os.Remove(tf.Name())
	}
}

func TestHostAction(t *testing.T) {
	hosts := []string{"host1", "host2", "host3"}

	testcases := []struct {
		name           string
		args           []string
		expectedOut    string
		initList       bool
		actionFunction func(io.Writer, string, []string) error
	}{
		{name: "add action", args: hosts, expectedOut: "Added host: host1\nAdded host: host2\nAdded host: host3\n", initList: false, actionFunction: addAction},
		{name: "init list", expectedOut: "host1\nhost2\nhost3\n", initList: true, actionFunction: listAction},
		{name: "remove action", args: hosts, expectedOut: "Deleted host: host1\nDeleted host: host2\nDeleted host: host3\n", initList: true, actionFunction: deleteAction},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			hostFile, cleanupFunc := setup(t, hosts, tc.initList)
			defer cleanupFunc()

			var buf bytes.Buffer
			if err := tc.actionFunction(&buf, hostFile, tc.args); err != nil {
				t.Errorf("expect no error, got %v", err)
			}

			actualOut := buf.String()
			if tc.expectedOut != actualOut {
				t.Errorf("expect %s, got %s", tc.expectedOut, actualOut)
			}
		})
	}
}

func TestIntegration(t *testing.T) {
	hosts := []string{"host1", "host2", "host3"}
	hostFile, cleanupFunc := setup(t, hosts, false)
	defer cleanupFunc()

	delHost := "host2"
	hostsEnd := []string{"host1", "host3"}

	var buf bytes.Buffer
	expectedOutput := ""
	for _, v := range hosts {
		expectedOutput += fmt.Sprintf("Added host: %s\n", v)
	}
	expectedOutput += strings.Join(hosts, "\n")
	expectedOutput += fmt.Sprintln()
	expectedOutput += fmt.Sprintf("Deleted host: %s\n", delHost)
	expectedOutput += strings.Join(hostsEnd, "\n")
	expectedOutput += fmt.Sprintln()
	for _, v := range hostsEnd {
		expectedOutput += fmt.Sprintf("%s: Host not found\n", v)
		expectedOutput += fmt.Sprintln()
	}

	if err := addAction(&buf, hostFile, hosts); err != nil {
		t.Fatalf("expect not error, got %v", err)
	}

	if err := listAction(&buf, hostFile, nil); err != nil {
		t.Fatalf("expect not error, got %v", err)
	}

	if err := deleteAction(&buf, hostFile, []string{delHost}); err != nil {
		t.Fatalf("expect not error, got %v", err)
	}

	if err := listAction(&buf, hostFile, nil); err != nil {
		t.Fatalf("expect not error, got %v", err)
	}

	if err := scanAction(&buf, hostFile, nil); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}

	if buf.String() != expectedOutput {
		t.Errorf("expect %s, got %s", expectedOutput, buf.String())
	}
}

func TestScanAction(t *testing.T) {
	hosts := []string{
		"localhost",
		"unknownhostoutthere",
	}

	tf, cleanup := setup(t, hosts, true)
	defer cleanup()

	ports := []int{}

	for i := 0; i < 2; i++ {
		ln, err := net.Listen("tcp", net.JoinHostPort("localhost", "0"))
		if err != nil {
			t.Fatal(err)
		}
		defer ln.Close()

		_, portStr, err := net.SplitHostPort(ln.Addr().String())
		if err != nil {
			t.Fatal(err)
		}

		port, err := strconv.Atoi(portStr)
		if err != nil {
			t.Fatal(err)
		}

		ports = append(ports, port)
		if i == 1 {
			ln.Close()
		}
	}

	expectedOut := fmt.Sprintln("localhost:")
	expectedOut += fmt.Sprintf("\t%d: open\n", ports[0])
	expectedOut += fmt.Sprintf("\t%d: closed\n", ports[1])
	expectedOut += fmt.Sprintln()
	expectedOut += fmt.Sprintln("unknownhostoutthere: Host not found")
	expectedOut += fmt.Sprintln()

	// Define var to capture scan output
	var out bytes.Buffer
	// Execute scan and capture output
	if err := scanAction(&out, tf, ports); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}
	// Test scan output
	if out.String() != expectedOut {
		t.Errorf("Expected output %q, got %q\n", expectedOut, out.String())
	}
}
