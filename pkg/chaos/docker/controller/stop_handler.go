package controller

import (
	"net/http"

	"github.com/alexei-led/pumba/pkg/chaos"
	"github.com/alexei-led/pumba/pkg/chaos/docker"

	"github.com/gin-gonic/gin"
)

// StopCommand REST API message
type StopCommand struct {
	Random   bool     `json:"random,omitempty"`
	DryRun   bool     `json:"dry-run,omitempty"`
	Interval string   `json:"interval,omitempty"`
	Pattern  string   `json:"pattern,omitempty"`
	Names    []string `json:"names,omitempty"`
	Restart  bool     `json:"restart,omitempty"`
	Duration string   `json:"duration,omitempty"`
	WaitTime int      `json:"wait-time,omitempty"`
	Limit    int      `json:"limit,omitempty"`
}

// Stop handler
func (cmd *serverContext) Stop(c *gin.Context) {
	// get stop message
	var stop StopCommand
	err := c.ShouldBindJSON(&stop)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// init kill command
	stopCommand, err := docker.NewStopCommand(chaos.DockerClient, stop.Names, stop.Pattern, stop.Restart, stop.Interval, stop.Duration, stop.WaitTime, stop.Limit, stop.DryRun)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// run stop command in goroutine
	go chaos.RunChaosCommand(cmd.context, stopCommand, stop.Interval, stop.Random)
	c.JSON(http.StatusAccepted, gin.H{"status": "running stop command ..."})
	return
}
