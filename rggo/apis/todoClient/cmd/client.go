package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var (
	ErrConnection      = errors.New("Connection error")
	ErrNotFound        = errors.New("Not found")
	ErrInvalidResponse = errors.New("Invalid data")
	ErrNotNumber       = errors.New("Not a number")
)

type item struct {
	Task        string
	Done        bool
	CreatedAt   time.Time
	CompletedAt time.Time
}

type response struct {
	Result       []item `json:"results"`
	Date         int    `json:"date"`
	TotalResults int    `json:"total_results"`
}

func newClient() *http.Client {
	c := &http.Client{
		Timeout: 10 * time.Second,
	}

	return c
}

func getItems(url string) ([]item, error) {
	r, err := newClient().Get(url)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrConnection, err)
	}

	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		msg, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, fmt.Errorf("Cannot read body: %w", err)
		}

		err = ErrInvalidResponse
		if r.StatusCode == http.StatusNotFound {
			err = ErrNotFound
		}

		return nil, fmt.Errorf("%w: %s", err, msg)
	}

	var resp response

	if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
		return nil, err
	}

	return resp.Result, nil
}

func getAll(apiRoot string) ([]item, error) {
	u := fmt.Sprintf("%s/todo", apiRoot)

	return getItems(u)
}

func getOne(apiRoot string, id int) (item, error) {
	u := fmt.Sprintf("%s/todo/%d", apiRoot, id)

	items, err := getItems(u)
	if err != nil {
		return item{}, err
	}

	if len(items) != 1 {
		return item{}, fmt.Errorf("%w: Invalid results", ErrInvalidResponse)
	}

	return items[0], nil
}

const timeFormat = "Jan/02 @15:04"

func sendRequest(url, method, contentType string, expStatus int, body io.Reader) error {
	r, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}

	if contentType != "" {
		r.Header.Set("Content-Type", contentType)
	}

	resp, err := newClient().Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != expStatus {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		err = ErrInvalidResponse
		if resp.StatusCode == http.StatusNotFound {
			err = ErrNotFound
		}

		return fmt.Errorf("%w: %s", err, b)
	}

	return nil
}

func addItem(apiRoot, task string) error {
	t := struct {
		Task string `json:"task"`
	}{
		Task: task,
	}
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(t); err != nil {
		return err
	}

	return sendRequest(apiRoot+"/todo", http.MethodPost, "application/json", http.StatusCreated, &body)
}

func completeItem(apiRoot, id string) error {
	url := fmt.Sprintf("%s/todo/%s?complete=true", apiRoot, id)
	return sendRequest(url, http.MethodPatch, "", http.StatusNoContent, nil)
}

func deleteItem(apiRoot, id string) error {
	url := fmt.Sprintf("%s/todo/%s", apiRoot, id)
	return sendRequest(url, http.MethodDelete, "", http.StatusNoContent, nil)
}
