package controller

import (
	"net/http"

	"github.com/alexei-led/pumba/pkg/chaos"
	"github.com/alexei-led/pumba/pkg/chaos/docker"

	"github.com/gin-gonic/gin"
)

// RemoveCommand REST API message
type RemoveCommand struct {
	Random   bool     `json:"random,omitempty"`
	DryRun   bool     `json:"dry-run,omitempty"`
	Interval string   `json:"interval,omitempty"`
	Pattern  string   `json:"pattern,omitempty"`
	Names    []string `json:"names,omitempty"`
	Force    bool     `json:"force,omitempty"`
	Volumes  bool     `json:"volumes,omitempty"`
	Links    bool     `json:"links,omitempty"`
	Limit    int      `json:"limit,omitempty"`
}

// Remove handler
func (cmd *serverContext) Remove(c *gin.Context) {
	// get names or pattern
	var remove RemoveCommand
	err := c.ShouldBindJSON(&remove)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// init remove command
	removeCommand, err := docker.NewRemoveCommand(chaos.DockerClient, remove.Names, remove.Pattern, remove.Force, remove.Links, remove.Volumes, remove.Limit, remove.DryRun)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// run remove command in goroutine
	go chaos.RunChaosCommand(cmd.context, removeCommand, remove.Interval, remove.Random)
	c.JSON(http.StatusAccepted, gin.H{"status": "running remove command..."})
	return
}
