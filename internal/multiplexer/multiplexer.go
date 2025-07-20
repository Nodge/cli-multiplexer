package multiplexer

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/nodge/multiplexer/internal/process"
	tcellterm "github.com/nodge/multiplexer/internal/tcell-term"
)

type Multiplexer struct {
	ctx       context.Context
	panes     []*pane
	ui        *UI
	eventLoop *EventLoop
}

func New(ctx context.Context) (*Multiplexer, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}

	err = screen.Init()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize screen: %w", err)
	}

	screen.EnableMouse()
	screen.Show()

	result := &Multiplexer{
		ctx:   ctx,
		panes: []*pane{},
		ui:    NewUI(screen),
	}

	result.enableTmuxClipboard()

	return result, nil
}

// Start begins the main event loop for the multiplexer.
// Handles all user input, terminal events, and process management.
// The loop continues until the context is cancelled or an exit event is received.
func (s *Multiplexer) Start() {
	defer func() {
		s.ui.stop()
	}()

	s.ui.start()

	eventLoop := NewEventLoop(s)
	eventLoop.Run(s.ctx)
}

// Posts an exit event to the event queue, triggering graceful shutdown.
func (s *Multiplexer) Exit() {
	s.ui.screen.PostEvent(&EventExit{})
}

// AddProcess posts an event to add a new process to the multiplexer
func (s *Multiplexer) AddProcess(key string, args []string, icon string, title string, cwd string, killable bool, autostart bool, env ...string) {
	s.ui.screen.PostEvent(&EventProcess{
		Key:       key,
		Args:      args,
		Icon:      icon,
		Title:     title,
		Cwd:       cwd,
		Killable:  killable,
		Autostart: autostart,
		Env:       env,
	})
}

// Creates new pane and attaches virtual terminal
func (s *Multiplexer) addPane(p *pane) *pane {
	p.vt = tcellterm.New()
	p.vt.SetSurface(s.ui.activePaneView)
	// Forward terminal events back to the main event loop
	p.vt.Attach(func(ev tcell.Event) {
		s.ui.screen.PostEvent(ev)
	})

	s.panes = append(s.panes, p)
	s.ui.addPane(p)

	return p
}

// resize delegates to the UI's Resize method
func (s *Multiplexer) resize(width int, height int) {
	s.ui.resize(width, height)
}

// Scrolls the selected terminal down by n lines and refreshes the display.
func (s *Multiplexer) scrollDown(n int) {
	selected := s.ui.selectedPane()
	if selected == nil {
		return
	}
	selected.scrollDown(n)
	s.ui.draw()
	s.ui.screen.Sync()
}

// Scrolls the selected terminal up by n lines and refreshes the display.
func (s *Multiplexer) scrollUp(n int) {
	selected := s.ui.selectedPane()
	if selected == nil {
		return
	}
	selected.scrollUp(n)
	s.ui.draw()
}

// Handles clipboard operations for selected text.
// Uses platform-specific methods: pbcopy on macOS Terminal, OSC 52 escape sequences elsewhere.
// OSC 52 allows terminal applications to set clipboard content via escape sequences.
func (s *Multiplexer) copy() {
	selected := s.ui.selectedPane()
	if selected == nil {
		return
	}

	data := selected.vt.Copy()
	if data == "" {
		return
	}

	// Use pbcopy on macOS Terminal for better integration
	if os.Getenv("TERM_PROGRAM") == "Apple_Terminal" {
		cmd := process.Command("pbcopy")
		cmd.Stdin = strings.NewReader(data)
		err := cmd.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to copy to clipboard: %v\n", err)
		}
		return
	}

	// Use OSC 52 escape sequence for universal clipboard support
	encoded := base64.StdEncoding.EncodeToString([]byte(data))
	fmt.Fprintf(os.Stdout, "\x1b]52;c;%s\x07", encoded)
}

// Enables clipboard integration when the multiplexer is running within a tmux session.
func (m *Multiplexer) enableTmuxClipboard() {
	const tmuxEnvVar = "TMUX"
	const tmuxClipboardOption = "set-clipboard"

	if os.Getenv("TMUX") == "" {
		return // Not running inside tmux, no action needed
	}

	process.Command("tmux", "set-option", "-p", "set-clipboard", "on").Run()
}
