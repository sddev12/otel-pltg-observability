package handlers

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func HandleSlow(c *gin.Context) {
	slog.InfoContext(c.Request.Context(), "Slow endpoint called")
	slowReqCounter.Add(c.Request.Context(), 1)

	time.Sleep(3 * time.Second)

	c.JSON(200, gin.H{
		"status": "slow endpoint successful",
	})
}
