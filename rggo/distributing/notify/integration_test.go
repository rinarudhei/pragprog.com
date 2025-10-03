//go:build integration

package notify

import "testing"

func TestSend(t *testing.T) {
	n := New("Title", "message", SeverityNormal)
	if err := n.Send(); err != nil {
		t.Error(err)
	}
}
