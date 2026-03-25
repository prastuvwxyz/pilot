# Pilot

Engineering task board for AI agent automation. A Go web app that manages a kanban board backed by markdown files and auto-triggers AI agents when tasks move through the pipeline.

## What it does

- **Kanban UI** — view and manage tasks across columns (backlog → lead-queue → lead-review → impl-queue → in-progress → blocked → done)
- **File watcher** — detects new files in `backlog/` and `impl-queue/`, auto-triggers AI agents via `openclaw` CLI
- **Task creation** — creates markdown task cards with YAML frontmatter directly in the filesystem
- **RFC approval** — one-click approve moves subtasks from `lead-review/` to `impl-queue/`
- **Git sync** — every write ops: `git pull --rebase → commit → push`
- **JWT auth** — simple username/password login, no database

## Stack

- **Go 1.24+** — single binary
- **Gin** — HTTP router
- **a-h/templ** — type-safe HTML templates
- **HTMX + Alpine.js** — interactive UI without a JS framework
- **TailwindCSS + DaisyUI** — styling
- **fsnotify** — file system watcher
- **Viper** — config + env binding
- **golang-jwt** — JWT auth

## Architecture

```
pilot/
├── cmd/web/main.go              ← server entry, wires everything
├── internal/
│   ├── config/config.go         ← Viper config loader
│   ├── handler/
│   │   ├── auth/                ← login/logout
│   │   ├── dashboard/           ← summary stats
│   │   ├── kanban/              ← task CRUD + approve + move
│   │   ├── health/              ← /healthz
│   │   └── middleware/jwt.go    ← JWT auth middleware
│   ├── store/
│   │   ├── task.go              ← read/write markdown files
│   │   └── git.go              ← pull/commit/push via os/exec
│   └── watcher/watcher.go      ← fsnotify + debounce + agent trigger
└── web/templates/               ← Templ components and pages
```

Two goroutines: Gin HTTP server + fsnotify file watcher.

No database — reads/writes markdown files directly.

## Setup

**Requirements:** Go 1.24+, `templ`, `air` (dev), `tailwindcss`

```bash
# Install dev tools
go install github.com/a-h/templ/cmd/templ@latest
go install github.com/air-verse/air@latest
brew install tailwindcss

# Clone and configure
git clone https://github.com/prastuvwxyz/pilot.git
cd pilot
cp .env.example .env
# Edit .env with your values
```

**Required env vars:**
```bash
PILOT_USERNAME=yourname
PILOT_PASSWORD=yourpassword
PILOT_JWT_SECRET=min-32-chars-secret
ENGINEERING_TASKS_PATH=/path/to/your/engineering-tasks
PRAS_MEMORY_PATH=/path/to/your/pras-memory
```

`ENGINEERING_TASKS_PATH` must exist and contain the kanban column subdirectories (`backlog/`, `lead-queue/`, etc).

## Development

```bash
make dev        # hot reload with air
make css-watch  # watch TailwindCSS in another terminal
```

## Build & Deploy

```bash
# Build local
make build

# Build for Linux server
GOOS=linux GOARCH=amd64 make build

# Deploy
make deploy     # scp + systemctl restart
```

See `deploy/` for systemd service and Nginx config.

## Task format

Tasks are markdown files with YAML frontmatter:

```markdown
---
id: TASK-001
title: Add user authentication
type: feature
priority: high
project: prastya.com
status: backlog
created: 2026-03-25
---

## Context
Why this task exists.

## Acceptance Criteria
- [ ] Done
```

## Agent integration

Pilot watches for `CREATE` events via fsnotify with a 2-second debounce:

- New file in `backlog/` → triggers `openclaw agent --agent lead-engineer`
- New file in `impl-queue/` → triggers `openclaw agent --agent software-engineer`

Replace the `openclaw` CLI calls in `internal/watcher/watcher.go` with whatever agent runner you use.

## License

MIT
