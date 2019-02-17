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

func TestNewStopCommand(t *testing.T) {
	tests := []struct {
		name    string
		msg     StopMessage
		want    chaos.Command
		wantErr bool
	}{
		{
			name: "new stop command",
			msg: StopMessage{
				Names:    []string{"c1", "c2"},
				Pattern:  "pattern",
				Restart:  true,
				Interval: "20s",
				Duration: "10s",
				WaitTime: 100,
				Limit:    15,
			},
			want: &StopCommand{
				StopMessage{
					Names:    []string{"c1", "c2"},
					Pattern:  "pattern",
					Restart:  true,
					Interval: "20s",
					Duration: "10s",
					WaitTime: 100,
					Limit:    15,
				},
				nil,
			},
		},
		{
			name: "new stop command with default wait",
			msg: StopMessage{
				Names:    []string{"c1", "c2"},
				Pattern:  "pattern",
				Duration: "10s",
				WaitTime: 0,
				Limit:    15,
			},
			want: &StopCommand{
				StopMessage{
					Names:    []string{"c1", "c2"},
					Pattern:  "pattern",
					Duration: "10s",
					WaitTime: DeafultWaitTime,
					Limit:    15,
				},
				nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewStopCommand(nil, tt.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewStopCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewStopCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStopCommand_Run(t *testing.T) {
	type wantErrors struct {
		listError  bool
		stopError  bool
		startError bool
	}
	tests := []struct {
		name     string
		msg      StopMessage
		ctx      context.Context
		expected []container.Container
		wantErr  bool
		errs     wantErrors
	}{
		{
			name: "stop matching containers by names",
			msg: StopMessage{
				Names:    []string{"c1", "c2", "c3"},
				WaitTime: 20,
			},
			ctx:      context.TODO(),
			expected: container.CreateTestContainers(3),
		},
		{
			name: "stop matching containers by names and restart",
			msg: StopMessage{
				Names:    []string{"c1", "c2", "c3"},
				WaitTime: 20,
				Restart:  true,
			},
			ctx:      context.TODO(),
			expected: container.CreateTestContainers(3),
		},
		{
			name: "stop matching containers by filter with limit",
			msg: StopMessage{
				Pattern:  "^c?",
				WaitTime: 20,
				Limit:    2,
			},
			ctx:      context.TODO(),
			expected: container.CreateTestContainers(3),
		},
		{
			name: "stop random matching container by names",
			msg: StopMessage{
				Names:    []string{"c1", "c2", "c3"},
				WaitTime: 20,
				Random:   true,
			},
			ctx:      context.TODO(),
			expected: container.CreateTestContainers(3),
		},
		{
			name: "stop random matching container by names and restart",
			msg: StopMessage{
				Names:    []string{"c1", "c2", "c3"},
				WaitTime: 20,
				Restart:  true,
				Random:   true,
			},
			ctx:      context.TODO(),
			expected: container.CreateTestContainers(3),
		},
		{
			name: "no matching containers by names",
			msg: StopMessage{
				Names:    []string{"c1", "c2", "c3"},
				WaitTime: 20,
			},
			ctx: context.TODO(),
		},
		{
			name: "error listing containers",
			msg: StopMessage{
				Names:    []string{"c1", "c2", "c3"},
				WaitTime: 0,
			},
			ctx:     context.TODO(),
			wantErr: true,
			errs:    wantErrors{listError: true},
		},
		{
			name: "error stopping container",
			msg: StopMessage{
				Names:    []string{"c1", "c2", "c3"},
				WaitTime: 20,
			},
			ctx:      context.TODO(),
			expected: container.CreateTestContainers(3),
			wantErr:  true,
			errs:     wantErrors{stopError: true},
		},
		{
			name: "error starting stopped container",
			msg: StopMessage{
				Names:    []string{"c1", "c2", "c3"},
				WaitTime: 20,
				Restart:  true,
			},
			ctx:      context.TODO(),
			expected: container.CreateTestContainers(3),
			wantErr:  true,
			errs:     wantErrors{startError: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(container.MockClient)
			s := &StopCommand{tt.msg, mockClient}
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
				mockClient.On("StopContainer", tt.ctx, mock.AnythingOfType("container.Container"), tt.msg.WaitTime, tt.msg.DryRun).Return(nil)
				if tt.msg.Restart {
					mockClient.On("StartContainer", tt.ctx, mock.AnythingOfType("container.Container"), tt.msg.DryRun).Return(nil)
				}
			} else {
				for i := range tt.expected {
					if tt.msg.Limit == 0 || i < tt.msg.Limit {
						call = mockClient.On("StopContainer", tt.ctx, mock.AnythingOfType("container.Container"), tt.msg.WaitTime, tt.msg.DryRun)
						if tt.errs.stopError {
							call.Return(errors.New("ERROR"))
							goto Invoke
						} else {
							call.Return(nil)
						}
						if tt.msg.Restart {
							call = mockClient.On("StartContainer", tt.ctx, mock.AnythingOfType("container.Container"), tt.msg.DryRun)
							if tt.errs.startError {
								call.Return(errors.New("ERROR"))
								goto Invoke
							} else {
								call.Return(nil)
							}
						}
					}
				}
			}
		Invoke:
			if err := s.Run(tt.ctx, tt.msg.Random); (err != nil) != tt.wantErr {
				t.Errorf("StopCommand.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
			mockClient.AssertExpectations(t)
		})
	}
}
