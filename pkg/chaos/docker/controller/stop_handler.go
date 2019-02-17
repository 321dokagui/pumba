package controller

import (
	"net/http"

	"github.com/alexei-led/pumba/pkg/chaos"
	"github.com/alexei-led/pumba/pkg/chaos/docker"

	"github.com/gin-gonic/gin"
)

// Stop handler
func (cmd *serverContext) Stop(c *gin.Context) {
	// get stop message
	var msg docker.StopMessage
	err := c.ShouldBindJSON(&msg)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// init stop command
	stopCommand, err := docker.NewStopCommand(chaos.DockerClient, msg)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// run stop command in goroutine
	go chaos.RunChaosCommand(cmd.context, stopCommand, msg.Interval, msg.Random)
	c.JSON(http.StatusAccepted, gin.H{"status": "running stop command ..."})
	return
}
