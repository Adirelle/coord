package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type StateReader interface {
	Get(string) Status
	Wait(string, Status, time.Duration) bool
}

type StateWriter interface {
	Put(string, Status) <-chan bool
	Remove(string) <-chan bool
}

type StatusMsg struct {
	Status Status `json:"status"`
}

func MakeServer(state interface{}) *gin.Engine {
	g := gin.Default()

	if reader, ok := state.(StateReader); ok {
		g.GET("/*path", func(c *gin.Context) {
			path := c.Param("path")
			fmt.Printf("path=%#v\n", path)
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
				c.AbortWithError(http.StatusBadRequest, err)
			}
			fmt.Printf("path=%#v, msg=%#v\n", path, msg)
			writer.Put(path, msg.Status)
			c.Status(http.StatusNoContent)
		})
	}

	return g
}
