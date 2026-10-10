package tui

import (
	"fmt"
	"strings"
)

const minTermWidth = 20
const minTermHeight = 5

// Kitty image placements and the minimap's CSI cursor move shift the cursor,
// but Bubbletea's line-diff renderer assumes the cursor stays where its own
// text output leaves it. Saving and restoring the cursor around the frame
// keeps Bubbletea's row tracking correct so the status bar is not drawn on
// unexpected rows.
const (
	saveCursor    = "\x1b7" // DECSC
	restoreCursor = "\x1b8" // DECRC
)

// View renders the current state to a string.
func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\nPress q to quit.\n", m.err)
	}

	if !m.ready {
		return "Loading..."
	}

	if m.viewport.TermWidth < minTermWidth || m.viewport.TermHeight < minTermHeight {
		return "Terminal too small. Please resize."
	}

	// Pad with newlines so the status bar lands on the last terminal row
	// (TermHeight excludes the status bar row).
	return saveCursor + m.renderedFrame + restoreCursor +
		strings.Repeat("\n", m.viewport.TermHeight) +
		m.statusBar()
}
