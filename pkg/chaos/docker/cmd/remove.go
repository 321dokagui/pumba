package cmd

import (
	"context"
	"fmt"

	"github.com/urfave/cli"

	"github.com/alexei-led/pumba/pkg/chaos"
	"github.com/alexei-led/pumba/pkg/chaos/docker"
)

type removeContext struct {
	context context.Context
}

// NewRemoveCLICommand initialize CLI remove command and bind it to the remove4Context
func NewRemoveCLICommand(ctx context.Context) *cli.Command {
	cmdContext := &removeContext{context: ctx}
	return &cli.Command{
		Name: "rm",
		Flags: []cli.Flag{
			cli.BoolTFlag{
				Name:  "force, f",
				Usage: "force the removal of a running container (with SIGKILL)",
			},
			cli.BoolFlag{
				Name:  "links, n",
				Usage: "remove container links",
			},
			cli.BoolTFlag{
				Name:  "volumes, v",
				Usage: "remove volumes associated with the container",
			},
			cli.IntFlag{
				Name:  "limit, l",
				Usage: "limit to number of container to kill (0: kill all matching)",
				Value: 0,
			},
		},
		Usage:       "remove containers",
		ArgsUsage:   fmt.Sprintf("containers (name, list of names, or RE2 regex if prefixed with %q", chaos.Re2Prefix),
		Description: "remove target containers, with links and volumes",
		Action:      cmdContext.remove,
	}
}

// REMOVE Command
func (cmd *removeContext) remove(c *cli.Context) error {
	message := docker.RemoveMessage{}
	// get random
	message.Random = c.GlobalBool("random")
	// get dry-run mode
	message.DryRun = c.GlobalBool("dry-run")
	// get interval
	message.Interval = c.GlobalString("interval")
	// get names or pattern
	message.Names, message.Pattern = chaos.GetNamesOrPattern(c)
	// get force flag
	message.Force = c.BoolT("force")
	// get links flag
	message.Links = c.BoolT("links")
	// get volumes flag
	message.Volumes = c.BoolT("volumes")
	// get limit for number of containers to remove
	message.Limit = c.Int("limit")
	// init remove command
	removeCommand, err := docker.NewRemoveCommand(chaos.DockerClient, message)
	if err != nil {
		return err
	}
	// run remove command
	return chaos.RunChaosCommand(cmd.context, removeCommand, message.Interval, message.Random)
}
