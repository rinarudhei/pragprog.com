package main

import (
	"encoding/json"
	"time"

	"pragprog.com/rggo/interacting/todo"
)

type todoResponse struct {
	Results todo.List `json:"results"`
}

func (t *todoResponse) MarshallJSON() ([]byte, error) {
	resp := struct {
		Results      todo.List `json:"results"`
		Date         int64     `json:"date"`
		TotalResults int       `json:"total_results"`
	}{
		Results:      t.Results,
		Date:         time.Now().Unix(),
		TotalResults: len(t.Results),
	}

	return json.Marshal(resp)
}
