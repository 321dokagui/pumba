package docker

import (
	"context"

	"github.com/alexei-led/pumba/pkg/chaos"
	"github.com/alexei-led/pumba/pkg/container"
	log "github.com/sirupsen/logrus"
)

// RemoveMessage REST API message
type RemoveMessage struct {
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

// RemoveCommand `docker kill` command
type RemoveCommand struct {
	RemoveMessage
	client container.Client
}

// NewRemoveCommand create new Kill Command instance
func NewRemoveCommand(client container.Client, msg RemoveMessage) (chaos.Command, error) {
	return &RemoveCommand{msg, client}, nil
}

// Run remove command
func (r *RemoveCommand) Run(ctx context.Context, random bool) error {
	log.Debug("removing all matching containers")
	log.WithFields(log.Fields{
		"names":   r.Names,
		"pattern": r.Pattern,
		"limit":   r.Limit,
	}).Debug("listing matching containers")
	containers, err := container.ListNContainers(ctx, r.client, r.Names, r.Pattern, r.Limit)
	if err != nil {
		log.WithError(err).Error("failed to list containers")
		return err
	}
	if len(containers) == 0 {
		log.Warning("no containers to remove")
		return nil
	}

	// select single random container from matching container and replace list with selected item
	if r.Random {
		log.Debug("selecting single random container")
		if c := container.RandomContainer(containers); c != nil {
			containers = []container.Container{*c}
		}
	}

	for _, container := range containers {
		log.WithFields(log.Fields{
			"container": container,
			"force":     r.Force,
			"links":     r.Links,
			"volumes":   r.Volumes,
		}).Debug("removing container")
		err := r.client.RemoveContainer(ctx, container, r.Force, r.Links, r.Volumes, r.DryRun)
		if err != nil {
			log.WithError(err).Error("failed to remove container")
			return err
		}
	}
	return nil
}
