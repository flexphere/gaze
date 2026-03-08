package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

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

	playStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82"))

	pauseStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))
)

func (m Model) statusBar() string {
	if m.isVideoMode() {
		return m.videoStatusBar()
	}
	return m.imageStatusBar()
}

func (m Model) imageStatusBar() string {
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

func (m Model) videoStatusBar() string {
	filename := filepath.Base(m.image.Path)
	dims := fmt.Sprintf("%dx%d", m.videoInfo.Width, m.videoInfo.Height)

	var state string
	if m.playing {
		state = playStyle.Render("PLAY")
	} else {
		state = pauseStyle.Render("PAUSE")
	}

	posStr := formatTime(m.position)
	durStr := formatTime(m.videoInfo.Duration)
	timeStr := dimStyle.Render(fmt.Sprintf("%s / %s", posStr, durStr))

	left := fileStyle.Render(filename) + " " +
		dimStyle.Render(dims) + " " +
		state

	right := timeStr

	gap := m.viewport.TermWidth - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}

	bar := left + strings.Repeat(" ", gap) + right
	return statusStyle.Render(bar)
}

func formatTime(d time.Duration) string {
	total := int(d.Seconds())
	m := total / 60
	s := total % 60
	return fmt.Sprintf("%d:%02d", m, s)
}
