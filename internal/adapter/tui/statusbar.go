package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	statusStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("252")).
			Padding(0, 1)

	fileStyle = lipgloss.NewStyle().
			Bold(true)

	zoomStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244"))
)

func (m Model) statusBar() string {
	filename := filepath.Base(m.image.Path)
	zoom := m.viewport.ZoomPercentage()
	dims := fmt.Sprintf("%dx%d", m.image.Width, m.image.Height)

	rect := m.viewport.VisibleRect()
	pos := fmt.Sprintf("(%d,%d)", rect.Min.X, rect.Min.Y)

	left := fileStyle.Render(filename) + " " +
		dimStyle.Render(dims) + " " +
		dimStyle.Render(strings.ToUpper(m.image.Format))

	right := dimStyle.Render(pos) + " " +
		zoomStyle.Render(fmt.Sprintf("%d%%", zoom))

	gap := m.viewport.TermWidth - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}

	bar := left + strings.Repeat(" ", gap) + right
	return statusStyle.Render(bar)
}
