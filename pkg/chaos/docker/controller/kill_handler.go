package controller

import (
	"net/http"

	"github.com/alexei-led/pumba/pkg/chaos"
	"github.com/alexei-led/pumba/pkg/chaos/docker"

	"github.com/gin-gonic/gin"
)

// Kill handler
func (cmd *serverContext) Kill(c *gin.Context) {
	// get REST API message
	var msg docker.KillMessage
	err := c.ShouldBindJSON(&msg)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// init kill command
	killCommand, err := docker.NewKillCommand(chaos.DockerClient, msg)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// run kill command in goroutine
	go chaos.RunChaosCommand(cmd.context, killCommand, msg.Interval, msg.Random)
	c.JSON(http.StatusAccepted, gin.H{"status": "running kill command ..."})
	return
}
