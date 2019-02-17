package controller

import (
	"net/http"

	"github.com/alexei-led/pumba/pkg/chaos"
	"github.com/alexei-led/pumba/pkg/chaos/docker"

	"github.com/gin-gonic/gin"
)

// PauseCommand REST API message
type PauseCommand struct {
	Random   bool     `json:"random,omitempty"`
	DryRun   bool     `json:"dry-run,omitempty"`
	Interval string   `json:"interval,omitempty"`
	Pattern  string   `json:"pattern,omitempty"`
	Names    []string `json:"names,omitempty"`
	Duration string   `json:"duration,omitempty"`
	Limit    int      `json:"limit,omitempty"`
}

// Pause handler
func (cmd *serverContext) Pause(c *gin.Context) {
	// pause command message
	var pause PauseCommand
	err := c.ShouldBindJSON(&pause)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// init kill command
	pauseCommand, err := docker.NewPauseCommand(chaos.DockerClient, pause.Names, pause.Pattern, pause.Interval, pause.Duration, pause.Limit, pause.DryRun)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// run pause command in goroutine
	go chaos.RunChaosCommand(cmd.context, pauseCommand, pause.Interval, pause.Random)
	c.JSON(http.StatusAccepted, gin.H{"status": "running pause command ..."})
	return
}
