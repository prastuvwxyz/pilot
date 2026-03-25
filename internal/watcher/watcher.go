package watcher

import (
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Config struct {
	EngineeringTasksPath string
	LeadChannel          string // Discord channel ID for #engineering-lead
	DevChannel           string // Discord channel ID for #engineering-dev
}

type Watcher struct {
	cfg           Config
	fsw           *fsnotify.Watcher
	jiwooFlight   atomic.Bool // in-flight flag per watcher
	wasawhoFlight atomic.Bool
}

func New(cfg Config) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &Watcher{cfg: cfg, fsw: fsw}, nil
}

func (w *Watcher) Start() {
	backlogPath := filepath.Join(w.cfg.EngineeringTasksPath, "backlog")
	implQueuePath := filepath.Join(w.cfg.EngineeringTasksPath, "impl-queue")

	for _, path := range []string{backlogPath, implQueuePath} {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			slog.Warn("watcher: path does not exist, skipping", "path", path)
			continue
		}
		if err := w.fsw.Add(path); err != nil {
			slog.Error("watcher: failed to watch path", "path", path, "err", err)
		}
	}

	slog.Info("watcher: started", "backlog", backlogPath, "impl-queue", implQueuePath)

	var (
		jiwooTimer   *time.Timer
		wasawhoTimer *time.Timer
		mu           sync.Mutex
	)

	for {
		select {
		case event, ok := <-w.fsw.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Create == 0 {
				continue
			}

			mu.Lock()
			if filepath.Dir(event.Name) == backlogPath {
				if jiwooTimer != nil {
					jiwooTimer.Stop()
				}
				jiwooTimer = time.AfterFunc(2*time.Second, func() {
					if w.jiwooFlight.CompareAndSwap(false, true) {
						defer w.jiwooFlight.Store(false)
						w.triggerJiwoo()
					}
				})
			} else if filepath.Dir(event.Name) == implQueuePath {
				if wasawhoTimer != nil {
					wasawhoTimer.Stop()
				}
				wasawhoTimer = time.AfterFunc(2*time.Second, func() {
					if w.wasawhoFlight.CompareAndSwap(false, true) {
						defer w.wasawhoFlight.Store(false)
						w.triggerWasawho()
					}
				})
			}
			mu.Unlock()

		case err, ok := <-w.fsw.Errors:
			if !ok {
				return
			}
			slog.Error("watcher: fsnotify error", "err", err)
		}
	}
}

func (w *Watcher) Close() error {
	return w.fsw.Close()
}

func (w *Watcher) triggerJiwoo() {
	slog.Info("watcher: triggering Jiwoo (lead-engineer)")
	cmd := exec.Command("openclaw", "agent",
		"--agent", "lead-engineer",
		"--message", "Ada task baru di backlog.",
		"--deliver",
		"--reply-channel", "discord",
		"--reply-to", w.cfg.LeadChannel,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		slog.Error("watcher: failed to trigger Jiwoo", "err", err, "output", string(out))
	} else {
		slog.Info("watcher: Jiwoo triggered successfully")
	}
}

func (w *Watcher) triggerWasawho() {
	slog.Info("watcher: triggering Wasawho (software-engineer)")
	cmd := exec.Command("openclaw", "agent",
		"--agent", "software-engineer",
		"--message", "Ada subtask baru di impl-queue.",
		"--deliver",
		"--reply-channel", "discord",
		"--reply-to", w.cfg.DevChannel,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		slog.Error("watcher: failed to trigger Wasawho", "err", err, "output", string(out))
	} else {
		slog.Info("watcher: Wasawho triggered successfully")
	}
}
