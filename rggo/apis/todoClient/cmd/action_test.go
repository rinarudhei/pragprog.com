// go: build !integration

package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"
)

func TestListAction(t *testing.T) {
	testCases := []struct {
		name     string
		expError error
		expOut   string
		resp     struct {
			Status int
			Body   string
		}
		closeServer bool
	}{
		{
			name:     "Results",
			expError: nil,
			expOut:   "-  1  Task 1\n-  2  Task 2\n",
			resp:     testResp["resultsMany"],
		},
		{
			name:     "NotFound",
			expError: ErrNotFound,
			resp:     testResp["notFound"],
		},
		{
			name:        "InvalidURL",
			expError:    ErrConnection,
			resp:        testResp["noResults"],
			closeServer: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url, cleanup := mockServer(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.resp.Status)
				fmt.Fprintln(w, tc.resp.Body)
			})
			defer cleanup()
			if tc.closeServer {
				cleanup()
			}

			var out bytes.Buffer
			err := listAction(&out, url, false)

			if tc.expError != nil {
				if err == nil {
					t.Fatalf("Expected error %q, got no error.", tc.expError)
				}
				if !errors.Is(err, tc.expError) {
					t.Errorf("Expected error %q, got %q.", tc.expError, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Expected no error, got %q.", err)
			}
			if tc.expOut != out.String() {
				t.Errorf("Expected output %q, got %q", tc.expOut, out.String())
			}
		})
	}
}

func TestViewAction(t *testing.T) {
	testCases := []struct {
		name     string
		expError error
		expOut   string
		resp     struct {
			Status int
			Body   string
		}
		id string
	}{
		{
			name:   "ResultsOne",
			expOut: "Task:         Task 1\nCreated at:   Oct/28 @08:23\nCompleted:    No\n",
			resp:   testResp["resultsOne"],
			id:     "1",
		},
		{
			name:     "NotFound",
			expError: ErrNotFound,
			resp:     testResp["notFound"],
			id:       "1",
		},
		{
			name:     "InvalidID",
			expError: ErrNotNumber,
			resp:     testResp["noResults"],
			id:       "a",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url, cleanup := mockServer(
				func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tc.resp.Status)
					fmt.Fprintln(w, tc.resp.Body)
				})
			defer cleanup()
			var out bytes.Buffer
			err := viewAction(&out, url, tc.id)
			if tc.expError != nil {
				if err == nil {
					t.Fatalf("Expected error %q, got no error.", tc.expError)
				}
				if !errors.Is(err, tc.expError) {
					t.Errorf("Expected error %q, got %q.", tc.expError, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Expected no error, got %q.", err)
			}
			if tc.expOut != out.String() {
				t.Errorf("Expected output %q, got %q", tc.expOut, out.String())
			}
		})
	}
}

func TestAddAction(t *testing.T) {
	expURLPath := "/todo"
	expMethod := http.MethodPost
	expBody := "{\"task\":\"Task 1\"}\n"
	expContentType := "application/json"
	expOut := "Added task \"Task 1\" to the list.\n"
	args := []string{"Task", "1"}

	url, cleanup := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != expURLPath {
			t.Errorf("Expected path %q, got %q", expURLPath, r.URL.Path)
		}
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		defer r.Body.Close()

		if string(b) != expBody {
			t.Errorf("Expect body %s, got %s", expBody, string(b))
		}
		if r.Method != expMethod {
			t.Errorf("Expected method %q, got %q", expMethod, r.Method)
		}

		contentType := r.Header.Get("Content-Type")
		if contentType != expContentType {
			t.Errorf("Expected Content-Type %q, got %q",
				expContentType, contentType)
		}

		w.WriteHeader(testResp["created"].Status)
		fmt.Fprintln(w, testResp["created"].Body)
	})
	defer cleanup()

	var out bytes.Buffer
	if err := addAction(&out, url, args); err != nil {
		t.Errorf("Expect no error, got %v", err)
	}
	if expOut != out.String() {
		t.Errorf("Expect %s, got %s", expOut, out.String())
	}
}

func TestCompleteAction(t *testing.T) {
	expURLPath := "/todo/1"
	expQueryKey := "complete"
	expQueryValue := "true"
	expMethod := http.MethodPatch
	expOut := "Task 1 completed.\n"
	args := []string{"1"}

	url, cleanup := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != expURLPath {
			t.Errorf("Expected path %q, got %q", expURLPath, r.URL.Path)
		}
		if r.URL.Query().Get(expQueryKey) != expQueryValue {
			t.Errorf("Expected query %q equal %q, got %q", expQueryKey, expQueryValue, r.URL.Query().Get(expQueryKey))
		}
		if r.Method != expMethod {
			t.Errorf("Expected method %q, got %q", expMethod, r.Method)
		}

		w.WriteHeader(testResp["noContent"].Status)
		fmt.Fprintln(w, testResp["noContent"].Body)
	})
	defer cleanup()

	var out bytes.Buffer
	if err := completeAction(&out, url, args[0]); err != nil {
		t.Errorf("Expect no error, got %v", err)
	}
	if expOut != out.String() {
		t.Errorf("Expect %s, got %s", expOut, out.String())
	}
}

func TestDeleteAction(t *testing.T) {
	expURLPath := "/todo/1"
	expMethod := http.MethodDelete
	expOut := "Deleted task id 1.\n"
	args := []string{"1"}

	url, cleanup := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != expURLPath {
			t.Errorf("Expected path %q, got %q", expURLPath, r.URL.Path)
		}
		if r.Method != expMethod {
			t.Errorf("Expected method %q, got %q", expMethod, r.Method)
		}

		w.WriteHeader(testResp["noContent"].Status)
		fmt.Fprintln(w, testResp["noContent"].Body)
	})
	defer cleanup()

	var out bytes.Buffer
	if err := deleteAction(&out, url, args[0]); err != nil {
		t.Errorf("Expect no error, got %v", err)
	}
	if expOut != out.String() {
		t.Errorf("Expect %s, got %s", expOut, out.String())
	}
}
