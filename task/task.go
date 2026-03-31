package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/alexeyco/simpletable"
)

type item struct {
	Task        string
	Doing       bool
	Done        bool
	CreatedAt   time.Time
	CompletedAt *time.Time
}

type Task []item

func (t *Task) Add(task string) {
	todo := item{
		Task:        task,
		Done:        false,
		CreatedAt:   time.Now(),
		CompletedAt: nil,
	}

	*t = append(*t, todo)
}

func (t *Task) Complete(index int) error {
	ls := *t

	if index <= 0 || index > len(ls) {
		return errors.New("invalid index")
	}

	now := time.Now()
	ls[index-1].CompletedAt = &now
	ls[index-1].Done = true

	return nil
}

func (t *Task) Delete(index int) error {
	ls := *t

	if index <= 0 || index > len(ls) {
		return errors.New("invalid index")
	}

	*t = append(ls[:index-1], ls[index:]...)

	return nil
}

func (t *Task) Doing(index int) error {
	ls := *t

	if index <= 0 || index > len(ls) {
		return errors.New("invalid index")
	}

	ls[index-1].Doing = true

	return nil
}

func (t *Task) Load(filename string) error {
	file, err := os.ReadFile(filename)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	if len(file) == 0 {
		return err
	}

	err = json.Unmarshal(file, t)

	if err != nil {
		return nil
	}

	return nil
}

func (t *Task) Store(filename string) error {

	data, err := json.Marshal(t)

	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

func (t *Task) PrintTasks() {
	table := simpletable.New()

	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Text: "#"},
			{Align: simpletable.AlignCenter, Text: "Task"},
			{Align: simpletable.AlignCenter, Text: "Done"},
			{Align: simpletable.AlignCenter, Text: "Doing"},
			{Align: simpletable.AlignCenter, Text: "CreatedAt"},
			{Align: simpletable.AlignCenter, Text: "CompletedAt"},
		},
	}

	var cells [][]*simpletable.Cell

	for i, item := range *t {
		i++
		task := blue(item.Task)

		if item.Done {
			task = green(fmt.Sprintf("\u2705 %s", item.Task))
		}

		completed := ""
		if item.CompletedAt != nil {
			completed = item.CompletedAt.Format(time.RFC822)
		}

		cells = append(cells, []*simpletable.Cell{
			{Text: fmt.Sprintf("%d", i)},
			{Text: task},
			{Text: fmt.Sprintf("%t", item.Done)},
			{Text: fmt.Sprintf("%t", item.Doing)},
			{Text: item.CreatedAt.Format(time.RFC822)},
			{Text: completed},
		})
	}

	table.Body = &simpletable.Body{Cells: cells}
	table.Footer = &simpletable.Footer{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Span: 6, Text: red(fmt.Sprintf("There are %d pendig tasks", t.Counter()))},
		},
	}

	table.SetStyle(simpletable.StyleUnicode)
	table.Println()
}

func (t *Task) Counter() int {
	total := 0

	for _, item := range *t {
		if !item.Done {
			total++
		}
	}

	return total
}
