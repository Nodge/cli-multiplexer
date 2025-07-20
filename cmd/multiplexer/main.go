package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/nodge/multiplexer/internal/multiplexer"
	"github.com/nodge/multiplexer/internal/process"
)

func main() {
	var commands stringSliceFlag
	flag.Var(&commands, "cmd", "Command to run in the multiplexer (can be specified multiple times)")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	process.Cleanup()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	multi, err := multiplexer.New(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating multiplexer: %v\n", err)
		os.Exit(1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current working directory: %v\n", err)
		os.Exit(1)
	}

	for i, cmd := range commands {
		parts := strings.Fields(cmd)
		if len(parts) == 0 {
			continue
		}

		name := fmt.Sprintf("cmd%d", i+1)
		title := parts[0]

		multi.AddProcess(
			name,
			parts,
			"→",
			title,
			cwd,
			true,
			true,
		)
	}

	multi.Start()
}

// stringSliceFlag implements flag.Value interface for string slice flags
type stringSliceFlag []string

func (s *stringSliceFlag) String() string {
	return strings.Join(*s, ", ")
}

func (s *stringSliceFlag) Set(value string) error {
	*s = append(*s, value)
	return nil
}
