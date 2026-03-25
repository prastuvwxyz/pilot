package store

import (
	"fmt"
	"os/exec"
	"strings"
)

type GitStore struct {
	repoPath string
}

func NewGitStore(repoPath string) *GitStore {
	return &GitStore{repoPath: repoPath}
}

func (g *GitStore) run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.repoPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s failed: %s", strings.Join(args, " "), string(out))
	}
	return string(out), nil
}

// PullRebase runs git pull --rebase. Returns error if rebase fails.
func (g *GitStore) PullRebase() error {
	_, err := g.run("pull", "--rebase")
	return err
}

// CommitAll stages all changes and commits with the given message.
func (g *GitStore) CommitAll(message string) error {
	if _, err := g.run("add", "-A"); err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}
	// Check if there's anything to commit
	status, err := g.run("status", "--porcelain")
	if err != nil {
		return err
	}
	if strings.TrimSpace(status) == "" {
		return nil // nothing to commit
	}
	if _, err := g.run("commit", "-m", message); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}
	return nil
}

// Push runs git push.
func (g *GitStore) Push() error {
	_, err := g.run("push")
	return err
}

// PullCommitPush is the standard Pilot git flow:
// pull --rebase → stage all → commit → push.
func (g *GitStore) PullCommitPush(message string) error {
	if err := g.PullRebase(); err != nil {
		return fmt.Errorf("pull --rebase failed (possible conflict): %w", err)
	}
	if err := g.CommitAll(message); err != nil {
		return err
	}
	return g.Push()
}
