package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"pragprog.com/rggo/interacting/todo"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		replyError(w, r, http.StatusNotFound, "path not found")
		return
	}

	content := "There's an API here"
	replyTextContent(w, r, http.StatusOK, content)
}

var (
	ErrNotFound    = errors.New("not found")
	ErrInvalidData = errors.New("invalid data")
)

func todoRouter(todoFile string, l sync.Locker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list := &todo.List{}
		l.Lock()
		defer l.Unlock()
		if err := list.Get(todoFile); err != nil {
			replyError(w, r, http.StatusInternalServerError, err.Error())
		}

		if r.URL.Path == "" {
			switch r.Method {
			case http.MethodGet:
				getAllHandler(w, r, list)
			case http.MethodPost:
				addHandler(w, r, list, todoFile)
			default:
				message := "Method not supported"
				replyError(w, r, http.StatusMethodNotAllowed, message)
			}
			return
		}

		id, err := validateID(r.URL.Path, list)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				replyError(w, r, http.StatusNotFound, err.Error())
			}
			replyError(w, r, http.StatusBadRequest, err.Error())
		}

		switch r.Method {
		case http.MethodGet:
			getOneHandler(w, r, list, id)
		case http.MethodDelete:
			deleteHandler(w, r, list, id, todoFile)
		case http.MethodPatch:
			patchHandler(w, r, list, id, todoFile)
		default:
			message := "Method not supported"
			replyError(w, r, http.StatusMethodNotAllowed, message)
		}
	}
}

func getAllHandler(w http.ResponseWriter, r *http.Request, list *todo.List) {
	resp := &todoResponse{
		Results: *list,
	}

	replyJSONContent(w, r, http.StatusOK, resp)
}

func getOneHandler(w http.ResponseWriter, r *http.Request, list *todo.List, id int) {
	ls := *list

	resp := &todoResponse{
		Results: ls[id-1 : id],
	}

	replyJSONContent(w, r, http.StatusOK, resp)
}

func deleteHandler(w http.ResponseWriter, r *http.Request, list *todo.List, id int, todoFile string) {
	if err := list.Delete(id); err != nil {
		replyError(w, r, http.StatusBadRequest, err.Error())
	}

	if err := list.Save(todoFile); err != nil {
		replyError(w, r, http.StatusInternalServerError, "failed saving todo file")
	}

	replyTextContent(w, r, http.StatusNoContent, "")
}

func patchHandler(w http.ResponseWriter, r *http.Request, list *todo.List, id int, todoFile string) {
	queries := r.URL.Query()
	_, ok := queries["complete"]
	if !ok {
		replyError(w, r, http.StatusBadRequest, "invalid query param")
	}

	if err := list.Complete(id); err != nil {
		replyError(w, r, http.StatusInternalServerError, err.Error())
	}

	if err := list.Save(todoFile); err != nil {
		replyError(w, r, http.StatusInternalServerError, "failed saving todo file")
	}

	replyTextContent(w, r, http.StatusNoContent, "")
}

func addHandler(w http.ResponseWriter, r *http.Request, list *todo.List, todoFile string) {
	reqBody := struct {
		Task string `json:"task"`
	}{}

	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		replyError(w, r, http.StatusBadRequest, fmt.Sprintf("invalid json: %s", err.Error()))
	}

	list.Add(reqBody.Task)
	if err := list.Save(todoFile); err != nil {
		replyError(w, r, http.StatusInternalServerError, "error saving new todo")
	}

	replyTextContent(w, r, http.StatusCreated, "")
}

func validateID(path string, list *todo.List) (int, error) {
	id, err := strconv.Atoi(path)
	if err != nil {
		return 0, err
	}

	if id <= 0 || id > len(*list) {
		return 0, fmt.Errorf("invalid id: %d", id)
	}

	return id, nil
}
