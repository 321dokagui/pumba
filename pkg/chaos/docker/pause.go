package docker

import (
	"context"
	"time"

	"github.com/alexei-led/pumba/pkg/chaos"
	"github.com/alexei-led/pumba/pkg/container"
	"github.com/alexei-led/pumba/pkg/util"
	log "github.com/sirupsen/logrus"
)

// PauseMessage REST API message
type PauseMessage struct {
	Random   bool     `json:"random,omitempty"`
	DryRun   bool     `json:"dry-run,omitempty"`
	Interval string   `json:"interval,omitempty"`
	Pattern  string   `json:"pattern,omitempty"`
	Names    []string `json:"names,omitempty"`
	Duration string   `json:"duration,omitempty"`
	Limit    int      `json:"limit,omitempty"`
}

// PauseCommand `docker pause` command
type PauseCommand struct {
	PauseMessage
	client container.Client
}

// NewPauseCommand create new Pause Command instance
func NewPauseCommand(client container.Client, msg PauseMessage) (chaos.Command, error) {
	// get interval
	interval, err := util.GetIntervalValue(msg.Interval)
	if err != nil {
		return nil, err
	}
	// validate duration
	_, err = util.GetDurationValue(msg.Duration, interval)
	if err != nil {
		return nil, err
	}
	return &PauseCommand{msg, client}, nil
}

// Run pause command
func (p *PauseCommand) Run(ctx context.Context, random bool) error {
	log.Debug("pausing all matching containers")
	log.WithFields(log.Fields{
		"names":    p.Names,
		"pattern":  p.Pattern,
		"duration": p.Duration,
		"limit":    p.Limit,
	}).Debug("listing matching containers")
	containers, err := container.ListNContainers(ctx, p.client, p.Names, p.Pattern, p.Limit)
	if err != nil {
		log.WithError(err).Error("failed to list containers")
		return err
	}
	if len(containers) == 0 {
		log.Warning("no containers to stop")
		return nil
	}

	// get duration, error is already checked
	duration, _ := time.ParseDuration(p.Duration)

	// select single random container from matching container and replace list with selected item
	if random {
		log.Debug("selecting single random container")
		if c := container.RandomContainer(containers); c != nil {
			containers = []container.Container{*c}
		}
	}

	// keep paused containers
	pausedContainers := []container.Container{}
	// pause containers
	for _, container := range containers {
		log.WithFields(log.Fields{
			"container": container,
			"duration":  duration,
		}).Debug("pausing container for duration")
		err = p.client.PauseContainer(ctx, container, p.DryRun)
		if err != nil {
			log.WithError(err).Error("failed to pause container")
			break
		}
		pausedContainers = append(pausedContainers, container)
	}

	// if there are paused containers unpause them
	if len(pausedContainers) > 0 {
		// wait for specified duration and then unpause containers or unpause on ctx.Done()
		select {
		case <-ctx.Done():
			log.Debug("unpause containers by stop event")
			// NOTE: use different context to stop netem since parent context is canceled
			err = p.unpauseContainers(context.Background(), pausedContainers)
		case <-time.After(duration):
			log.WithField("duration", duration).Debug("unpause containers after duration")
			err = p.unpauseContainers(ctx, pausedContainers)
		}
	}
	if err != nil {
		log.WithError(err).Error("failed to unpause paused containers")
	}
	return err
}

// unpause containers
func (p *PauseCommand) unpauseContainers(ctx context.Context, containers []container.Container) error {
	var err error
	for _, container := range containers {
		log.WithField("container", container).Debug("unpause container")
		if e := p.client.UnpauseContainer(ctx, container, p.DryRun); e != nil {
			log.WithError(e).Error("failed to unpause container")
			err = e
		}
	}
	return err // last non nil error
}
