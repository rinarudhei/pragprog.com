package main

import (
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
		// {name: "GetRoot", path: "/", expCode: http.StatusOK, expContent: "There's an API here"},
		{name: "GetTodos", path: "/todo", expCode: http.StatusOK, expContent: `{"task":"Test ativity 1"}`},
		// {name: "GetTodo", path: "/todo/2", expCode: http.StatusOK, expContent: "There's an API here"},
		// {name: "NotFound", path: "/unknown-path", expCode: http.StatusNotFound},
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
				t.Errorf("Expect http status %q, got %q", http.StatusText(tc.expCode), http.StatusText(r.StatusCode))
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
				t.Errorf("Expect content %q, got %q", tc.expContent, string(body))
			}

		})
	}
}
