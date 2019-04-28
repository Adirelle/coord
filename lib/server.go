package lib

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type StateReader interface {
	Get(string) Status
	Wait(string, StatusPredicate, <-chan struct{}) bool
}

type StateWriter interface {
	Update(string, StatusUpdate)
	Remove(string)
}

func MakeServer(state interface{}, g *gin.RouterGroup) {
	if reader, ok := state.(StateReader); ok {
		rs := ReadServer{reader}
		g.GET("/*path", rs.GET)
	}

	if writer, ok := state.(StateWriter); ok {
		rw := WriteServer{writer}
		g.POST("/*path", rw.POST)
		g.DELETE("/*path", rw.DELETE)
	}
}

type StatusResponse struct {
	Status Status `json:"status"`
}

type WaitRequest struct {
	StatusPredicate `json:"wait"`
}

type ReadServer struct {
	state StateReader
}

func (s ReadServer) GET(c *gin.Context) {
	path := c.Param("path")

	var req WaitRequest
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !s.state.Wait(path, req, c.Done()) {
		c.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{"error": "time out"})
		return
	}

	if status := s.state.Get(path); status == UNDEFINED {
		c.Status(http.StatusNotFound)
	} else {
		c.JSON(http.StatusOK, StatusResponse{status})
	}
}

type WriteServer struct {
	state StateWriter
}

type UpdateRequest struct {
	StatusUpdate `json:"action"`
}

func (s WriteServer) POST(c *gin.Context) {
	path := c.Param("path")

	var req UpdateRequest
	if err := c.Bind(req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s.state.Update(path, req)
	c.Status(http.StatusNoContent)
}

func (s WriteServer) DELETE(c *gin.Context) {
	path := c.Param("path")
	s.state.Remove(path)
	c.Status(http.StatusNoContent)
}
