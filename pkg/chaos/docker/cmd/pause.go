package cmd

import (
	"context"
	"fmt"

	"github.com/urfave/cli"

	"github.com/alexei-led/pumba/pkg/chaos"
	"github.com/alexei-led/pumba/pkg/chaos/docker"
)

type pauseContext struct {
	context context.Context
}

// NewPauseCLICommand initialize CLI pause command and bind it to the CommandContext
func NewPauseCLICommand(ctx context.Context) *cli.Command {
	cmdContext := &pauseContext{context: ctx}
	return &cli.Command{
		Name: "pause",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "duration, d",
				Usage: "pause duration: must be shorter than recurrent interval; use with optional unit suffix: 'ms/s/m/h'",
			},
			cli.IntFlag{
				Name:  "limit, l",
				Usage: "limit to number of container to kill (0: kill all matching)",
				Value: 0,
			},
		},
		Usage:       "pause all processes",
		ArgsUsage:   fmt.Sprintf("containers (name, list of names, or RE2 regex if prefixed with %q", chaos.Re2Prefix),
		Description: "pause all running processes within target containers",
		Action:      cmdContext.pause,
	}
}

// PAUSE Command
func (cmd *pauseContext) pause(c *cli.Context) error {
	message := docker.PauseMessage{}
	// get random flag
	message.Random = c.GlobalBool("random")
	// get dry-run mode
	message.DryRun = c.GlobalBool("dry-run")
	// get global chaos interval
	message.Interval = c.GlobalString("interval")
	// get limit for number of containers to pause
	message.Limit = c.Int("limit")
	// get names or pattern
	message.Names, message.Pattern = chaos.GetNamesOrPattern(c)
	// get chaos command duration
	message.Duration = c.String("duration")
	// init pause command
	pauseCommand, err := docker.NewPauseCommand(chaos.DockerClient, message)
	if err != nil {
		return err
	}
	// run pause command
	return chaos.RunChaosCommand(cmd.context, pauseCommand, message.Interval, message.Random)
}
