package multiplexer

import (
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
)

// Layout constants defining the UI structure.
const PADDING_WIDTH = 0
const PADDING_HEIGHT = 0
const SIDEBAR_WIDTH = 20

// UI handles all user interface rendering and state management
type UI struct {
	// State
	panes        []*pane
	focused      bool   // true when user is interacting with the active terminal
	selected     string // key of currently selected process
	screen       tcell.Screen
	screenWidth  int
	screenHeight int

	// Mouse interaction state for text selection
	dragging bool              // true during text selection drag operation
	click    *tcell.EventMouse // stores last click for double-click detection

	// UI elements
	sidebarView    *views.ViewPort
	activePaneView *views.ViewPort
	menuBox        *views.BoxLayout
	sidebarWidget  *PaneListWidget
	hotkeysWidget  *HotkeysWidget
}

// NewUI creates a new UI instance
func NewUI(screen tcell.Screen) *UI {
	activePane := views.NewViewPort(screen, 0, 0, 0, 0)
	sidebar := views.NewViewPort(screen, 0, 0, 0, 0)
	menu := views.NewBoxLayout(views.Vertical)
	menu.SetView(sidebar)

	ui := &UI{
		screen:         screen,
		screenWidth:    0,
		screenHeight:   0,
		sidebarView:    sidebar,
		activePaneView: activePane,
		menuBox:        menu,
		sidebarWidget:  NewPaneList(menu),
		hotkeysWidget:  NewHotkeysWidget(menu),
	}

	return ui
}

// Initializes the screen and prepares for drawing
func (ui *UI) start() {
	ui.resize(ui.screen.Size())
}

// Finalizes the screen and releasing the resources
func (ui *UI) stop() {
	ui.screen.Fini()
}

// Recalculates and updates viewport dimensions when the terminal is resized
func (ui *UI) resize(width int, height int) {
	ui.screenWidth = width
	ui.screenHeight = height

	ui.sidebarView.Resize(PADDING_WIDTH, PADDING_HEIGHT, SIDEBAR_WIDTH, height-PADDING_HEIGHT*2)
	ui.activePaneView.Resize(PADDING_WIDTH+SIDEBAR_WIDTH+PADDING_WIDTH+1, PADDING_HEIGHT, width-PADDING_WIDTH-SIDEBAR_WIDTH-PADDING_WIDTH-PADDING_WIDTH-1, height-PADDING_HEIGHT*2)

	mw, mh := ui.activePaneView.Size()
	for _, p := range ui.panes {
		p.vt.Resize(mw, mh)
	}
}

// Changes the selected process by the given offset
func (ui *UI) move(offset int) {
	index := max(ui.selectedPaneIndex()+offset, 0)
	if index >= len(ui.panes) {
		index = len(ui.panes) - 1
	}
	ui.selected = ui.panes[index].key
	ui.draw()
}

// Enters focus mode
func (ui *UI) focus() {
	ui.focused = true
	ui.draw()
}

// Exits focus mode
func (ui *UI) blur() {
	ui.focused = false
	selected := ui.selectedPane()
	if selected != nil {
		selected.scrollReset()
	}
	ui.screen.HideCursor()
	ui.draw()
}

// Returns the currently selected pane
func (ui *UI) selectedPane() *pane {
	for _, p := range ui.panes {
		if p.key == ui.selected {
			return p
		}
	}

	return nil
}

// Returns the currently selected pane index
func (ui *UI) selectedPaneIndex() int {
	for i, p := range ui.panes {
		if p.key == ui.selected {
			return i
		}
	}

	return -1
}

// Renders the entire UI including sidebar, hotkeys, and active terminal
func (ui *UI) draw() {
	defer ui.screen.Show()
	selected := ui.selectedPane()

	// Clear existing widgets
	for _, w := range ui.menuBox.Widgets() {
		ui.menuBox.RemoveWidget(w)
	}

	// Render sidebar with process list
	ui.sidebarWidget.render(ui.panes, selected, ui.focused)

	// Add spacer between sidebar and hotkeys
	ui.menuBox.AddWidget(views.NewSpacer(), 1)

	// Render hotkeys
	ui.hotkeysWidget.render(selected, ui.focused)

	// Draw the menu (sidebar + hotkeys)
	ui.menuBox.Draw()

	// Draw border between sidebar and main area
	borderStyle := tcell.StyleDefault.Foreground(tcell.ColorGray)
	for i := 0; i < ui.screenHeight; i++ {
		ui.screen.SetContent(SIDEBAR_WIDTH-1, i, '│', nil, borderStyle)
	}

	// Render virtual terminal
	if selected != nil {
		selected.vt.Draw()
		if ui.focused {
			y, x, _, _ := selected.vt.Cursor()
			ui.screen.ShowCursor(SIDEBAR_WIDTH+1+x, y+PADDING_HEIGHT)
		}
		if !ui.focused {
			ui.screen.HideCursor()
		}
	}
}

func (ui *UI) addPane(p *pane) {
	ui.panes = append(ui.panes, p)
	ui.sort()

	if len(ui.panes) == 1 {
		ui.selected = p.key
	}
}

// Sorts the panes and updates the selected index
func (ui *UI) sort() {
	if len(ui.panes) == 0 {
		return
	}

	sort.Slice(ui.panes, func(i, j int) bool {
		if ui.panes[i].killable && !ui.panes[j].killable {
			return false
		}
		if !ui.panes[i].dead && ui.panes[j].dead {
			return true
		}
		if ui.panes[i].dead && !ui.panes[j].dead {
			return false
		}
		return len(ui.panes[i].title) < len(ui.panes[j].title)
	})
}

func (ui *UI) hasSelection() bool {
	// if eh.ui.dragging && selected != nil {
	// 	eh.multiplexer.copy()
	// }
	return false
}

func (ui *UI) resetDragging() {
	// eh.ui.dragging = false
}

func (ui *UI) isSidebarClick(x int) bool {
	return x < SIDEBAR_WIDTH && !ui.dragging
}

func (ui *UI) isTerminalClick(x int) bool {
	return x > SIDEBAR_WIDTH
}

func (ui *UI) selectPaneByCoordinates(x int, y int) {
	// alive := 0
	// for _, p := range eh.multiplexer.panes {
	// 	if !p.dead {
	// 		alive++
	// 	}
	// }
	// // Adjust y coordinate if there's a separator between alive and dead processes
	// if alive != len(eh.multiplexer.panes) {
	// 	if y == alive {
	// 		return
	// 	}
	// 	if y > alive {
	// 		y--
	// 	}
	// }
	// if y >= len(eh.multiplexer.panes) {
	// 	return
	// }
	// eh.ui.selected = y
	// eh.ui.Blur()
}

func (ui *UI) handleSelection(x int, y int) {
	// Double-click detection for line selection
	// if !eh.ui.dragging && eh.ui.click != nil && time.Since(eh.ui.click.When()) < time.Millisecond*500 {
	// 	oldX, oldY := eh.ui.click.Position()
	// 	if oldX == x && oldY == y {
	// 		selected.vt.SelectStart(0, y)
	// 		selected.vt.SelectEnd(eh.ui.screenWidth-1, y)
	// 		eh.ui.dragging = true
	// 		eh.ui.Draw()
	// 		return
	// 	}
	// }
	// eh.ui.click = evt
	// offsetX := x - SIDEBAR_WIDTH - 1
	// if eh.ui.dragging {
	// 	selected.vt.SelectEnd(offsetX, y)
	// }
	// if !eh.ui.dragging {
	// 	eh.ui.dragging = true
	// 	selected.vt.SelectStart(offsetX, y)
	// }
	// eh.ui.draw()
}
