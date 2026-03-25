package store_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/prastuvwxyz/pilot/internal/store"
)

func TestGitStore_CommitAll(t *testing.T) {
	// Setup temp git repo
	dir := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %s", args, out)
		}
	}
	run("init")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")

	// Write a file
	os.WriteFile(filepath.Join(dir, "test.md"), []byte("hello"), 0644)

	gs := store.NewGitStore(dir)
	if err := gs.CommitAll("[pilot] test commit"); err != nil {
		t.Fatalf("CommitAll failed: %v", err)
	}

	// Verify commit exists
	cmd := exec.Command("git", "log", "--oneline")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git log failed: %v", err)
	}
	if len(out) == 0 {
		t.Error("expected at least one commit")
	}
}
