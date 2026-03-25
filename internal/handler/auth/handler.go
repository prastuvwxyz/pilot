package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prastuvwxyz/pilot/internal/config"
	"github.com/prastuvwxyz/pilot/internal/handler/middleware"
	"github.com/prastuvwxyz/pilot/web/templates/pages"
)

type Handler struct {
	cfg *config.Config
}

func NewHandler(cfg *config.Config) *Handler {
	return &Handler{cfg: cfg}
}

func (h *Handler) ShowLogin(c *gin.Context) {
	// Already logged in → redirect to dashboard
	if _, err := c.Cookie("sid"); err == nil {
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}
	pages.Login("", "").Render(c.Request.Context(), c.Writer)
}

func (h *Handler) HandleLogin(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	if username != h.cfg.Pilot.Username || password != h.cfg.Pilot.Password {
		pages.Login("Invalid username or password", username).Render(c.Request.Context(), c.Writer)
		return
	}

	token, err := middleware.GenerateToken(username, h.cfg.Pilot.JWTSecret, h.cfg.Pilot.JWTExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.SetCookie("sid", token, int(h.cfg.Pilot.JWTExpiry.Seconds()), "/", "", false, true)
	c.Redirect(http.StatusFound, "/dashboard")
}

func (h *Handler) HandleLogout(c *gin.Context) {
	c.SetCookie("sid", "", -1, "/", "", false, true)
	c.Redirect(http.StatusFound, "/login")
}
