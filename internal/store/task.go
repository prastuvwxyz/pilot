package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Columns in order for kanban display
var Columns = []string{
	"backlog",
	"lead-queue",
	"lead-review",
	"impl-queue",
	"in-progress",
	"blocked",
	"done",
}

type TaskCard struct {
	// Frontmatter fields
	ID         string   `yaml:"id"`
	Title      string   `yaml:"title"`
	Type       string   `yaml:"type"`
	Priority   string   `yaml:"priority"`
	Project    string   `yaml:"project"`
	AssignedTo string   `yaml:"assigned_to"`
	DependsOn  []string `yaml:"depends_on"`
	Created    string   `yaml:"created"`
	Due        string   `yaml:"due"`

	// Computed fields (not from frontmatter)
	Slug   string `yaml:"-"` // folder/file name without extension
	Column string `yaml:"-"` // current column
	Body   string `yaml:"-"` // markdown body after frontmatter
}

type TaskStore struct {
	basePath string
}

func NewTaskStore(basePath string) *TaskStore {
	return &TaskStore{basePath: basePath}
}

// ParseTaskCard parses a markdown file with YAML frontmatter.
// slug is the filename without extension (e.g. "TASK-001").
func ParseTaskCard(slug, content string) (*TaskCard, error) {
	if !strings.HasPrefix(content, "---") {
		return nil, fmt.Errorf("no frontmatter found in task card %s", slug)
	}

	// Find closing ---
	rest := content[3:] // skip opening ---
	idx := strings.Index(rest, "\n---")
	if idx == -1 {
		return nil, fmt.Errorf("frontmatter not closed in task card %s", slug)
	}

	frontmatter := rest[:idx]
	body := strings.TrimSpace(rest[idx+4:]) // skip closing ---\n

	var card TaskCard
	if err := yaml.Unmarshal([]byte(frontmatter), &card); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter in %s: %w", slug, err)
	}

	card.Slug = slug
	card.Body = body
	return &card, nil
}

// ListByColumn returns all task cards in a given column.
// Handles both flat files (backlog/TASK-001.md) and
// folders (lead-review/TASK-001/TASK-001.md or any .md in the folder root).
func (ts *TaskStore) ListByColumn(column string) ([]TaskCard, error) {
	colPath := filepath.Join(ts.basePath, column)
	entries, err := os.ReadDir(colPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read column %s: %w", column, err)
	}

	var tasks []TaskCard
	for _, entry := range entries {
		if entry.Name() == ".gitkeep" {
			continue
		}

		var card *TaskCard
		var slug string

		if entry.IsDir() {
			// Folder-based task (lead-review, impl-queue, etc.)
			slug = entry.Name()
			card, err = ts.readTaskFromFolder(column, slug)
			if err != nil {
				continue // skip if can't read
			}
		} else if strings.HasSuffix(entry.Name(), ".md") {
			// Flat file task (backlog, lead-queue)
			slug = strings.TrimSuffix(entry.Name(), ".md")
			card, err = ts.readTaskFromFile(column, slug)
			if err != nil {
				continue
			}
		} else {
			continue
		}

		if card != nil {
			card.Slug = slug
			card.Column = column
			tasks = append(tasks, *card)
		}
	}
	return tasks, nil
}

func (ts *TaskStore) readTaskFromFile(column, slug string) (*TaskCard, error) {
	path := filepath.Join(ts.basePath, column, slug+".md")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseTaskCard(slug, string(data))
}

func (ts *TaskStore) readTaskFromFolder(column, slug string) (*TaskCard, error) {
	folderPath := filepath.Join(ts.basePath, column, slug)
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		name := e.Name()
		if name == "rfc.md" || name == "RFC.md" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(folderPath, name))
		if err != nil {
			continue
		}
		return ParseTaskCard(slug, string(data))
	}
	return nil, fmt.Errorf("no task card found in folder %s/%s", column, slug)
}

// AllTasks returns tasks across all columns as a map[column][]TaskCard.
func (ts *TaskStore) AllTasks() (map[string][]TaskCard, error) {
	result := make(map[string][]TaskCard)
	for _, col := range Columns {
		tasks, err := ts.ListByColumn(col)
		if err != nil {
			return nil, err
		}
		result[col] = tasks
	}
	return result, nil
}

// CreateTask writes a new task card to backlog/.
func (ts *TaskStore) CreateTask(card TaskCard) error {
	if card.Slug == "" {
		return fmt.Errorf("task slug is required")
	}
	path := filepath.Join(ts.basePath, "backlog", card.Slug+".md")

	frontmatter, err := yaml.Marshal(card)
	if err != nil {
		return fmt.Errorf("failed to marshal frontmatter: %w", err)
	}

	content := fmt.Sprintf("---\n%s---\n\n%s", string(frontmatter), card.Body)
	return os.WriteFile(path, []byte(content), 0644)
}

// ApproveTask copies subtask files from lead-review/{slug}/subtasks/
// to impl-queue/{slug}/subtasks/. RFC stays in lead-review.
func (ts *TaskStore) ApproveTask(slug string) error {
	srcSubtasks := filepath.Join(ts.basePath, "lead-review", slug, "subtasks")
	dstSubtasks := filepath.Join(ts.basePath, "impl-queue", slug, "subtasks")

	if _, err := os.Stat(srcSubtasks); os.IsNotExist(err) {
		return fmt.Errorf("lead-review/%s/subtasks not found", slug)
	}

	if err := os.MkdirAll(dstSubtasks, 0755); err != nil {
		return fmt.Errorf("failed to create impl-queue subtasks dir: %w", err)
	}

	entries, err := os.ReadDir(srcSubtasks)
	if err != nil {
		return fmt.Errorf("failed to read subtasks: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		src := filepath.Join(srcSubtasks, entry.Name())
		dst := filepath.Join(dstSubtasks, entry.Name())
		data, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("failed to read subtask %s: %w", entry.Name(), err)
		}
		if err := os.WriteFile(dst, data, 0644); err != nil {
			return fmt.Errorf("failed to write subtask %s: %w", entry.Name(), err)
		}
	}
	return nil
}

// MoveTask moves a task from one column to another.
// Handles both file-based and folder-based tasks.
func (ts *TaskStore) MoveTask(slug, fromColumn, toColumn string) error {
	// Try file first
	srcFile := filepath.Join(ts.basePath, fromColumn, slug+".md")
	dstFile := filepath.Join(ts.basePath, toColumn, slug+".md")
	if _, err := os.Stat(srcFile); err == nil {
		return os.Rename(srcFile, dstFile)
	}

	// Try folder
	srcDir := filepath.Join(ts.basePath, fromColumn, slug)
	dstDir := filepath.Join(ts.basePath, toColumn, slug)
	if _, err := os.Stat(srcDir); err == nil {
		if err := os.MkdirAll(filepath.Join(ts.basePath, toColumn), 0755); err != nil {
			return err
		}
		return os.Rename(srcDir, dstDir)
	}

	return fmt.Errorf("task %s not found in column %s", slug, fromColumn)
}
