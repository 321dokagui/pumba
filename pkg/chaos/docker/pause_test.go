package docker

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/alexei-led/pumba/pkg/chaos"
	"github.com/alexei-led/pumba/pkg/container"
	"github.com/stretchr/testify/mock"
)

func TestNewPauseCommand(t *testing.T) {
	tests := []struct {
		name    string
		msg     PauseMessage
		want    chaos.Command
		wantErr bool
	}{
		{
			name: "new pause command",
			msg: PauseMessage{
				Names:    []string{"c1", "c2"},
				Pattern:  "pattern",
				Interval: "20s",
				Duration: "10s",
				Limit:    15,
			},
			want: &PauseCommand{
				PauseMessage{
					Names:    []string{"c1", "c2"},
					Pattern:  "pattern",
					Interval: "20s",
					Duration: "10s",
					Limit:    15,
				},
				nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPauseCommand(nil, tt.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPauseCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPauseCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPauseCommand_Run(t *testing.T) {
	type wantErrors struct {
		listError    bool
		pauseError   bool
		unpauseError bool
	}
	tests := []struct {
		name     string
		ctx      context.Context
		msg      PauseMessage
		expected []container.Container
		wantErr  bool
		errs     wantErrors
	}{
		{
			name: "pause matching containers by names",
			msg: PauseMessage{
				Names: []string{"c1", "c2", "c3"},
			},
			ctx:      context.TODO(),
			expected: container.CreateTestContainers(3),
		},
		{
			name: "pause matching containers by filter with limit",
			msg: PauseMessage{
				Pattern: "^c?",
				Limit:   2,
			},
			ctx:      context.TODO(),
			expected: container.CreateTestContainers(3),
		},
		{
			name: "pause random matching container by names",
			msg: PauseMessage{
				Names:  []string{"c1", "c2", "c3"},
				Random: true,
			},
			ctx:      context.TODO(),
			expected: container.CreateTestContainers(3),
		},
		{
			name: "no matching containers by names",
			msg: PauseMessage{
				Names: []string{"c1", "c2", "c3"},
			},
			ctx: context.TODO(),
		},
		{
			name: "error listing containers",
			msg: PauseMessage{
				Names: []string{"c1", "c2", "c3"},
			},
			ctx:     context.TODO(),
			wantErr: true,
			errs:    wantErrors{listError: true},
		},
		{
			name: "error pausing container",
			msg: PauseMessage{
				Names: []string{"c1", "c2", "c3"},
			},
			ctx:      context.TODO(),
			expected: container.CreateTestContainers(3),
			wantErr:  true,
			errs:     wantErrors{pauseError: true},
		},
		{
			name: "error unpausing paused container",
			msg: PauseMessage{
				Names: []string{"c1", "c2", "c3"},
			},
			ctx:      context.TODO(),
			expected: container.CreateTestContainers(3),
			wantErr:  true,
			errs:     wantErrors{unpauseError: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(container.MockClient)
			s := &PauseCommand{tt.msg, mockClient}
			call := mockClient.On("ListContainers", tt.ctx, mock.AnythingOfType("container.Filter"))
			if tt.errs.listError {
				call.Return(tt.expected, errors.New("ERROR"))
				goto Invoke
			} else {
				call.Return(tt.expected, nil)
				if tt.expected == nil {
					goto Invoke
				}
			}
			if tt.msg.Random {
				mockClient.On("PauseContainer", tt.ctx, mock.AnythingOfType("container.Container"), tt.msg.DryRun).Return(nil)
				mockClient.On("UnpauseContainer", tt.ctx, mock.AnythingOfType("container.Container"), tt.msg.DryRun).Return(nil)
			} else {
				for i := range tt.expected {
					if tt.msg.Limit == 0 || i < tt.msg.Limit {
						call = mockClient.On("PauseContainer", tt.ctx, mock.AnythingOfType("container.Container"), tt.msg.DryRun)
						if tt.errs.pauseError {
							call.Return(errors.New("ERROR"))
							goto Invoke
						} else {
							call.Return(nil)
						}
						call = mockClient.On("UnpauseContainer", tt.ctx, mock.AnythingOfType("container.Container"), tt.msg.DryRun)
						if tt.errs.unpauseError {
							call.Return(errors.New("ERROR"))
							goto Invoke
						} else {
							call.Return(nil)
						}
					}
				}
			}
		Invoke:
			if err := s.Run(tt.ctx, tt.msg.Random); (err != nil) != tt.wantErr {
				t.Errorf("PauseCommand.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
			mockClient.AssertExpectations(t)
		})
	}
}
