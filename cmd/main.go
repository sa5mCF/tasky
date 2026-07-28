package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/samEscom/tasky/adapter/sqlite"
	"github.com/samEscom/tasky/application"
	"github.com/samEscom/tasky/render"
	"github.com/samEscom/tasky/tui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	add := flag.Bool("add", false, "can add new task to do")
	complete := flag.Int64("complete", 0, "mark a task as completed")
	doing := flag.Int64("doing", 0, "mark a task as doing")
	deleted := flag.Int64("delete", 0, "delete a task")
	list := flag.Bool("list", false, "list of all tasks")

	flag.Parse()

	hasCommand := *add || *complete > 0 || *doing > 0 || *deleted > 0 || *list

	databasePath, legacyJSONPath, err := resolveDataFiles()
	if err != nil {
		return err
	}

	ctx := context.Background()
	repository, err := sqlite.Open(ctx, databasePath, legacyJSONPath)
	if err != nil {
		return err
	}
	defer repository.Close()

	service := application.NewService(repository, nil)

	if !hasCommand && len(flag.Args()) == 0 {
		_, err := tui.Run(ctx, service)
		return err
	}

	switch {
	case *add:
		taskText, err := getInput(os.Stdin, flag.Args()...)
		if err != nil {
			return err
		}

		_, err = service.Add(ctx, taskText)
		return err
	case *complete > 0:
		return service.Complete(ctx, *complete)
	case *doing > 0:
		return service.MarkDoing(ctx, *doing)
	case *deleted > 0:
		return service.Delete(ctx, *deleted)
	case *list:
		todos, err := service.List(ctx)
		if err != nil {
			return err
		}
		render.PrintTasks(todos)
	default:
		return errors.New("invalid command")
	}

	return nil
}

func resolveDataFiles() (string, string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("could not find home directory: %w", err)
	}

	return filepath.Join(home, ".dataTodo.db"), filepath.Join(home, ".dataTodo.json"), nil
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
