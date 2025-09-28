//go:build !inmemory

package pomodoro_test

import (
	"os"
	"testing"

	"pragprog.com/rggo/interactiveTools/pomo/pomodoro"
	"pragprog.com/rggo/interactiveTools/pomo/pomodoro/repository"
)

func getRepo(t *testing.T) (pomodoro.Repository, func()) {
	t.Helper()

	tf, err := os.CreateTemp("", "pomo")
	if err != nil {
		t.Fatal(err)
	}

	dbRepo, err := repository.NewSQLite3Repo(tf.Name())
	if err != nil {
		t.Fatal(err)
	}

	return dbRepo, func() {
		os.Remove(tf.Name())
	}
}
