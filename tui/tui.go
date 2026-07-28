package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samEscom/tasky/application"
	"github.com/samEscom/tasky/task"
)

const (
	opList = iota
	opAdd
	opDoing
	opComplete
	opDelete
	opQuit
)

const panelGap = 2

type focus int

const (
	menuFocus focus = iota
	tasksFocus
)

type operation struct {
	label  string
	detail string
}

var operations = []operation{
	{label: "List tasks", detail: "Review your task list"},
	{label: "Add task", detail: "Create a new task"},
	{label: "Mark doing", detail: "Start a task"},
	{label: "Complete task", detail: "Mark a task as done"},
	{label: "Delete task", detail: "Remove a task"},
	{label: "Quit", detail: "Close Tasky"},
}

var (
	backgroundColor = lipgloss.Color("#15121d")
	panelColor      = lipgloss.Color("#211b2d")
	borderColor     = lipgloss.Color("#594472")
	lilacColor      = lipgloss.Color("#d0b5f5")
	mutedColor      = lipgloss.Color("#a495b7")
	selectedColor   = lipgloss.Color("#4c3565")
	doingColor      = lipgloss.Color("#e6c27a")
	greenColor      = lipgloss.Color("#9bd6b3")
	errorColor      = lipgloss.Color("#ff6b6b")
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lilacColor)
	panelStyle = lipgloss.NewStyle().
			Background(panelColor).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1, 1)
	selectedStyle = lipgloss.NewStyle().
			Background(selectedColor).
			Foreground(lilacColor).
			Bold(true)
	mutedStyle       = lipgloss.NewStyle().Foreground(mutedColor)
	doingStyle       = lipgloss.NewStyle().Foreground(doingColor).Bold(true)
	doneStyle        = lipgloss.NewStyle().Foreground(greenColor).Bold(true)
	activeValueStyle = lipgloss.NewStyle().Foreground(greenColor)
)

// Model contains the interactive task manager state.
type Model struct {
	ctx          context.Context
	service      *application.Service
	tasks        task.Task
	operation    int
	selectedTask int
	focus        focus
	input        textinput.Model
	editing      bool
	status       string
	err          error
	width        int
	height       int
}

// New creates a TUI model backed by the task application service.
func New(ctx context.Context, service *application.Service, todos task.Task) Model {
	input := textinput.New()
	input.Prompt = "Task: "
	input.Placeholder = "What needs to be done?"
	input.CharLimit = 200
	input.Width = 36

	return Model{
		ctx:     ctx,
		service: service,
		tasks:   todos,
		input:   input,
		width:   100,
		height:  30,
	}
}

// Run starts the interactive task manager and returns the final in-memory list.
func Run(ctx context.Context, service *application.Service) (task.Task, error) {
	todos, err := service.List(ctx)
	if err != nil {
		return nil, err
	}

	program := tea.NewProgram(New(ctx, service, todos), tea.WithAltScreen())
	finalModel, err := program.Run()
	if err != nil {
		return todos, err
	}

	model, ok := finalModel.(Model)
	if !ok {
		return todos, errors.New("could not read TUI state")
	}

	return model.tasks, nil
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		if m.editing {
			return m.updateInput(msg)
		}

		return m.updateKey(msg)
	}

	return m, nil
}

func (m Model) updateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.input.Blur()
		m.input.Reset()
		m.editing = false
		m.status = "Add cancelled"
		return m, nil
	case "enter":
		text := strings.TrimSpace(m.input.Value())
		if text == "" {
			m.setError(errors.New("empty task, not allowed"))
			return m, nil
		}

		created, err := m.service.Add(m.ctx, text)
		if err != nil {
			m.setError(err)
			return m, nil
		}

		m.input.Blur()
		m.input.Reset()
		m.editing = false
		m.tasks = append(m.tasks, created)
		if err := m.reloadTasks(); err != nil {
			m.setError(fmt.Errorf("task added but list refresh failed: %w", err))
			return m, nil
		}
		m.selectTask(created.ID)
		m.operation = opList
		m.focus = tasksFocus
		m.status = "Task added"
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) updateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.err != nil {
		m.err = nil
	}

	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "tab", "shift+tab":
		if m.focus == menuFocus {
			m.focus = tasksFocus
		} else {
			m.focus = menuFocus
		}
		return m, nil
	case "up", "k":
		m.moveSelection(-1)
		return m, nil
	case "down", "j":
		m.moveSelection(1)
		return m, nil
	case "enter":
		if m.focus == menuFocus {
			return m.activateOperation()
		}

		m.executeTaskOperation()
		return m, nil
	}

	return m, nil
}

func (m *Model) moveSelection(direction int) {
	if m.focus == menuFocus {
		m.operation = wrapIndex(m.operation+direction, len(operations))
		m.status = operations[m.operation].detail
		return
	}

	if len(m.tasks) == 0 {
		return
	}

	m.selectedTask = wrapIndex(m.selectedTask+direction, len(m.tasks))
}

func (m Model) activateOperation() (tea.Model, tea.Cmd) {
	switch m.operation {
	case opQuit:
		return m, tea.Quit
	case opAdd:
		m.input.Reset()
		cmd := m.input.Focus()
		m.editing = true
		m.status = "Type a task and press enter"
		return m, cmd
	case opList:
		m.focus = tasksFocus
		m.status = "Use the arrows to browse tasks"
		return m, nil
	default:
		m.focus = tasksFocus
		if len(m.tasks) == 0 {
			m.setError(errors.New("there are no tasks to update"))
			return m, nil
		}
		m.status = fmt.Sprintf("Select a task to %s", strings.ToLower(operations[m.operation].label))
		return m, nil
	}
}

func (m *Model) executeTaskOperation() {
	if m.operation == opList {
		m.status = "Select an operation from the left"
		return
	}

	if m.operation < opDoing || m.operation > opDelete || len(m.tasks) == 0 {
		return
	}

	id := m.tasks[m.selectedTask].ID
	var action func(context.Context, int64) error
	var status string
	switch m.operation {
	case opDoing:
		action = m.service.MarkDoing
		status = "Task marked as doing"
	case opComplete:
		action = m.service.Complete
		status = "Task completed"
	case opDelete:
		action = m.service.Delete
		status = "Task deleted"
	}

	if err := action(m.ctx, id); err != nil {
		m.setError(err)
		return
	}
	if err := m.reloadTasks(); err != nil {
		m.operation = opList
		m.setError(fmt.Errorf("task updated but list refresh failed: %w", err))
		return
	}

	if m.selectedTask >= len(m.tasks) && len(m.tasks) > 0 {
		m.selectedTask = len(m.tasks) - 1
	}
	m.operation = opList
	m.status = status
}

func (m *Model) reloadTasks() error {
	todos, err := m.service.List(m.ctx)
	if err != nil {
		return err
	}

	m.tasks = todos
	return nil
}

func (m *Model) selectTask(id int64) {
	for index, item := range m.tasks {
		if item.ID == id {
			m.selectedTask = index
			return
		}
	}
}

func (m *Model) setError(err error) {
	m.err = err
	m.status = ""
}

func (m Model) View() string {
	width := m.width
	if width < 70 {
		width = 70
	}

	leftWidth := 25
	rightWidth := width - leftWidth - panelGap - 8
	if rightWidth < 35 {
		rightWidth = 35
	}

	title := titleStyle.Render("TASKY") + "  " + mutedStyle.Render("local tasks, focused work")
	panels := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.menuView(leftWidth),
		strings.Repeat(" ", panelGap),
		m.taskView(rightWidth),
	)

	footer := mutedStyle.Render("tab switch  •  ↑/↓ move  •  enter select  •  q quit")
	if m.editing {
		footer = mutedStyle.Render("enter save  •  esc cancel  •  ctrl+c quit")
	}
	if m.status != "" {
		footer = activeValueStyle.Render(m.status) + "\n" + footer
	}
	if m.err != nil {
		footer = lipgloss.NewStyle().Foreground(errorColor).Render("Error: "+m.err.Error()) + "\n" + footer
	}

	return lipgloss.NewStyle().
		Background(backgroundColor).
		Width(width).
		Padding(1, 0).
		Render(title + "\n\n" + panels + "\n\n" + footer)
}

func (m Model) menuView(width int) string {
	lines := []string{titleStyle.Render("Operations"), ""}
	for index, item := range operations {
		line := "  " + item.label
		if index == m.operation {
			line = "▸ " + item.label
			if m.focus == menuFocus {
				line = selectedStyle.Width(width - 4).Render(line)
			} else {
				line = lipgloss.NewStyle().Foreground(lilacColor).Width(width - 4).Render(line)
			}
		} else {
			line = mutedStyle.Width(width - 4).Render(line)
		}
		lines = append(lines, line)
	}

	return panelStyle.Width(width).Render(strings.Join(lines, "\n"))
}

func (m Model) taskView(width int) string {
	lines := []string{
		titleStyle.Render(fmt.Sprintf("Tasks  ·  %d pending", m.tasks.Counter())),
		statusLegend(),
		mutedStyle.Render("Select a task when an operation needs one"),
		"",
	}

	if len(m.tasks) == 0 {
		lines = append(lines, mutedStyle.Render("No tasks yet. Choose Add task."))
	} else {
		for index, item := range m.tasks {
			state, style := taskState(item)

			prefix := fmt.Sprintf("%d  [%s]  ", item.ID, state)
			line := prefix + truncate(item.Task, width-lipgloss.Width(prefix)-4)
			if index == m.selectedTask {
				line = selectedStyle.Width(width - 4).Render(line)
			} else {
				line = style.Width(width - 4).Render(line)
			}
			lines = append(lines, line)
		}
	}

	if m.editing {
		lines = append(lines, "", titleStyle.Render("New task"), m.input.View())
	}

	return panelStyle.Width(width).Render(strings.Join(lines, "\n"))
}

func statusLegend() string {
	return mutedStyle.Render("Status: ") +
		mutedStyle.Render("[TODO]") + " pending  " +
		doingStyle.Render("[DOING]") + " active  " +
		doneStyle.Render("[DONE]") + " complete"
}

func taskState(item task.Item) (string, lipgloss.Style) {
	if item.Status == task.StatusDone {
		return "DONE", doneStyle
	}
	if item.Status == task.StatusDoing {
		return "DOING", doingStyle
	}
	return "TODO", mutedStyle
}

func wrapIndex(index, length int) int {
	if length == 0 {
		return 0
	}
	if index < 0 {
		return length - 1
	}
	if index >= length {
		return 0
	}
	return index
}

func truncate(value string, max int) string {
	if max < 4 {
		return value
	}

	runes := []rune(value)
	if len(runes) <= max {
		return value
	}

	return string(runes[:max-3]) + "..."
}
