package dashboard

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prastuvwxyz/pilot/internal/store"
	"github.com/prastuvwxyz/pilot/web/templates/pages"
)

type Handler struct {
	tasks         *store.TaskStore
	prasMemoryPath string
}

func NewHandler(tasks *store.TaskStore, prasMemoryPath string) *Handler {
	return &Handler{tasks: tasks, prasMemoryPath: prasMemoryPath}
}

func (h *Handler) Show(c *gin.Context) {
	board, err := h.tasks.AllTasks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	agents, _ := store.LoadAgents(h.prasMemoryPath)
	pages.Dashboard(board, agents).Render(c.Request.Context(), c.Writer)
}
