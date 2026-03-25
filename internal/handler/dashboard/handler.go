package dashboard

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prastuvwxyz/pilot/internal/store"
	"github.com/prastuvwxyz/pilot/web/templates/pages"
)

type Handler struct {
	tasks *store.TaskStore
}

func NewHandler(tasks *store.TaskStore) *Handler {
	return &Handler{tasks: tasks}
}

func (h *Handler) Show(c *gin.Context) {
	board, err := h.tasks.AllTasks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	pages.Dashboard(board).Render(c.Request.Context(), c.Writer)
}
