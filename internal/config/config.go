package config

import (
	"fmt"
	"path/filepath"
)

// Config represents the main configuration file structure
type Config struct {
	Commands []Command `json:"commands" yaml:"commands"`
}

// Command represents the configuration for a single command
type Command struct {
	// Required fields
	Name    string   `json:"name" yaml:"name"`       // Unique identifier for the command
	Command []string `json:"command" yaml:"command"` // Array of command and arguments to execute

	// Optional fields with default values
	Title     string            `json:"title,omitempty" yaml:"title,omitempty"`         // Display name in the UI (defaults to `name`)
	CWD       string            `json:"cwd,omitempty" yaml:"cwd,omitempty"`             // Working directory for the command (relative or absolute)
	Env       map[string]string `json:"env,omitempty" yaml:"env,omitempty"`             // Environment variables to set for the command
	Autostart *bool             `json:"autostart,omitempty" yaml:"autostart,omitempty"` // Whether to start the command automatically (default: `true`)
	Killable  *bool             `json:"killable,omitempty" yaml:"killable,omitempty"`   // Whether the command can be killed manually (default: `true`)
}

// GetTitle returns the command title or name if title is not set
func (c *Command) GetTitle() string {
	if c.Title != "" {
		return c.Title
	}
	return "→ " + c.Name
}

// GetCWD returns the working directory or current directory
func (c *Command) GetCWD(defaultCWD string) string {
	if c.CWD != "" {
		if filepath.IsAbs(c.CWD) {
			return c.CWD
		}
		return filepath.Join(defaultCWD, c.CWD)
	}
	return defaultCWD
}

// IsAutostart returns true if the command should start automatically
func (c *Command) IsAutostart() bool {
	if c.Autostart != nil {
		return *c.Autostart
	}
	return true // autostart enabled by default
}

// IsKillable returns true if the command can be killed manually
func (c *Command) IsKillable() bool {
	if c.Killable != nil {
		return *c.Killable
	}
	return true // killable enabled by default
}

// Validate checks the correctness of the command configuration
func (c *Command) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("command name cannot be empty")
	}

	if len(c.Command) == 0 {
		return fmt.Errorf("command '%s': command array cannot be empty", c.Name)
	}

	if c.Command[0] == "" {
		return fmt.Errorf("command '%s': first element of command array cannot be empty", c.Name)
	}

	return nil
}

// Validate checks the correctness of the entire configuration
func (cfg *Config) Validate() error {
	if len(cfg.Commands) == 0 {
		return fmt.Errorf("configuration must contain at least one command")
	}

	names := make(map[string]bool)
	for i, cmd := range cfg.Commands {
		if err := cmd.Validate(); err != nil {
			return fmt.Errorf("command %d: %w", i+1, err)
		}

		if names[cmd.Name] {
			return fmt.Errorf("duplicate command name: '%s'", cmd.Name)
		}
		names[cmd.Name] = true
	}

	return nil
}
