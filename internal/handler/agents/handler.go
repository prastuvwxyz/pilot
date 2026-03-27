package agents

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prastuvwxyz/pilot/internal/store"
	"github.com/prastuvwxyz/pilot/web/templates/pages"
)

type Handler struct {
	memoryPath string
}

func NewHandler(memoryPath string) *Handler {
	return &Handler{memoryPath: memoryPath}
}

func (h *Handler) List(c *gin.Context) {
	agents, _ := store.LoadAgents(h.memoryPath)
	pages.AgentsList(agents).Render(c.Request.Context(), c.Writer)
}

func (h *Handler) Detail(c *gin.Context) {
	id := c.Param("id")
	agent, err := store.GetAgent(h.memoryPath, id)
	if err != nil {
		c.Redirect(http.StatusFound, "/agents")
		return
	}
	files := store.AgentFiles(h.memoryPath, id)

	// Default: open first file
	var defaultFile, defaultContent string
	if len(files) > 0 {
		defaultFile = files[0]
		defaultContent, _ = store.ReadAgentFile(h.memoryPath, id, defaultFile)
	}

	pages.AgentDetail(agent, files, defaultFile, defaultContent).Render(c.Request.Context(), c.Writer)
}

func (h *Handler) GetFile(c *gin.Context) {
	id := c.Param("id")
	filename := c.Param("filename")
	content, err := store.ReadAgentFile(h.memoryPath, id, filename)
	if err != nil {
		c.String(http.StatusNotFound, "file not found")
		return
	}
	pages.AgentFileContent(id, filename, content).Render(c.Request.Context(), c.Writer)
}

func (h *Handler) SaveFile(c *gin.Context) {
	id := c.Param("id")
	filename := c.Param("filename")
	content := c.PostForm("content")
	if err := store.WriteAgentFile(h.memoryPath, id, filename, content); err != nil {
		c.String(http.StatusInternalServerError, "failed to save: "+err.Error())
		return
	}
	// Return updated content panel with a saved toast
	pages.AgentFileContent(id, filename, content).Render(c.Request.Context(), c.Writer)
}
