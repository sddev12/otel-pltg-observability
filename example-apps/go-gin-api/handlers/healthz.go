package handlers

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

func HandleHealthz(c *gin.Context) {
	slog.InfoContext(c.Request.Context(), "Healthcheck endpoint called")
	healthzReqCounter.Add(c.Request.Context(), 1)

	c.JSON(200, gin.H{
		"status": "ok",
	})
}
