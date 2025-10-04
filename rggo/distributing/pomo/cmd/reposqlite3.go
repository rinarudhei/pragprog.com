//go:build !inmemory && !containers

package cmd

import (
	"github.com/spf13/viper"
	"pragprog.com/rggo/interactiveTools/pomo/pomodoro"
	"pragprog.com/rggo/interactiveTools/pomo/pomodoro/repository"
)

func getRepo() (pomodoro.Repository, error) {
	return repository.NewSQLite3Repo(viper.GetString("db"))
}
