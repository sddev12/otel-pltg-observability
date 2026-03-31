package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandleErrorGen(c *gin.Context) {
	slog.InfoContext(c.Request.Context(), "ErrorGen endpont called")
	healthzReqCounter.Add(c.Request.Context(), 1)

	c.JSON(http.StatusInternalServerError, gin.H{
		"status": "an unknown error ocurred",
	})
}
