package store_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/prastuvwxyz/pilot/internal/store"
)

func TestParseTaskCard(t *testing.T) {
	content := `---
id: TASK-001
title: Test Task
type: feature
priority: high
project: my-project
status: backlog
created: 2026-03-25
---

## Context
Test context.

## Acceptance Criteria
- [ ] Done
`
	card, err := store.ParseTaskCard("TASK-001", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if card.ID != "TASK-001" {
		t.Errorf("expected ID TASK-001, got %s", card.ID)
	}
	if card.Title != "Test Task" {
		t.Errorf("expected title 'Test Task', got %s", card.Title)
	}
	if card.Priority != "high" {
		t.Errorf("expected priority high, got %s", card.Priority)
	}
}

func TestListByColumn(t *testing.T) {
	// Setup temp dir
	dir := t.TempDir()
	backlog := filepath.Join(dir, "backlog")
	os.MkdirAll(backlog, 0755)

	// Write test task
	content := "---\nid: TASK-001\ntitle: Test\ntype: feature\npriority: high\nproject: test\nstatus: backlog\ncreated: 2026-03-25\n---\n\n## Context\ntest"
	os.WriteFile(filepath.Join(backlog, "TASK-001.md"), []byte(content), 0644)

	ts := store.NewTaskStore(dir)
	tasks, err := ts.ListByColumn("backlog")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Slug != "TASK-001" {
		t.Errorf("expected slug TASK-001, got %s", tasks[0].Slug)
	}
}
