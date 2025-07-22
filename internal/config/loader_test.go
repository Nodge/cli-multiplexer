package config

import (
	"strings"
	"testing"
)

func TestLoadFromReader_JSON(t *testing.T) {
	jsonConfig := `{
		"commands": [
			{
				"name": "test",
				"command": ["echo", "hello"],
				"title": "Test Command",
				"prompt": "🚀"
			}
		]
	}`

	reader := strings.NewReader(jsonConfig)
	config, err := LoadFromReader(reader, "json")

	if err != nil {
		t.Fatalf("LoadFromReader() error = %v", err)
	}

	if len(config.Commands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(config.Commands))
	}

	cmd := config.Commands[0]
	if cmd.Name != "test" {
		t.Errorf("Expected name 'test', got '%s'", cmd.Name)
	}

	if cmd.GetTitle() != "Test Command" {
		t.Errorf("Expected title 'Test Command', got '%s'", cmd.GetTitle())
	}
}

func TestLoadFromReader_YAML(t *testing.T) {
	yamlConfig := `
commands:
  - name: test
    command: ["echo", "hello"]
    title: Test Command
    prompt: "🚀"
    autostart: false
`

	reader := strings.NewReader(yamlConfig)
	config, err := LoadFromReader(reader, "yaml")

	if err != nil {
		t.Fatalf("LoadFromReader() error = %v", err)
	}

	if len(config.Commands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(config.Commands))
	}

	cmd := config.Commands[0]
	if cmd.Name != "test" {
		t.Errorf("Expected name 'test', got '%s'", cmd.Name)
	}

	if cmd.IsAutostart() != false {
		t.Errorf("Expected autostart false, got %v", cmd.IsAutostart())
	}
}

func TestLoadFromReader_InvalidJSON(t *testing.T) {
	invalidJSON := `{"commands": [{"name": "test", "command":}]}`

	reader := strings.NewReader(invalidJSON)
	_, err := LoadFromReader(reader, "json")

	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestLoadFromReader_InvalidYAML(t *testing.T) {
	invalidYAML := `
commands:
  - name: test
    command: [
      - invalid yaml structure
`

	reader := strings.NewReader(invalidYAML)
	_, err := LoadFromReader(reader, "yaml")

	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

func TestLoadFromReader_ValidationError(t *testing.T) {
	invalidConfig := `{
		"commands": [
			{
				"name": "",
				"command": ["echo", "hello"]
			}
		]
	}`

	reader := strings.NewReader(invalidConfig)
	_, err := LoadFromReader(reader, "json")

	if err == nil {
		t.Error("Expected validation error for empty name, got nil")
	}

	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("Expected validation error message, got: %v", err)
	}
}

func TestLoadFromReader_EmptyData(t *testing.T) {
	reader := strings.NewReader("")
	_, err := LoadFromReader(reader, "json")

	if err == nil {
		t.Error("Expected error for empty data, got nil")
	}

	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("Expected empty data error message, got: %v", err)
	}
}

func TestLoadFromStdin_DefaultFormat(t *testing.T) {
	jsonConfig := `{"commands":[{"name":"test","command":["echo","hello"]}]}`

	// Test with empty format (should default to JSON)
	reader := strings.NewReader(jsonConfig)
	config, err := LoadFromReader(reader, "")

	if err != nil {
		t.Fatalf("LoadFromStdin() error = %v", err)
	}

	if len(config.Commands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(config.Commands))
	}
}
