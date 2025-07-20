# Console UI Multiplexer

A standalone terminal multiplexer that allows running multiple console commands simultaneously with a user-friendly UI.

## Features

- Run multiple commands in a single terminal window
- Switch between commands using keyboard or mouse
- View command output in a split view
- Copy text from command output
- Scroll through command history
- Keyboard shortcuts for navigation

## Installation

```bash
# Clone the repository
git clone https://github.com/nodge/multiplexer.git
cd multiplexer

# Build the multiplexer
go build ./cmd/multiplexer
```

## Usage

```bash
# Run the multiplexer with multiple commands
./multiplexer -cmd "ls -la" -cmd "echo Hello, World"

# Run more complex commands
./multiplexer -cmd "top" -cmd "tail -f /var/log/system.log" -cmd "htop"
```

## Keyboard Shortcuts

- `j/k` or `↓/↑`: Navigate between commands
- `Enter`: Focus on the selected command
- `Ctrl+Z`: Return to the sidebar from a focused command
- `Ctrl+U/D`: Scroll up/down
- `x`: Kill the selected command
- `Ctrl+C`: Exit the multiplexer

## How It Works

The multiplexer uses:
- `tcell` for terminal UI
- `pty` for pseudo-terminal handling
- Virtual terminal emulation for command output

Each command runs in its own pseudo-terminal, and the output is captured and displayed in the UI. The multiplexer handles keyboard and mouse input, and routes it to the appropriate command.

## Project Structure

```
multiplexer/
├── cmd/
│   └── multiplexer/
│       └── main.go         # Main entry point
├── internal/
│   ├── multiplexer/
│   │   ├── multiplexer.go  # Core multiplexer implementation
│   │   ├── process.go      # Process management
│   │   ├── draw.go         # UI rendering
│   │   ├── keycode.go      # Keyboard input handling
│   │   └── terminfo.go     # Terminal information
│   └── tcell-term/         # Terminal emulation
└── pkg/
    └── process/            # Process utilities
```

## Dependencies

- github.com/gdamore/tcell/v2
- github.com/creack/pty
- github.com/mattn/go-runewidth