package render

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/samEscom/tasky/task"
)

func PrintTasks(tasks task.Task) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "#\tTask\tDone\tDoing\tCreatedAt\tCompletedAt")

	for i, item := range tasks {
		taskText := blue(item.Task)

		if item.Doing {
			taskText = gray(item.Task)
		}

		if item.Done {
			taskText = green(fmt.Sprintf("\u2705 %s", item.Task))
		}

		completed := ""
		if item.CompletedAt != nil {
			completed = item.CompletedAt.Format(time.RFC822)
		}

		fmt.Fprintf(w, "%d\t%s\t%t\t%t\t%s\t%s\n",
			i+1,
			taskText,
			item.Done,
			item.Doing,
			item.CreatedAt.Format(time.RFC822),
			completed,
		)
	}

	fmt.Fprintln(w)
	fmt.Println(red(fmt.Sprintf("There are %d pending tasks", tasks.Counter())))
}
