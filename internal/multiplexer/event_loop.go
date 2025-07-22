package multiplexer

import (
	"context"
	"log/slog"
	"os"
	"runtime/debug"
	"syscall"

	"github.com/gdamore/tcell/v2"
	"github.com/nodge/multiplexer/internal/process"
	tcellterm "github.com/nodge/multiplexer/internal/tcell-term"
)

// Represents a request to create or manage a terminal process
type EventProcess struct {
	tcell.EventTime
	Key       string
	Cmd       []string
	Env       map[string]string
	Title     string
	Cwd       string
	Killable  bool
	Autostart bool
}

// EventExit is a custom event used to signal the multiplexer to shut down gracefully
type EventExit struct {
	tcell.EventTime
}

// Manages the main event loop and event processing for the multiplexer
type EventLoop struct {
	multiplexer *Multiplexer
	ui          *UI
}

// Creates a new event loop for the given multiplexer
func NewEventLoop(m *Multiplexer) *EventLoop {
	return &EventLoop{
		multiplexer: m,
		ui:          m.ui,
	}
}

// Runs the main event loop with panic recovery
func (eh *EventLoop) Run(ctx context.Context) {
	eventCh := make(chan tcell.Event, 1)

	go func() {
		for {
			evt := eh.ui.screen.PollEvent()
			if evt == nil {
				continue
			}

			select {
			case eventCh <- evt:
			case <-ctx.Done():
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return

		case evt := <-eventCh:
			shouldExit := false
			func() {
				defer func() {
					if r := recover(); r != nil {
						slog.Error("multiplexer panic", "err", r, "stack", string(debug.Stack()))
					}
				}()

				shouldExit = eh.handleEvent(evt)
			}()

			if shouldExit {
				return
			}
		}
	}
}

// Processes a single event and returns true if the event loop should exit
func (eh *EventLoop) handleEvent(evt tcell.Event) bool {
	switch e := evt.(type) {
	case *EventExit:
		return true

	case *EventProcess:
		eh.handleProcessEvent(e)

	case *tcell.EventMouse:
		eh.handleMouseEvent(e)

	case *tcell.EventResize:
		eh.handleResizeEvent(e)

	case *tcellterm.EventRedraw:
		eh.handleRedrawEvent(e)

	case *tcellterm.EventClosed:
		eh.handleClosedEvent(e)

	case *tcell.EventKey:
		eh.handleKeyEvent(e)
	}

	return false
}

// Handles EventProcess events
func (eh *EventLoop) handleProcessEvent(evt *EventProcess) {
	defer func() {
		eh.ui.sort()
		eh.ui.draw()
	}()

	for _, p := range eh.multiplexer.panes {
		if p.key == evt.Key {
			if p.dead && evt.Autostart {
				p.start()
			}
			return
		}
	}

	p := eh.multiplexer.addPane(&pane{
		key:      evt.Key,
		args:     evt.Cmd,
		env:      evt.Env,
		dir:      evt.Cwd,
		title:    evt.Title,
		killable: evt.Killable,
	})

	if evt.Autostart {
		p.start()
	}

	if !evt.Autostart {
		p.vt.Start(process.Command("echo", p.key+" has auto-start disabled, press enter to start."))
		p.dead = true
	}
}

// Handles mouse events for scrolling, selection, and process switching
func (eh *EventLoop) handleMouseEvent(evt *tcell.EventMouse) {
	const MOUSE_SCROLL_SPEED = 3

	if evt.Buttons()&tcell.WheelUp != 0 {
		eh.multiplexer.scrollUp(MOUSE_SCROLL_SPEED)
		return
	}

	if evt.Buttons()&tcell.WheelDown != 0 {
		eh.multiplexer.scrollDown(MOUSE_SCROLL_SPEED)
		return
	}

	// Mouse button released - complete any ongoing selection
	if evt.Buttons() == tcell.ButtonNone {
		if eh.ui.hasSelection() {
			eh.multiplexer.copy()
		}

		eh.ui.resetDragging()
		return
	}

	if evt.Buttons()&tcell.ButtonPrimary != 0 {
		x, y := evt.Position()

		// Click in sidebar - switch to selected process
		if eh.ui.isSidebarClick(x) {
			eh.ui.selectPaneByCoordinates(x, y)
			return
		}

		// Click in main terminal area - handle text selection
		if eh.ui.isTerminalClick(x) {
			eh.ui.handleSelection(x, y)
			return
		}
	}
}

// Handles terminal resize events
func (eh *EventLoop) handleResizeEvent(evt *tcell.EventResize) {
	eh.multiplexer.resize(evt.Size())
	eh.ui.draw()
	eh.ui.screen.Sync()
}

// Handles terminal redraw requests from the virtual terminal
func (eh *EventLoop) handleRedrawEvent(evt *tcellterm.EventRedraw) {
	selected := eh.ui.selectedPane()
	if selected != nil && selected.vt == evt.VT() {
		selected.vt.Draw()
		eh.ui.screen.Show()
	}
}

// Handles process termination events
func (eh *EventLoop) handleClosedEvent(evt *tcellterm.EventClosed) {
	for _, proc := range eh.multiplexer.panes {
		if proc.vt == evt.VT() {
			if !proc.dead {
				// Show exit message and mark process as dead
				proc.vt.Start(process.Command("echo", "\n[process exited]"))
				proc.dead = true

				// Exit focus mode if the closed process was selected
				if proc.key == eh.ui.selected {
					eh.ui.blur()
				}

				eh.ui.sort()
			}
		}
	}

	eh.ui.draw()
}

// Handles keyboard events for navigation and terminal interaction
func (eh *EventLoop) handleKeyEvent(evt *tcell.EventKey) {
	selected := eh.ui.selectedPane()
	PAGE_MOVE_SPEED := eh.ui.screenHeight/2 + 1

	switch evt.Key() {
	case 256: // Regular character keys
		switch evt.Rune() {
		case 'j': // Vi-style down navigation
			if !eh.ui.focused {
				eh.ui.move(1)
				return
			}

		case 'k': // Vi-style up navigation
			if !eh.ui.focused {
				eh.ui.move(-1)
				return
			}

		case 'x': // Kill selected process
			if selected != nil && selected.killable && !selected.dead && !eh.ui.focused {
				selected.kill()
			}
		}

	case tcell.KeyUp:
		if !eh.ui.focused {
			eh.ui.move(-1)
			return
		}

	case tcell.KeyDown:
		if !eh.ui.focused {
			eh.ui.move(1)
			return
		}

	case tcell.KeyCtrlU:
		if selected != nil {
			eh.multiplexer.scrollUp(PAGE_MOVE_SPEED)
			return
		}

	case tcell.KeyCtrlD:
		if selected != nil {
			eh.multiplexer.scrollDown(PAGE_MOVE_SPEED)
			return
		}

	case tcell.KeyEnter:
		// Copy selected text if there's an active selection
		if selected != nil && selected.vt.HasSelection() {
			eh.multiplexer.copy()
			selected.vt.ClearSelection()
			eh.ui.draw()
			return
		}

		// Reset scroll position when scrolled up
		if selected != nil && selected.isScrolling() && (eh.ui.focused || !selected.killable) {
			selected.scrollReset()
			eh.ui.draw()
			eh.ui.screen.Sync()
			return
		}

		// Enter focus mode or restart dead process
		if !eh.ui.focused {
			if selected != nil && selected.killable {
				if selected.dead {
					selected.start()
					eh.ui.sort()
					eh.ui.draw()
					return
				}
				eh.ui.focus()
			}
			return
		}

	case tcell.KeyCtrlC:
		// Exit multiplexer when not focused on a terminal
		if !eh.ui.focused {
			eh.ui.move(-99999) // Move to top
			pid := os.Getpid()
			process, _ := os.FindProcess(pid)
			process.Signal(syscall.SIGINT)
			return
		}

	case tcell.KeyCtrlZ: // Exit focus mode
		if eh.ui.focused {
			eh.ui.blur()
			return
		}
	}

	// Forward keyboard events to the focused terminal
	if selected != nil && eh.ui.focused && !selected.isScrolling() {
		selected.vt.HandleEvent(evt)
		eh.ui.draw()
	}
}
