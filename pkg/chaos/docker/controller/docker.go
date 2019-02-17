package controller

import (
	"context"

	"github.com/gin-gonic/gin"
)

type serverContext struct {
	context context.Context
}

// DockerChaos controller interface
type DockerChaos interface {
	Kill(c *gin.Context)
	Pause(c *gin.Context)
	Remove(c *gin.Context)
	Stop(c *gin.Context)
}

// NewDockerChaosController dockerContext
func NewDockerChaosController(ctx context.Context) DockerChaos {
	return &serverContext{context: ctx}
}
