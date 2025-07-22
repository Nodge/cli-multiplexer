package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// Implements flag.Value interface for string slice flags
type stringSliceFlag []string

func (s *stringSliceFlag) String() string {
	return strings.Join(*s, ", ")
}

func (s *stringSliceFlag) Set(value string) error {
	*s = append(*s, value)
	return nil
}

type flagConfig struct {
	commands     stringSliceFlag
	configPath   string
	configFormat string
	fromStdin    bool
}

func parseFlags() *flagConfig {
	cfg := &flagConfig{}

	flag.Var(&cfg.commands, "cmd", "Command to run in the multiplexer (can be specified multiple times)")
	flag.StringVar(&cfg.configPath, "config", "", "Path to configuration file (JSON or YAML, format detected by extension)")
	flag.BoolVar(&cfg.fromStdin, "stdin", false, "Read configuration from stdin")
	flag.StringVar(&cfg.configFormat, "format", "", "Configuration format when reading from stdin (json or yaml, defaults to json)")
	flag.Parse()

	return cfg
}

func validateFlags(flags *flagConfig) error {
	hasConfig := flags.configPath != ""
	hasCommands := len(flags.commands) > 0
	hasStdin := flags.fromStdin

	inputMethods := 0
	if hasConfig {
		inputMethods++
	}
	if hasCommands {
		inputMethods++
	}
	if hasStdin {
		inputMethods++
	}

	if inputMethods == 0 {
		return fmt.Errorf("no commands specified. Use --config, --stdin, or --cmd flags")
	}

	if inputMethods > 1 {
		return fmt.Errorf("cannot specify multiple input methods. Use only one of: --config, --stdin, or --cmd")
	}

	if hasConfig {
		if _, err := os.Stat(flags.configPath); os.IsNotExist(err) {
			return fmt.Errorf("config file does not exist: %s", flags.configPath)
		}
	}

	if flags.configFormat != "" {
		if flags.configFormat != "json" && flags.configFormat != "yaml" {
			return fmt.Errorf("config format must be 'json' or 'yaml', got: %s", flags.configFormat)
		}

		if !hasStdin {
			return fmt.Errorf("config format can only be specified when reading from stdin")
		}
	}

	return nil
}
