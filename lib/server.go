package lib

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type StateReader interface {
	Get(string) Status
	Wait(string, Status, time.Duration) bool
}

type StateWriter interface {
	Put(string, Status)
	Remove(string)
}

type StatusMsg struct {
	Status Status `json:"status"`
}

func MakeServer(state interface{}, g *gin.RouterGroup) {
	if reader, ok := state.(StateReader); ok {
		g.GET("/*path", func(c *gin.Context) {
			path := c.Param("path")

			if wait := c.Query("wait"); wait != "" {
				var expected Status
				if err := expected.UnmarshalText([]byte(wait)); err != nil {
					c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				if !reader.Wait(path, expected, 5*time.Minute) {
					c.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{"error": "time out"})
					return
				}
			}

			if status := reader.Get(path); status == UNDEFINED {
				c.Status(http.StatusNotFound)
			} else {
				c.JSON(http.StatusOK, StatusMsg{status})
			}
		})
	}

	if writer, ok := state.(StateWriter); ok {
		g.PUT("/*path", func(c *gin.Context) {
			path := c.Param("path")
			var msg StatusMsg
			if err := c.ShouldBindJSON(&msg); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			writer.Put(path, msg.Status)
			c.Status(http.StatusNoContent)
		})

		g.DELETE("/*path", func(c *gin.Context) {
			path := c.Param("path")
			writer.Remove(path)
			c.Status(http.StatusNoContent)
		})
	}
}
