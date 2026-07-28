package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/samEscom/tasky/render"
	"github.com/samEscom/tasky/store"
	"github.com/samEscom/tasky/task"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	add := flag.Bool("add", false, "can add new task to do")
	complete := flag.Int("complete", 0, "mark a task as completed")
	doing := flag.Int("doing", 0, "mark a task as doing")
	deleted := flag.Int("delete", 0, "delete a task")
	list := flag.Bool("list", false, "list of all tasks")

	flag.Parse()

	if !*add && *complete == 0 && *doing == 0 && *deleted == 0 && !*list {
		*list = true
	}

	dataFile, err := resolveDataFile()
	if err != nil {
		return err
	}

	todos, err := store.Load(dataFile)
	if err != nil {
		return err
	}

	switch {
	case *add:
		taskText, err := getInput(os.Stdin, flag.Args()...)
		if err != nil {
			return err
		}

		todos.Add(taskText)
		return store.Save(dataFile, todos)
	case *complete > 0:
		if err := mutateAndSave(dataFile, &todos, func() error {
			return todos.Complete(*complete)
		}); err != nil {
			return err
		}
		return nil
	case *doing > 0:
		return mutateAndSave(dataFile, &todos, func() error {
			return todos.Doing(*doing)
		})
	case *deleted > 0:
		return mutateAndSave(dataFile, &todos, func() error {
			return todos.Delete(*deleted)
		})
	case *list:
		render.PrintTasks(todos)
	default:
		return errors.New("invalid command")
	}

	return nil
}

func mutateAndSave(filename string, todos *task.Task, mutate func() error) error {
	if err := mutate(); err != nil {
		return err
	}

	return store.Save(filename, *todos)
}

func resolveDataFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not find home directory: %w", err)
	}

	return filepath.Join(home, ".dataTodo.json"), nil
}

func getInput(r io.Reader, args ...string) (string, error) {
	if len(args) > 0 {
		taskText := strings.TrimSpace(strings.Join(args, " "))
		if taskText == "" {
			return "", errors.New("empty task, not allowed")
		}
		return taskText, nil
	}

	scanner := bufio.NewScanner(r)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", err
		}
		return "", errors.New("empty task, not allowed")
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	taskText := strings.TrimSpace(scanner.Text())
	if len(taskText) == 0 {
		return "", errors.New("empty task, not allowed")
	}

	return taskText, nil

}
