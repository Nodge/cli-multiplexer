package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// validNamePattern defines valid characters for command names
	validNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
	Value   interface{}
}

func (e ValidationError) Error() string {
	if e.Value != nil {
		return fmt.Sprintf("field '%s': %s (value: %v)", e.Field, e.Message, e.Value)
	}
	return fmt.Sprintf("field '%s': %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "no validation errors"
	}

	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// ValidateStrict performs strict validation of the configuration
func (cfg *Config) ValidateStrict() error {
	var errors ValidationErrors

	if len(cfg.Commands) == 0 {
		errors = append(errors, ValidationError{
			Field:   "commands",
			Message: "must contain at least one command",
		})
		return errors
	}

	names := make(map[string]int)
	for i, cmd := range cfg.Commands {
		cmdErrors := validateCommandStrict(&cmd, i)
		errors = append(errors, cmdErrors...)

		if cmd.Name != "" {
			if prevIndex, exists := names[cmd.Name]; exists {
				errors = append(errors, ValidationError{
					Field:   fmt.Sprintf("commands[%d].name", i),
					Message: fmt.Sprintf("duplicate command name, first occurrence at index %d", prevIndex),
					Value:   cmd.Name,
				})
			} else {
				names[cmd.Name] = i
			}
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// validateCommandStrict performs strict validation of a single command
func validateCommandStrict(cmd *Command, index int) ValidationErrors {
	var errors ValidationErrors
	prefix := fmt.Sprintf("commands[%d]", index)

	// Validate name
	if cmd.Name == "" {
		errors = append(errors, ValidationError{
			Field:   prefix + ".name",
			Message: "cannot be empty",
		})
	} else if !validNamePattern.MatchString(cmd.Name) {
		errors = append(errors, ValidationError{
			Field:   prefix + ".name",
			Message: "must contain only alphanumeric characters, underscores, and hyphens",
			Value:   cmd.Name,
		})
	}

	// Validate command array
	if len(cmd.Command) == 0 {
		errors = append(errors, ValidationError{
			Field:   prefix + ".command",
			Message: "cannot be empty",
		})
	} else {
		for i, part := range cmd.Command {
			if part == "" {
				errors = append(errors, ValidationError{
					Field:   fmt.Sprintf("%s.command[%d]", prefix, i),
					Message: "cannot be empty string",
				})
			}
		}
	}

	// Validate CWD if specified
	if cmd.CWD != "" {
		if err := validateDirectory(cmd.CWD); err != nil {
			errors = append(errors, ValidationError{
				Field:   prefix + ".cwd",
				Message: err.Error(),
				Value:   cmd.CWD,
			})
		}
	}

	// Validate environment variables
	for key, value := range cmd.Env {
		if key == "" {
			errors = append(errors, ValidationError{
				Field:   prefix + ".env",
				Message: "environment variable key cannot be empty",
			})
		}
		if strings.Contains(key, "=") {
			errors = append(errors, ValidationError{
				Field:   prefix + ".env",
				Message: "environment variable key cannot contain '=' character",
				Value:   key,
			})
		}
		// Note: empty values are allowed for env vars
		_ = value
	}

	return errors
}

// validateDirectory checks if a directory path is valid
func validateDirectory(path string) error {
	if path == "" {
		return fmt.Errorf("directory path cannot be empty")
	}

	// If it's an absolute path, check if it exists
	if filepath.IsAbs(path) {
		if info, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("directory does not exist: %s", path)
			}
			return fmt.Errorf("cannot access directory: %s", err.Error())
		} else if !info.IsDir() {
			return fmt.Errorf("path is not a directory: %s", path)
		}
	}

	return nil
}

// ValidateAtRuntime performs runtime validation that requires context
func (cfg *Config) ValidateAtRuntime(baseCWD string) error {
	var errors ValidationErrors

	for i, cmd := range cfg.Commands {
		prefix := fmt.Sprintf("commands[%d]", i)

		// Validate actual working directory
		actualCWD := cmd.GetCWD(baseCWD)
		if info, err := os.Stat(actualCWD); err != nil {
			if os.IsNotExist(err) {
				errors = append(errors, ValidationError{
					Field:   prefix + ".cwd",
					Message: "resolved directory does not exist",
					Value:   actualCWD,
				})
			} else {
				errors = append(errors, ValidationError{
					Field:   prefix + ".cwd",
					Message: fmt.Sprintf("cannot access resolved directory: %s", err.Error()),
					Value:   actualCWD,
				})
			}
		} else if !info.IsDir() {
			errors = append(errors, ValidationError{
				Field:   prefix + ".cwd",
				Message: "resolved path is not a directory",
				Value:   actualCWD,
			})
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}
