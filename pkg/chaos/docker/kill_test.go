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

func TestKillCommand_Run(t *testing.T) {
	type wantErrors struct {
		listError bool
		killError bool
	}
	type args struct {
		ctx    context.Context
		random bool
	}
	tests := []struct {
		name     string
		message  KillMessage
		args     args
		expected []container.Container
		wantErr  bool
		errs     wantErrors
	}{
		{
			name: "kill matching containers by names",
			message: KillMessage{
				Names:  []string{"c1", "c2", "c3"},
				Signal: "SIGKILL",
			},
			args: args{
				ctx: context.TODO(),
			},
			expected: container.CreateTestContainers(3),
		},
		{
			name: "kill matching containers by filter with limit",
			message: KillMessage{
				Pattern: "^c?",
				Signal:  "SIGSTOP",
				Limit:   2,
			},
			args: args{
				ctx: context.TODO(),
			},
			expected: container.CreateTestContainers(3),
		},
		{
			name: "kill random matching container by names",
			message: KillMessage{
				Names:  []string{"c1", "c2", "c3"},
				Signal: "SIGKILL",
			},
			args: args{
				ctx:    context.TODO(),
				random: true,
			},
			expected: container.CreateTestContainers(3),
		},
		{
			name: "no matching containers by names",
			message: KillMessage{
				Names:  []string{"c1", "c2", "c3"},
				Signal: "SIGKILL",
			},
			args: args{
				ctx: context.TODO(),
			},
		},
		{
			name: "error listing containers",
			message: KillMessage{
				Names:  []string{"c1", "c2", "c3"},
				Signal: "SIGKILL",
			},
			args: args{
				ctx: context.TODO(),
			},
			wantErr: true,
			errs:    wantErrors{listError: true},
		},
		{
			name: "error killing container",
			message: KillMessage{
				Names:  []string{"c1", "c2", "c3"},
				Signal: "SIGKILL",
			},
			args: args{
				ctx: context.TODO(),
			},
			expected: container.CreateTestContainers(3),
			wantErr:  true,
			errs:     wantErrors{killError: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(container.MockClient)
			k := &KillCommand{
				tt.message,
				mockClient,
			}
			call := mockClient.On("ListContainers", tt.args.ctx, mock.AnythingOfType("container.Filter"))
			if tt.errs.listError {
				call.Return(tt.expected, errors.New("ERROR"))
				goto Invoke
			} else {
				call.Return(tt.expected, nil)
				if tt.expected == nil {
					goto Invoke
				}
			}
			if tt.args.random {
				mockClient.On("KillContainer", tt.args.ctx, mock.AnythingOfType("container.Container"), tt.message.Signal, tt.message.DryRun).Return(nil)
			} else {
				for i := range tt.expected {
					if tt.message.Limit == 0 || i < tt.message.Limit {
						call = mockClient.On("KillContainer", tt.args.ctx, mock.AnythingOfType("container.Container"), tt.message.Signal, tt.message.DryRun)
						if tt.errs.killError {
							call.Return(errors.New("ERROR"))
							goto Invoke
						} else {
							call.Return(nil)
						}
					}
				}
			}
		Invoke:
			if err := k.Run(tt.args.ctx, tt.args.random); (err != nil) != tt.wantErr {
				t.Errorf("KillCommand.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

func TestNewKillCommand(t *testing.T) {
	tests := []struct {
		name    string
		msg     KillMessage
		want    chaos.Command
		wantErr bool
	}{
		{
			name: "create new kill command",
			msg: KillMessage{
				Names:  []string{"c1", "c2"},
				Signal: "SIGTERM",
				Limit:  10,
			},
			want: &KillCommand{
				KillMessage{
					Names:  []string{"c1", "c2"},
					Signal: "SIGTERM",
					Limit:  10,
				},
				nil,
			},
		},
		{
			name: "invalid signal",
			msg: KillMessage{
				Names:  []string{"c1", "c2"},
				Signal: "SIGNONE",
			},
			wantErr: true,
		},
		{
			name: "empty signal",
			msg: KillMessage{
				Names:  []string{"c1", "c2"},
				Signal: "",
			},
			want: &KillCommand{
				KillMessage{
					Names:  []string{"c1", "c2"},
					Signal: DefaultKillSignal,
				},
				nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewKillCommand(nil, tt.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewKillCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewKillCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}
