package scan_test

import (
	"errors"
	"os"
	"testing"

	"pragprog.com/rggo/cobra/pScan/scan"
)

func TestAdd(t *testing.T) {
	testcases := []struct {
		name      string
		host      string
		expectLen int
		expectErr error
	}{
		{name: "Add new host", host: "host2", expectLen: 2, expectErr: nil},
		{name: "Add existing", host: "host1", expectLen: 1, expectErr: scan.ErrExists},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			hl := &scan.HostList{}

			if err := hl.Add("host1"); err != nil {
				t.Fatal("error add host1")
			}

			if err := hl.Add(tc.host); err != nil {
				if tc.expectErr != nil {
					if !errors.Is(err, tc.expectErr) {
						t.Fatalf("expect %v, got %v", tc.expectErr, err)
					}

					return
				}

				t.Fatalf("unexpected error: %v, host: %s", err, tc.host)
			}

			if len(hl.Hosts) != tc.expectLen {
				t.Fatalf("expect Hosts length %d, got %d", tc.expectLen, len(hl.Hosts))
			}

			if hl.Hosts[1] != tc.host {
				t.Fatalf("expect host %s, got %s", tc.host, hl.Hosts[1])
			}
		})
	}
}

func TestRemove(t *testing.T) {
	testcases := []struct {
		name      string
		host      string
		expectLen int
		expectErr error
	}{
		{name: "remove existing", host: "host1", expectLen: 1, expectErr: nil},
		{name: "remove not found", host: "host3", expectLen: 2, expectErr: scan.ErrNotExists},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			hl := &scan.HostList{}

			if err := hl.Add("host1"); err != nil {
				t.Fatal("error add host1")
			}

			if err := hl.Add("host2"); err != nil {
				t.Fatal("error add host2")
			}

			if err := hl.Remove(tc.host); err != nil {
				if tc.expectErr != nil {
					if !errors.Is(tc.expectErr, scan.ErrNotExists) {
						t.Fatalf("expect error: %v, got: %v", tc.expectErr, err)
					}

					return
				}

				t.Fatalf("unexpected error: %v, host: %s", err, tc.host)
			}

			if len(hl.Hosts) != tc.expectLen {
				t.Fatalf("expect Hosts length %d, got %d", tc.expectLen, len(hl.Hosts))
			}
		})
	}
}

func TestSaveLoad(t *testing.T) {
	hl1 := &scan.HostList{}
	hl2 := &scan.HostList{}

	if err := hl1.Add("host1"); err != nil {
		t.Fatalf("error adding host1: %v", err)
	}
	if err := hl1.Add("host2"); err != nil {
		t.Fatalf("error adding host2: %v", err)
	}

	tf, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatalf("error create temp file")
	}
	defer os.Remove(tf.Name())

	if err := hl1.Save(tf.Name()); err != nil {
		t.Fatalf("error saving hostList: %v", err)
	}

	if err := hl2.Load(tf.Name()); err != nil {
		t.Fatalf("error loading hostList: %v", err)
	}

	if len(hl1.Hosts) != len(hl2.Hosts) {
		t.Fatalf("first HostList lenght is %d, second HostList length is %d", len(hl1.Hosts), len(hl2.Hosts))
	}

	for i, h1 := range hl1.Hosts {
		if h1 != hl2.Hosts[i] {
			t.Fatalf("host 1: %s, host 2: %s", h1, hl2.Hosts[i])
		}
	}
}

func TestLoadNoFile(t *testing.T) {
	hl := &scan.HostList{}
	tf, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatal("error create temp file")
	}
	if err := os.Remove(tf.Name()); err != nil {
		t.Fatalf("error removing temp file: %s", tf.Name())
	}

	if err := hl.Load(tf.Name()); err != nil {
		t.Fatal("expect error, got nil")
	}
}
