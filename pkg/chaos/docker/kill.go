package docker

import (
	"context"
	"fmt"

	"github.com/alexei-led/pumba/pkg/chaos"
	"github.com/alexei-led/pumba/pkg/container"
	log "github.com/sirupsen/logrus"
)

const (
	// DefaultKillSignal default kill signal
	DefaultKillSignal = "SIGKILL"
)

// LinuxSignals valid Linux signal table
// http://www.comptechdoc.org/os/linux/programming/linux_pgsignals.html
var LinuxSignals = map[string]int{
	"SIGHUP":    1,
	"SIGINT":    2,
	"SIGQUIT":   3,
	"SIGILL":    4,
	"SIGTRAP":   5,
	"SIGIOT":    6,
	"SIGBUS":    7,
	"SIGFPE":    8,
	"SIGKILL":   9,
	"SIGUSR1":   10,
	"SIGSEGV":   11,
	"SIGUSR2":   12,
	"SIGPIPE":   13,
	"SIGALRM":   14,
	"SIGTERM":   15,
	"SIGSTKFLT": 16,
	"SIGCHLD":   17,
	"SIGCONT":   18,
	"SIGSTOP":   19,
	"SIGTSTP":   20,
	"SIGTTIN":   21,
	"SIGTTOU":   22,
	"SIGURG":    23,
	"SIGXCPU":   24,
	"SIGXFSZ":   25,
	"SIGVTALRM": 26,
	"SIGPROF":   27,
	"SIGWINCH":  28,
	"SIGIO":     29,
	"SIGPWR":    30,
}

// KillMessage message
type KillMessage struct {
	Random   bool     `json:"random,omitempty"`
	DryRun   bool     `json:"dry-run,omitempty"`
	Interval string   `json:"interval,omitempty"`
	Pattern  string   `json:"pattern,omitempty"`
	Names    []string `json:"names,omitempty"`
	Signal   string   `json:"signal,omitempty"`
	Limit    int      `json:"limit,omitempty"`
}

// KillCommand `docker kill` command
type KillCommand struct {
	KillMessage
	client container.Client
}

// NewKillCommand create new Kill Command instance
func NewKillCommand(client container.Client, msg KillMessage) (chaos.Command, error) {
	kill := &KillCommand{msg, client}
	if kill.Signal == "" {
		kill.Signal = DefaultKillSignal
	}
	if _, ok := LinuxSignals[kill.Signal]; !ok {
		err := fmt.Errorf("undefined Linux signal: %s", kill.Signal)
		log.WithError(err).Error("bad value for Linux signal")
		return nil, err
	}
	return kill, nil
}

// Run kill command
func (k *KillCommand) Run(ctx context.Context, random bool) error {
	log.Debug("killing all matching containers")
	log.WithFields(log.Fields{
		"names":   k.Names,
		"pattern": k.Pattern,
		"limit":   k.Limit,
	}).Debug("listing matching containers")
	containers, err := container.ListNContainers(ctx, k.client, k.Names, k.Pattern, k.Limit)
	if err != nil {
		log.WithError(err).Error("failed to list containers")
		return err
	}
	if len(containers) == 0 {
		log.Warning("no containers to kill")
		return nil
	}

	// select single random container from matching container and replace list with selected item
	if k.Random {
		log.Debug("selecting single random container")
		if c := container.RandomContainer(containers); c != nil {
			containers = []container.Container{*c}
		}
	}

	for _, container := range containers {
		log.WithFields(log.Fields{
			"container": container,
			"signal":    k.Signal,
		}).Debug("killing container")
		err := k.client.KillContainer(ctx, container, k.Signal, k.DryRun)
		if err != nil {
			log.WithError(err).Error("failed to kill container")
			return err
		}
	}
	return nil
}
