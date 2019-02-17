package docker

import (
	"context"
	"time"

	"github.com/alexei-led/pumba/pkg/chaos"
	"github.com/alexei-led/pumba/pkg/container"
	"github.com/alexei-led/pumba/pkg/util"
	log "github.com/sirupsen/logrus"
)

const (
	// DeafultWaitTime time to wait before stopping container (in seconds)
	DeafultWaitTime = 5
)

// StopMessage REST API message
type StopMessage struct {
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

// StopCommand `docker stop` command
type StopCommand struct {
	StopMessage
	client container.Client
}

// NewStopCommand create new Stop Command instance
func NewStopCommand(client container.Client, msg StopMessage) (chaos.Command, error) {
	if msg.WaitTime <= 0 {
		msg.WaitTime = DeafultWaitTime
	}
	// get interval
	interval, err := util.GetIntervalValue(msg.Interval)
	if err != nil {
		return nil, err
	}
	// validate duration vs interval
	_, err = util.GetDurationValue(msg.Duration, interval)
	if err != nil {
		return nil, err
	}
	return &StopCommand{msg, client}, nil
}

// Run stop command
func (s *StopCommand) Run(ctx context.Context, random bool) error {
	log.Debug("stopping all matching containers")
	log.WithFields(log.Fields{
		"names":    s.Names,
		"pattern":  s.Pattern,
		"duration": s.Duration,
		"waitTime": s.WaitTime,
		"limit":    s.Limit,
	}).Debug("listing matching containers")
	containers, err := container.ListNContainers(ctx, s.client, s.Names, s.Pattern, s.Limit)
	if err != nil {
		log.WithError(err).Error("failed to list containers")
		return err
	}
	if len(containers) == 0 {
		log.Warning("no containers to stop")
		return nil
	}

	// get duration, error is already checked
	duration, _ := time.ParseDuration(s.Duration)

	// select single random container from matching container and replace list with selected item
	if s.Random {
		log.Debug("selecting single random container")
		if c := container.RandomContainer(containers); c != nil {
			containers = []container.Container{*c}
		}
	}

	// keep stopped containers
	stoppedContainers := []container.Container{}
	// pause containers
	for _, container := range containers {
		log.WithFields(log.Fields{
			"container": container,
			"waitTime":  s.WaitTime,
		}).Debug("stopping container")
		err = s.client.StopContainer(ctx, container, s.WaitTime, s.DryRun)
		if err != nil {
			log.WithError(err).Error("failed to stop container")
			break
		}
		stoppedContainers = append(stoppedContainers, container)
	}

	// if there are stopped containers and want to (re)start ...
	if len(stoppedContainers) > 0 && s.Restart {
		// wait for specified duration and then unpause containers or unpause on ctx.Done()
		select {
		case <-ctx.Done():
			log.Debug("start stopped containers by stop event")
			// NOTE: use different context to stop netem since parent context is canceled
			err = s.startStoppedContainers(context.Background(), stoppedContainers)
		case <-time.After(duration):
			log.WithField("duration", duration).Debug("start stopped containers after duration")
			err = s.startStoppedContainers(ctx, stoppedContainers)
		}
	}
	if err != nil {
		log.WithError(err).Error("failed to start stopped containers")
	}
	return err
}

// start previously stopped containers after duration on exit
func (s *StopCommand) startStoppedContainers(ctx context.Context, containers []container.Container) error {
	var err error
	for _, container := range containers {
		log.WithField("container", container).Debug("start stopped container")
		if e := s.client.StartContainer(ctx, container, s.DryRun); e != nil {
			log.WithError(e).Error("failed to start stopped container")
			err = e
		}
	}
	return err // last non nil error
}
