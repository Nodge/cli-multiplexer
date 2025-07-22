# Console UI Multiplexer

A terminal multiplexer with a sidebar-based TUI interface. It manages multiple terminal processes in separate panes, allowing users to switch between them and interact with each process independently.

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

### Command Line Arguments

```bash
# Run the multiplexer with multiple commands
./multiplexer --cmd "ls -la" --cmd "echo Hello, World"

# Run more complex commands
./multiplexer --cmd "top" --cmd "tail -f /var/log/system.log" --cmd "htop"
```

### Configuration Files

The multiplexer supports configuration files in JSON and YAML formats for more advanced setups:

```bash
# Load configuration from a file (format auto-detected by extension)
./multiplexer --config config.json
./multiplexer --config config.yaml

# Load configuration from stdin
cat config.json | ./multiplexer --stdin
echo '{"commands":[...]}' | ./multiplexer --stdin --format json
```
#### Configuration File Options

Each command in the configuration supports the following options:

- **`name`** (required): Unique identifier for the command
- **`command`** (required): Array of command and arguments to execute
- **`title`** (optional): Display name in the UI (defaults to `name`)
- **`cwd`** (optional): Working directory for the command (relative or absolute)
- **`env`** (optional): Environment variables to set for the command
- **`autostart`** (optional): Whether to start the command automatically (default: `true`)
- **`killable`** (optional): Whether the command can be killed manually (default: `true`)

#### JSON Configuration Example

```json
{
  "commands": [
    {
      "name": "backend",
      "title": "🚀 API Server",
      "command": ["go", "run", "./cmd/server"],
      "cwd": "./backend",
      "autostart": true,
      "killable": false,
      "env": {
        "PORT": "8080",
        "NODE_ENV": "development"
      }
    },
    {
      "name": "frontend",
      "title": "⚡ Web UI",
      "command": ["npm", "run", "dev"],
      "autostart": false
    }
  ]
}
```

#### YAML Configuration Example

```yaml
commands:
  - name: "backend"
    title: "🚀 API Server"
    command: ["go", "run", "./cmd/server"]
    cwd: "./backend"
    autostart: true
    killable: false
    env:
      PORT: "8080"
      NODE_ENV: "development"
  
  - name: "frontend"
    title: "⚡ Web UI"
    command: ["npm", "run", "dev"]
    autostart: false
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
