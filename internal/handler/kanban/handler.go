package kanban

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prastuvwxyz/pilot/internal/store"
	"github.com/prastuvwxyz/pilot/web/templates/pages"
)

type Handler struct {
	tasks *store.TaskStore
	git   *store.GitStore
}

func NewHandler(tasks *store.TaskStore, git *store.GitStore) *Handler {
	return &Handler{tasks: tasks, git: git}
}

func (h *Handler) ShowKanban(c *gin.Context) {
	board, err := h.tasks.AllTasks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	pages.Kanban(board).Render(c.Request.Context(), c.Writer)
}

func (h *Handler) CreateTask(c *gin.Context) {
	slug := c.PostForm("slug")
	title := c.PostForm("title")
	project := c.PostForm("project")
	taskType := c.PostForm("type")
	priority := c.PostForm("priority")
	assignedTo := c.PostForm("assigned_to")
	due := c.PostForm("due")
	context := c.PostForm("context")

	if slug == "" || title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug and title are required"})
		return
	}

	body := ""
	if context != "" {
		body = fmt.Sprintf("## Context\n%s\n\n## Acceptance Criteria\n- [ ] TBD\n\n## Notes\n~", context)
	}

	card := store.TaskCard{
		ID:         slug,
		Slug:       slug,
		Title:      title,
		Type:       taskType,
		Priority:   priority,
		Project:    project,
		AssignedTo: assignedTo,
		Due:        due,
		Created:    time.Now().Format("2006-01-02"),
		Body:       body,
	}

	if err := h.tasks.CreateTask(card); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.git.PullCommitPush(fmt.Sprintf("[pilot] add task: %s", slug)); err != nil {
		// Log but don't fail — task file is written, git is best-effort
		c.Header("X-Git-Error", err.Error())
	}

	c.Redirect(http.StatusFound, "/kanban")
}

func (h *Handler) ApproveTask(c *gin.Context) {
	slug := c.Param("id")

	if err := h.tasks.ApproveTask(slug); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.git.PullCommitPush(fmt.Sprintf("[pilot] approved: %s → impl-queue", slug)); err != nil {
		c.Header("X-Git-Error", err.Error())
	}

	// Return empty card (HTMX replaces card with nothing after approve)
	c.Status(http.StatusOK)
}

func (h *Handler) DeleteTask(c *gin.Context) {
	slug := c.Param("id")

	if err := h.tasks.DeleteTask(slug); err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.git.PullCommitPush(fmt.Sprintf("[pilot] delete task: %s", slug)); err != nil {
		c.Header("X-Git-Error", err.Error())
	}

	c.Status(http.StatusOK)
}

func (h *Handler) MoveTask(c *gin.Context) {
	slug := c.Param("id")

	var req struct {
		From string `json:"from" binding:"required"`
		To   string `json:"to" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.tasks.MoveTask(slug, req.From, req.To); err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.git.PullCommitPush(fmt.Sprintf("[pilot] move: %s → %s", slug, req.To)); err != nil {
		c.Header("X-Git-Error", err.Error())
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
