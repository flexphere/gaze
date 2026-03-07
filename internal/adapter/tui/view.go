package tui

import "fmt"

const minTermWidth = 20
const minTermHeight = 5

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

	return m.renderedFrame + "\n" + m.statusBar()
}
