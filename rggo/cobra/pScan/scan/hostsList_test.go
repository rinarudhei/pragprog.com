package scan_test

import (
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

			err := hl.Add(tc.host)

			if err != tc.expectErr {
				t.Fatalf("expect %v, got %v", tc.expectErr, err)
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
