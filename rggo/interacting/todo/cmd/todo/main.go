package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"pragprog.com/rggo/interacting/todo"
)

func main() {
	todoFileName := os.Getenv("TODO_FILE_NAME_ENV")
	if todoFileName == "" {
		todoFileName = ".todo.json"
	}
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			"%s tool. Developed for The Pragmatic Bookshelf\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "Copyright 2020\n")
		fmt.Fprintln(flag.CommandLine.Output(), "Usage information:")
		fmt.Fprintln(flag.CommandLine.Output(), "use -add task_name to add task")
		fmt.Fprintln(flag.CommandLine.Output(), "use -add without argument to add task by prompt")
		flag.PrintDefaults()
	}
	add := flag.Bool("add", false, "add task into todo list")
	list := flag.Bool("list", false, "list all tasks")
	ulist := flag.Bool("ulist", false, "show only uncompleted tasks")
	complete := flag.Int("complete", 0, "item to be completed")
	verbose := flag.Bool("verbose", false, "verbose")
	del := flag.Int("del", 0, "item to be deleted")
	flag.Parse()
	l := &todo.List{}

	if err := l.Get(todoFileName); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	switch {
	case *list:
		fmt.Print(l)
	case *ulist:
		ul := make(todo.List, 0, len(*l))
		for _, item := range *l {
			if !item.Done {
				ul = append(ul, item)
			}
		}
		fmt.Print(&ul)
	case *verbose:
		formatted := "             Done  CreatedAt           CompletedAt\n"
		for k, t := range *l {
			status := " "
			taskDisplay := []byte("          ")
			for i, b := range []byte(t.Task) {
				taskDisplay[i] = b
			}
			if t.Done {
				status = "X"
			}
			formatted += fmt.Sprintf("%d: %s %v    %v %v\n", k+1, taskDisplay, status, t.CreatedAt.Format(time.DateTime), t.CompletedAt.Format(time.DateTime))
		}
		fmt.Print(formatted)
	case *complete > 0:
		if err := l.Complete(*complete); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if err := l.Save(todoFileName); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case *add:
		tasks, err := getTask(os.Stdin, flag.Args()...)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		for _, t := range tasks {
			l.Add(t)
		}
		if err := l.Save(todoFileName); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case *del > 0:
		if err := l.Delete(*del); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if err := l.Save(todoFileName); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		fmt.Fprintln(os.Stderr, "invalid option")
		os.Exit(1)
	}
}
func getTask(r io.Reader, args ...string) ([]string, error) {
	if len(args) > 0 {
		return []string{strings.Join(args, " ")}, nil
	}
	fmt.Println("each task is separated by line")
	fmt.Println("press CTRL+D when finish adding task/s")

	fmt.Println("please input a new task/s:")
	s := bufio.NewScanner(r)
	var out []string
	for s.Scan() {
		if err := s.Err(); err != nil {
			if err == io.EOF {
				return out, nil
			}
			return []string{}, err
		}
		if len(s.Text()) == 0 {
			fmt.Println("task cannot be blank!")
		}
		out = append(out, s.Text())
	}
	fmt.Println("todo list updated...")
	return out, nil
}
