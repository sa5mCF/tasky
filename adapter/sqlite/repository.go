package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/samEscom/tasky/task"
)

const legacyImportKey = "legacy_json_v1"

const createTasksTable = `
CREATE TABLE IF NOT EXISTS tasks (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	description TEXT NOT NULL,
	status TEXT NOT NULL CHECK (status IN ('todo', 'doing', 'done')),
	created_at TEXT NOT NULL,
	completed_at TEXT NULL
)`

const createMetadataTable = `
CREATE TABLE IF NOT EXISTS app_metadata (
	key TEXT PRIMARY KEY,
	value TEXT NOT NULL
)`

type Repository struct {
	db *sql.DB
}

var _ task.Repository = (*Repository)(nil)

type legacyItem struct {
	Task        string
	Doing       bool
	Done        bool
	CreatedAt   time.Time
	CompletedAt *time.Time
}

type rowScanner interface {
	Scan(...any) error
}

func Open(ctx context.Context, databasePath, legacyJSONPath string) (*Repository, error) {
	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	// Tasky is a local single-user CLI. One connection also keeps connection-level
	// SQLite settings consistent for the lifetime of the process.
	db.SetMaxOpenConns(1)

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("connect to sqlite database: %w", err)
	}

	if _, err := db.ExecContext(ctx, "PRAGMA busy_timeout = 5000"); err != nil {
		db.Close()
		return nil, fmt.Errorf("configure sqlite database: %w", err)
	}

	repository := &Repository{db: db}
	if err := repository.initialize(ctx, legacyJSONPath); err != nil {
		db.Close()
		return nil, err
	}

	return repository, nil
}

func (r *Repository) Close() error {
	return r.db.Close()
}

func (r *Repository) List(ctx context.Context) (task.Task, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, description, status, created_at, completed_at
		FROM tasks
		ORDER BY id ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	tasks := task.Task{}
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read tasks: %w", err)
	}

	return tasks, nil
}

func (r *Repository) FindByID(ctx context.Context, id int64) (task.Item, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, description, status, created_at, completed_at
		FROM tasks
		WHERE id = ?
	`, id)

	item, err := scanItem(row)
	if errors.Is(err, sql.ErrNoRows) {
		return task.Item{}, task.ErrNotFound
	}
	if err != nil {
		return task.Item{}, err
	}

	return item, nil
}

func (r *Repository) Create(ctx context.Context, item task.Item) (task.Item, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO tasks (description, status, created_at, completed_at)
		VALUES (?, ?, ?, ?)
	`, item.Task, item.Status, formatTime(item.CreatedAt), formatOptionalTime(item.CompletedAt))
	if err != nil {
		return task.Item{}, fmt.Errorf("create task: %w", err)
	}

	item.ID, err = result.LastInsertId()
	if err != nil {
		return task.Item{}, fmt.Errorf("read created task id: %w", err)
	}

	return item, nil
}

func (r *Repository) Update(ctx context.Context, item task.Item) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE tasks
		SET description = ?, status = ?, created_at = ?, completed_at = ?
		WHERE id = ?
	`, item.Task, item.Status, formatTime(item.CreatedAt), formatOptionalTime(item.CompletedAt), item.ID)
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	return ensureAffected(result)
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	return ensureAffected(result)
}

func (r *Repository) initialize(ctx context.Context, legacyJSONPath string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin database initialization: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, createTasksTable); err != nil {
		return fmt.Errorf("create tasks table: %w", err)
	}
	if _, err := tx.ExecContext(ctx, createMetadataTable); err != nil {
		return fmt.Errorf("create metadata table: %w", err)
	}

	var marker string
	err = tx.QueryRowContext(
		ctx,
		"SELECT value FROM app_metadata WHERE key = ?",
		legacyImportKey,
	).Scan(&marker)
	switch {
	case err == nil:
		return commit(tx)
	case !errors.Is(err, sql.ErrNoRows):
		return fmt.Errorf("read legacy import marker: %w", err)
	}

	var taskCount int
	if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks").Scan(&taskCount); err != nil {
		return fmt.Errorf("count existing tasks: %w", err)
	}

	if taskCount == 0 {
		if err := importLegacyJSON(ctx, tx, legacyJSONPath); err != nil {
			return err
		}
	}

	if _, err := tx.ExecContext(
		ctx,
		"INSERT INTO app_metadata (key, value) VALUES (?, ?)",
		legacyImportKey,
		"complete",
	); err != nil {
		return fmt.Errorf("record legacy import: %w", err)
	}

	return commit(tx)
}

func importLegacyJSON(ctx context.Context, tx *sql.Tx, legacyJSONPath string) error {
	data, err := os.ReadFile(legacyJSONPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read legacy task file %q: %w", legacyJSONPath, err)
	}

	if strings.TrimSpace(string(data)) == "" {
		return nil
	}

	var legacyTasks []legacyItem
	if err := json.Unmarshal(data, &legacyTasks); err != nil {
		return fmt.Errorf("decode legacy task file %q: %w", legacyJSONPath, err)
	}

	for _, legacy := range legacyTasks {
		status := task.StatusTodo
		completedAt := (*time.Time)(nil)
		switch {
		case legacy.Done:
			status = task.StatusDone
			completedAt = legacy.CompletedAt
		case legacy.Doing:
			status = task.StatusDoing
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO tasks (description, status, created_at, completed_at)
			VALUES (?, ?, ?, ?)
		`, legacy.Task, status, formatTime(legacy.CreatedAt), formatOptionalTime(completedAt)); err != nil {
			return fmt.Errorf("import legacy task: %w", err)
		}
	}

	return nil
}

func scanItem(scanner rowScanner) (task.Item, error) {
	var (
		item        task.Item
		status      string
		createdAt   string
		completedAt sql.NullString
	)

	if err := scanner.Scan(&item.ID, &item.Task, &status, &createdAt, &completedAt); err != nil {
		return task.Item{}, err
	}

	item.Status = task.Status(status)
	switch item.Status {
	case task.StatusTodo, task.StatusDoing, task.StatusDone:
	default:
		return task.Item{}, fmt.Errorf("read task %d: invalid status %q", item.ID, status)
	}

	var err error
	item.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return task.Item{}, fmt.Errorf("read task %d created_at: %w", item.ID, err)
	}

	if completedAt.Valid {
		value, err := time.Parse(time.RFC3339Nano, completedAt.String)
		if err != nil {
			return task.Item{}, fmt.Errorf("read task %d completed_at: %w", item.ID, err)
		}
		item.CompletedAt = &value
	}

	return item, nil
}

func formatTime(value time.Time) string {
	return value.Format(time.RFC3339Nano)
}

func formatOptionalTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return formatTime(*value)
}

func ensureAffected(result sql.Result) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read affected rows: %w", err)
	}
	if affected == 0 {
		return task.ErrNotFound
	}
	return nil
}

func commit(tx *sql.Tx) error {
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit database initialization: %w", err)
	}
	return nil
}
