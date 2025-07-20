package multiplexer

import (
	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
)

// Widget for displaying the process list
type PaneListWidget struct {
	box *views.BoxLayout
}

// Creates a new sidebar widget
func NewPaneList(box *views.BoxLayout) *PaneListWidget {
	return &PaneListWidget{
		box: box,
	}
}

// Draws the process list in the sidebar
func (s *PaneListWidget) render(panes []*pane, selected *pane, focused bool) {
	for index, item := range panes {
		// Add separator between alive and dead processes
		if index > 0 && !panes[index-1].dead && item.dead {
			spacer := views.NewTextBar()
			spacer.SetLeft("──────────────────────", tcell.StyleDefault.Foreground(tcell.ColorGray))
			s.box.AddWidget(spacer, 0)
		}

		style := tcell.StyleDefault
		if item.dead {
			style = style.Foreground(tcell.ColorGray)
		}
		if item.key == selected.key {
			style = style.Bold(true)
			if !focused {
				style = style.Foreground(tcell.ColorOrange)
			}
		}

		title := views.NewTextBar()
		title.SetStyle(style)
		title.SetLeft(" "+item.icon+" "+item.title, tcell.StyleDefault)
		s.box.AddWidget(title, 0)
	}
}
