package lib

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type StateReader interface {
	Get(string) Status
	Wait(string, Status, <-chan struct{}) bool
}

type StateWriter interface {
	Put(string, Status)
	Remove(string)
}

func MakeServer(state interface{}, g *gin.RouterGroup) {
	if reader, ok := state.(StateReader); ok {
		rs := ReadServer{reader}
		g.GET("/*path", rs.GET)
	}

	if writer, ok := state.(StateWriter); ok {
		rw := WriteServer{writer}
		g.PUT("/*path", rw.PUT)
		g.DELETE("/*path", rw.DELETE)
	}
}

type StatusMsg struct {
	Status Status `json:"status"`
}

type ReadServer struct {
	state StateReader
}

func (s ReadServer) GET(c *gin.Context) {
	path := c.Param("path")

	if waitParam := c.Query("wait"); waitParam != "" {
		wait := UNDEFINED
		if err := wait.UnmarshalText([]byte(waitParam)); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if wait != UNDEFINED {
			if !s.state.Wait(path, wait, c.Done()) {
				c.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{"error": "time out"})
				return
			}
		}
	}

	if status := s.state.Get(path); status == UNDEFINED {
		c.Status(http.StatusNotFound)
	} else {
		c.JSON(http.StatusOK, StatusMsg{status})
	}
}

type WriteServer struct {
	state StateWriter
}

func (s WriteServer) PUT(c *gin.Context) {
	path := c.Param("path")
	var msg StatusMsg
	if err := c.ShouldBindJSON(&msg); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s.state.Put(path, msg.Status)
	c.Status(http.StatusNoContent)
}

func (s WriteServer) DELETE(c *gin.Context) {
	path := c.Param("path")
	s.state.Remove(path)
	c.Status(http.StatusNoContent)
}
