package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/nodge/multiplexer/internal/config"
	"github.com/nodge/multiplexer/internal/multiplexer"
	"github.com/nodge/multiplexer/internal/process"
)

func main() {
	flags := parseFlags()

	if err := validateFlags(flags); err != nil {
		showUsage(err.Error())
	}

	ctx, cancel := setupContext()
	defer cancel()

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting current working directory: %v", err)
		os.Exit(1)
	}

	m, err := multiplexer.New(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating multiplexer: %v", err)
		os.Exit(1)
	}

	defer process.Cleanup()

	if flags.configPath != "" || flags.fromStdin {
		cfg, err := loadConfiguration(flags.configPath, flags.fromStdin, flags.configFormat, cwd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		addProcessesFromConfig(m, cfg, cwd)
	} else if len(flags.commands) > 0 {
		addProcessesFromFlags(m, flags.commands, cwd)
	}

	m.Start()
}

func setupContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	return ctx, cancel
}

func loadConfiguration(configPath string, fromStdin bool, configFormat string, cwd string) (*config.Config, error) {
	cfg, err := config.Load(configPath, fromStdin, configFormat)
	if err != nil {
		return nil, fmt.Errorf("error loading configuration: %v", err)
	}

	if err := cfg.ValidateAtRuntime(cwd); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %v", err)
	}

	return cfg, nil
}

func addProcessesFromConfig(m *multiplexer.Multiplexer, cfg *config.Config, cwd string) {
	for _, cmd := range cfg.Commands {
		m.AddProcess(
			cmd.Name,
			cmd.Command,
			cmd.Env,
			cmd.GetTitle(),
			cmd.GetCWD(cwd),
			cmd.IsKillable(),
			cmd.IsAutostart(),
		)
	}
}

func addProcessesFromFlags(m *multiplexer.Multiplexer, commands []string, cwd string) {
	for i, command := range commands {
		cmd := strings.Fields(command)
		if len(cmd) == 0 {
			continue
		}

		name := fmt.Sprintf("cmd%d", i+1)
		title := "→ " + name
		env := make(map[string]string)

		m.AddProcess(
			name,
			cmd,
			env,
			title,
			cwd,
			true,
			true,
		)
	}
}

func showUsage(errMsg string) {
	if errMsg != "" {
		fmt.Fprintf(os.Stderr, "Error: %s\n", errMsg)
	}
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  %s --config config.yaml\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s --stdin --format yaml < config.yaml\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s --cmd \"go run main.go\" --cmd \"npm start\"\n", os.Args[0])
	os.Exit(1)
}
