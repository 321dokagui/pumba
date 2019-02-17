package cmd

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/urfave/cli"

	"github.com/alexei-led/pumba/pkg/chaos/docker/controller"
)

type serverContext struct {
	context context.Context
}

// NewServerCommand initialize CLI server command and bind it to the serverContext
func NewServerCommand(ctx context.Context) *cli.Command {
	cmdContext := &serverContext{context: ctx}
	return &cli.Command{
		Name: "server",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:   "port, p",
				Usage:  "server port",
				Value:  8080,
				EnvVar: "PORT",
			},
		},
		Usage:       "run in a headless server mode",
		Description: "pumba server provides a REST API for running chaos commands",
		Action:      cmdContext.run,
	}
}

// Server Command
func (cmd *serverContext) run(c *cli.Context) error {
	// get dry-run mode
	// dryRun := c.GlobalBool("dry-run")
	// get server listener port
	port := c.Int("port")

	// configure server
	r := gin.Default()
	// report logs to stdout
	r.Use(gin.Logger())
	// Recovery middleware recovers from any panics and writes a 500 if there was one
	r.Use(gin.Recovery())

	dockerChaos := controller.NewDockerChaosController(cmd.context)

	r.POST("/docker/kill", dockerChaos.Kill)
	r.POST("/docker/pause", dockerChaos.Pause)
	r.POST("/docker/stop", dockerChaos.Stop)
	r.POST("/docker/remove", dockerChaos.Remove)

	// By default it serves on :8080 unless a
	// PORT environment variable was defined.
	return r.Run(fmt.Sprintf(":%v", port))
}
