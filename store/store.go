package store

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/samEscom/tasky/task"
)

func Load(filename string) (task.Task, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return task.Task{}, nil
		}
		return nil, err
	}

	if len(data) == 0 {
		return task.Task{}, nil
	}

	var todos task.Task
	if err := json.Unmarshal(data, &todos); err != nil {
		return nil, err
	}

	return todos, nil
}

func Save(filename string, todos task.Task) error {
	data, err := json.Marshal(todos)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}
