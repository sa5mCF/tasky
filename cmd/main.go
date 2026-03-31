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

	"github.com/samEscom/tasky/task"
)

var (
	dataFile string
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not find home directory:", err)
		os.Exit(1)
	}
	dataFile = filepath.Join(home, ".dataTodo.json")
}

func main() {
	add := flag.Bool("add", false, "can add new task to do")
	complete := flag.Int("complete", 0, "mark a task as completed")
	doing := flag.Int("doing", 0, "mark a task as doing")
	deleted := flag.Int("delete", 0, "delete a task")
	list := flag.Bool("list", false, "list of all tasks")

	flag.Parse()

	todos := &task.Task{}

	if *add == false && *complete == 0 && *doing == 0 && *deleted == 0 && *list == false {
		*list = true
	}

	err := todos.Load(dataFile)

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	switch {
	case *add:
		task, err := getInput(os.Stdin, flag.Args()...)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		todos.Add(task)
		err = todos.Store(dataFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	case *complete > 0:
		err := todos.Complete(*complete)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		err = todos.Store(dataFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	case *doing > 0:
		err := todos.Doing(*doing)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		err = todos.Store(dataFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	case *deleted > 0:
		err := todos.Delete(*deleted)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		err = todos.Store(dataFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	case *list:
		todos.PrintTasks()
	default:
		fmt.Fprintln(os.Stdout, "invalid a command")
		os.Exit(0)
	}
}

func getInput(r io.Reader, args ...string) (string, error) {

	if len(args) > 0 {
		return strings.Join(args, " "), nil
	}

	scanner := bufio.NewScanner(r)
	scanner.Scan()

	if err := scanner.Err(); err != nil {
		return "", err
	}

	if len(scanner.Text()) == 0 {
		return "", errors.New("empty task, not allowed")
	}

	return scanner.Text(), nil

}
