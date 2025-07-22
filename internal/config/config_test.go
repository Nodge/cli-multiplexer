package config

import (
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Commands: []Command{
					{
						Name:    "test",
						Command: []string{"echo", "hello"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty commands",
			config: Config{
				Commands: []Command{},
			},
			wantErr: true,
		},
		{
			name: "duplicate command names",
			config: Config{
				Commands: []Command{
					{Name: "test", Command: []string{"echo", "1"}},
					{Name: "test", Command: []string{"echo", "2"}},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommand_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cmd     Command
		wantErr bool
	}{
		{
			name: "valid command",
			cmd: Command{
				Name:    "test",
				Command: []string{"echo", "hello"},
			},
			wantErr: false,
		},
		{
			name: "empty name",
			cmd: Command{
				Name:    "",
				Command: []string{"echo", "hello"},
			},
			wantErr: true,
		},
		{
			name: "empty command array",
			cmd: Command{
				Name:    "test",
				Command: []string{},
			},
			wantErr: true,
		},
		{
			name: "empty first command element",
			cmd: Command{
				Name:    "test",
				Command: []string{"", "hello"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cmd.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Command.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommand_GetTitle(t *testing.T) {
	tests := []struct {
		name string
		cmd  Command
		want string
	}{
		{
			name: "with title",
			cmd:  Command{Name: "test", Title: "Test Title"},
			want: "Test Title",
		},
		{
			name: "without title",
			cmd:  Command{Name: "test"},
			want: "→ test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cmd.GetTitle(); got != tt.want {
				t.Errorf("Command.GetTitle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCommand_IsAutostart(t *testing.T) {
	trueVal := true
	falseVal := false

	tests := []struct {
		name string
		cmd  Command
		want bool
	}{
		{
			name: "autostart true",
			cmd:  Command{Autostart: &trueVal},
			want: true,
		},
		{
			name: "autostart false",
			cmd:  Command{Autostart: &falseVal},
			want: false,
		},
		{
			name: "autostart nil (default)",
			cmd:  Command{},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cmd.IsAutostart(); got != tt.want {
				t.Errorf("Command.IsAutostart() = %v, want %v", got, tt.want)
			}
		})
	}
}
