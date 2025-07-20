package multiplexer

import (
	"os/exec"

	"github.com/nodge/multiplexer/internal/process"
	tcellterm "github.com/nodge/multiplexer/internal/tcell-term"
)

// Pane represents a terminal process with its associated state and virtual terminal
type pane struct {
	key      string
	title    string
	icon     string
	dir      string
	cmd      *exec.Cmd
	args     []string
	env      []string
	killable bool
	vt       *tcellterm.VT
	dead     bool
}

// Initializes and starts the terminal process for this pane
func (p *pane) start() error {
	p.cmd = process.Command(p.args[0], p.args[1:]...)
	p.cmd.Env = p.env
	if p.dir != "" {
		p.cmd.Dir = p.dir
	}

	p.vt.Clear()

	err := p.vt.Start(p.cmd)
	if err != nil {
		return err
	}

	p.dead = false

	return nil
}

// Terminates the terminal process for this pane
func (p *pane) kill() {
	p.vt.Close()
}

// Scrolls the terminal view up by the specified offset
func (p *pane) scrollUp(offset int) {
	p.vt.ScrollUp(offset)
}

// Scrolls the terminal view down by the specified offset
func (p *pane) scrollDown(offset int) {
	p.vt.ScrollDown(offset)
}

// Resets the terminal scroll position to the bottom
func (p *pane) scrollReset() {
	p.vt.ScrollReset()
}

// Returns true if the terminal is currently scrolled up from the bottom
func (p *pane) isScrolling() bool {
	return p.vt.IsScrolling()
}
