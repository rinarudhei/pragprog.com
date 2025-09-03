package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"pragprog.com/rggo/interacting/todo"
)

func setupAPI(t *testing.T) (string, func()) {
	t.Helper()
	fileTemp, err := os.CreateTemp("", "tmp_*")
	if err != nil {
		t.Fatal(err)
	}
	defer fileTemp.Close()
	ls := &todo.List{}
	ls.Add("Test activity 1")
	ls.Add("Test activity 2")
	ls.Add("Test activity 3")
	ls.Save(fileTemp.Name())
	server := httptest.NewServer(newMux(fileTemp.Name()))

	return server.URL, func() {
		server.Close()
		os.Remove(fileTemp.Name())
	}
}

func TestGet(t *testing.T) {
	testCases := []struct {
		name       string
		path       string
		expCode    int
		expContent string
	}{
		{name: "GetRoot", path: "/", expCode: http.StatusOK, expContent: "There's an API here"},
		{name: "GetTodos", path: "/todo", expCode: http.StatusOK, expContent: "Test activity 1"},
		{name: "GetTodo", path: "/todo/2", expCode: http.StatusOK, expContent: "Test activity 2"},
		{name: "NotFound", path: "/unknown-path", expCode: http.StatusNotFound},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url, cleanUp := setupAPI(t)
			defer cleanUp()
			r, err := http.Get(url + tc.path)
			if err != nil {
				t.Error(err)
			}
			defer r.Body.Close()

			if r.StatusCode != tc.expCode {
				t.Errorf("expect http status %q, got %q", http.StatusText(tc.expCode), http.StatusText(r.StatusCode))
			}
			var body []byte
			var errReadBody error
			switch {
			case strings.Contains(r.Header.Get("Content-Type"), "text/plain"):
				if body, errReadBody = io.ReadAll(r.Body); errReadBody != nil {
					t.Fatal(errReadBody)
				}

			case strings.Contains(r.Header.Get("Content-Type"), "application/json"):
				if body, errReadBody = io.ReadAll(r.Body); errReadBody != nil {
					t.Fatal(errReadBody)
				}

			}

			if !strings.Contains(string(body), tc.expContent) {
				t.Errorf("expect content %q, got %q", tc.expContent, string(body))
			}
		})
	}
}

func TestAdd(t *testing.T) {
	url, cleanup := setupAPI(t)
	defer cleanup()

	var buf bytes.Buffer
	taskName := "my task"
	task := struct {
		Task string `json:"task"`
	}{
		Task: taskName,
	}

	t.Run("TestAdd", func(t *testing.T) {
		if err := json.NewEncoder(&buf).Encode(task); err != nil {
			t.Error(err)
		}
		r, err := http.Post(url+"/todo", "application/json", &buf)
		if err != nil {
			t.Error(err)
		}
		defer r.Body.Close()

		if r.StatusCode != http.StatusCreated {
			t.Errorf("expect status %q, got %q", http.StatusText(http.StatusCreated), http.StatusText(r.StatusCode))
		}
	})

	t.Run("CheckAdd", func(t *testing.T) {
		r, err := http.Get(url + "/todo")
		if err != nil {
			t.Error(err)
		}
		defer r.Body.Close()

		if r.StatusCode != http.StatusOK {
			t.Errorf("expect http status %q, got %q", http.StatusText(http.StatusOK), http.StatusText(r.StatusCode))
		}

		resp := &todoResponse{}
		if err := json.NewDecoder(r.Body).Decode(resp); err != nil {
			t.Error(err)
		}
		r.Body.Close()

		if len(resp.Results) != 4 {
			t.Errorf("expect list length 4 after adding item")
		}

		if resp.Results[3].Task != taskName {
			t.Errorf("expect task name %q, got %q", taskName, resp.Results[0].Task)
		}
	})
}

func TestDelete(t *testing.T) {
	url, cleanup := setupAPI(t)
	defer cleanup()
	t.Run("TestDelete", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, url+"/todo/2", nil)
		if err != nil {
			t.Error(err)
		}

		r, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
		}

		if r.StatusCode != http.StatusNoContent {
			t.Errorf("expect status %q, got %q", http.StatusText(http.StatusNoContent), http.StatusText(r.StatusCode))
		}
	})

	t.Run("CheckDelete", func(t *testing.T) {
		r, err := http.Get(url + "/todo")
		if err != nil {
			t.Error(err)
		}
		defer r.Body.Close()

		if r.StatusCode != http.StatusOK {
			t.Errorf("expect http status %q, got %q", http.StatusText(http.StatusOK), http.StatusText(r.StatusCode))
		}

		resp := &todoResponse{}
		if err := json.NewDecoder(r.Body).Decode(resp); err != nil {
			t.Error(err)
		}
		r.Body.Close()

		if len(resp.Results) != 2 {
			t.Errorf("expect list length 4 after adding item")
		}

		if resp.Results[1].Task != "Test activity 3" {
			t.Errorf("expect task name Test activity 3, got %q", resp.Results[0].Task)
		}
	})
}

func TestPatch(t *testing.T) {
	url, cleanup := setupAPI(t)
	defer cleanup()
	t.Run("TestPatch", func(t *testing.T) {
		r, err := http.NewRequest(http.MethodPatch, url+"/todo/2?complete=true", nil)
		if err != nil {
			t.Error(err)
		}

		resp, err := http.DefaultClient.Do(r)
		if err != nil {
			t.Error(err)
		}

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("expect status %q, got %q", http.StatusText(http.StatusNoContent), http.StatusText(resp.StatusCode))
		}
	})

	t.Run("CheckPatch", func(t *testing.T) {
		r, err := http.Get(url + "/todo/2")
		if err != nil {
			t.Error(err)
		}
		defer r.Body.Close()

		if r.StatusCode != http.StatusOK {
			t.Errorf("expect http status %q, got %q", http.StatusText(http.StatusOK), http.StatusText(r.StatusCode))
		}

		resp := &todoResponse{}
		if err := json.NewDecoder(r.Body).Decode(resp); err != nil {
			t.Error(err)
		}
		r.Body.Close()

		if !resp.Results[0].Done {
			t.Errorf("expect task %q is done", resp.Results[0].Task)
		}
	})
}
