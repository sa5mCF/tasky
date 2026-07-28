package render

import (
	"fmt"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/samEscom/tasky/task"
)

func PrintTasks(tasks task.Task) {
	tw := table.NewWriter()
	tw.SetStyle(table.StyleColoredDark)
	tw.AppendHeader(table.Row{"#", "Task", "Done", "Doing", "CreatedAt", "CompletedAt"})

	for i, item := range tasks {
		completed := ""
		if item.CompletedAt != nil {
			completed = item.CompletedAt.Format(time.RFC822)
		}

		taskText := item.Task
		if item.Done {
			taskText = fmt.Sprintf("\u2705 %s", item.Task)
		}

		tw.AppendRow(table.Row{
			i + 1,
			taskText,
			item.Done,
			item.Doing,
			item.CreatedAt.Format(time.RFC822),
			completed,
		})
	}

	tw.AppendFooter(table.Row{"", "", "", "", "", fmt.Sprintf("There are %d pending tasks", tasks.Counter())})
	fmt.Println(tw.Render())
}
