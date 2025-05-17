package todo_test

import (
	"os"
	"testing"

	"pragprog.com/rggo/interacting/todo"
)

func TestAdd(t *testing.T) {
	var l todo.List
	expectedTask := "read book"

	l.Add(expectedTask)

	if l[0].Task != expectedTask {
		t.Fatalf("actual task %s, expected task %s", l[0].Task, expectedTask)
	}
}


func TestComplete(t *testing.T) {
	var l todo.List
	task := "new task"
	l.Add(task)

	if l[0].Task != task {
		t.Fatalf("actual task %s, expected task %s", l[0].Task, task)
	}

	if l[0].Done {
		t.Fatalf("new task should not be completed")
	}

	l.Complete(1)
	if !l[0].Done {
		t.Fatalf("new task should be completed")
	}
}

func TestDelete(t *testing.T) {
	var l todo.List

	tasks := []string{"read book", "bug-fix", "exercise"}
	for _, task := range tasks {
		l.Add(task)
	}

	if l[0].Task != tasks[0] {
		t.Fatalf("actual %s, expected %s", l[0].Task, tasks[0])
	}

	l.Delete(2)

	if len(l) != 2 {
		t.Fatalf("expected list length %d, got %d instead", 2, len(l))
	}

	if l[1].Task != tasks[2] {
		t.Fatalf("expected task name %s, got %s instead", tasks[2], l[1].Task)
	}
}

func TestSaveGet(t *testing.T) {
	var l1 todo.List
	var l2 todo.List
	task := "New task for l1"
	l1.Add(task)

	if l1[0].Task != task {
		t.Fatalf("expected l1 new task: %s, actual: %s", task, l1[0].Task)
	}

	tmpFile, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatalf("error creating temp file: %s", err)
	}
	defer os.Remove(tmpFile.Name())

	if err := l1.Save(tmpFile.Name()); err != nil {
		t.Fatalf("error when saving file: %s", err)
	}

	if err := l2.Get(tmpFile.Name()); err != nil {
		t.Fatalf("error when getting file: %s", err)
	}

	if l1[0].Task != l2[0].Task {
		t.Fatalf("task %s should match %s task", l2[0].Task, l1[0].Task)
	}
}
