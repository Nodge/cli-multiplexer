package multiplexer

import (
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
)

// Widget for displaying available keyboard shortcuts
type HotkeysWidget struct {
	menu *views.BoxLayout
}

// Creates a new hotkeys widget
func NewHotkeysWidget(menu *views.BoxLayout) *HotkeysWidget {
	return &HotkeysWidget{
		menu: menu,
	}
}

// Generates and displays the hotkeys based on current state
func (h *HotkeysWidget) render(selected *pane, focused bool) {
	hotkeys := map[string]string{}

	if selected != nil && selected.killable && !focused {
		if !selected.dead {
			hotkeys["x"] = "kill"
			hotkeys["enter"] = "focus"
		}

		if selected.dead {
			hotkeys["enter"] = "start"
		}
	}

	if !focused {
		hotkeys["j/k/↓/↑"] = "up/down"
	}

	if focused {
		hotkeys["ctrl-z"] = "sidebar"
	}

	if selected != nil && selected.isScrolling() && (focused || !selected.killable) {
		hotkeys["enter"] = "reset"
	}

	if selected != nil && selected.vt.HasSelection() {
		hotkeys["enter"] = "copy"
	}

	hotkeys["ctrl-u/d"] = "scroll"

	// Sort hotkeys by length first, then alphabetically
	keys := make([]string, 0, len(hotkeys))
	for key := range hotkeys {
		keys = append(keys, key)
	}
	slices.SortFunc(keys, func(i, j string) int {
		ilength := utf8.RuneCountInString(i)
		jlength := utf8.RuneCountInString(j)
		if ilength != jlength {
			return ilength - jlength
		}
		return strings.Compare(i, j)
	})

	for _, key := range keys {
		label := hotkeys[key]
		title := views.NewTextBar()
		title.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorGray))
		title.SetLeft(" "+key, tcell.StyleDefault.Foreground(tcell.ColorGray).Bold(true))
		title.SetRight(label+"  ", tcell.StyleDefault)
		h.menu.AddWidget(title, 0)
	}
}
