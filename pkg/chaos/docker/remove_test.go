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

func TestRemoveCommand_Run(t *testing.T) {
	type wantErrors struct {
		listError   bool
		removeError bool
	}
	tests := []struct {
		name     string
		msg      RemoveMessage
		ctx      context.Context
		expected []container.Container
		wantErr  bool
		errs     wantErrors
	}{
		{
			name: "remove matching containers by names",
			msg: RemoveMessage{
				Names: []string{"c1", "c2", "c3"},
				Force: true,
			},
			ctx:      context.TODO(),
			expected: container.CreateTestContainers(3),
		},
		{
			name: "remove matching containers by filter with limit",
			msg: RemoveMessage{
				Pattern: "^c?",
				Force:   true,
				Links:   true,
				Limit:   2,
			},
			ctx:      context.TODO(),
			expected: container.CreateTestContainers(3),
		},
		{
			name: "remove random matching container by names",
			msg: RemoveMessage{
				Names:   []string{"c1", "c2", "c3"},
				Force:   true,
				Links:   true,
				Volumes: true,
				Random:  true,
			},
			ctx:      context.TODO(),
			expected: container.CreateTestContainers(3),
		},
		{
			name: "no matching containers by names",
			msg: RemoveMessage{
				Names: []string{"c1", "c2", "c3"},
			},
			ctx: context.TODO(),
		},
		{
			name: "error listing containers",
			msg: RemoveMessage{
				Names: []string{"c1", "c2", "c3"},
			},
			ctx:     context.TODO(),
			wantErr: true,
			errs:    wantErrors{listError: true},
		},
		{
			name: "error removing container",
			msg: RemoveMessage{
				Names: []string{"c1", "c2", "c3"},
			},
			ctx:      context.TODO(),
			expected: container.CreateTestContainers(3),
			wantErr:  true,
			errs:     wantErrors{removeError: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(container.MockClient)
			k := &RemoveCommand{tt.msg, mockClient}
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
				mockClient.On("RemoveContainer", tt.ctx, mock.AnythingOfType("container.Container"), tt.msg.Force, tt.msg.Links, tt.msg.Volumes, tt.msg.DryRun).Return(nil)
			} else {
				for i := range tt.expected {
					if tt.msg.Limit == 0 || i < tt.msg.Limit {
						call = mockClient.On("RemoveContainer", tt.ctx, mock.AnythingOfType("container.Container"), tt.msg.Force, tt.msg.Links, tt.msg.Volumes, tt.msg.DryRun)
						if tt.errs.removeError {
							call.Return(errors.New("ERROR"))
							goto Invoke
						} else {
							call.Return(nil)
						}
					}
				}
			}
		Invoke:
			if err := k.Run(tt.ctx, tt.msg.Random); (err != nil) != tt.wantErr {
				t.Errorf("RemoveCommand.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

func TestNewRemoveCommand(t *testing.T) {
	tests := []struct {
		name    string
		msg     RemoveMessage
		want    chaos.Command
		wantErr bool
	}{
		{
			name: "create new remove command",
			msg: RemoveMessage{
				Names:   []string{"c1", "c2"},
				Force:   true,
				Links:   true,
				Volumes: false,
				Limit:   10,
			},
			want: &RemoveCommand{
				RemoveMessage{
					Names:   []string{"c1", "c2"},
					Force:   true,
					Links:   true,
					Volumes: false,
					Limit:   10,
				},
				nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRemoveCommand(nil, tt.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRemoveCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRemoveCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}
