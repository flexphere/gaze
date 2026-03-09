package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Init returns the initial command.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cellW, cellH := queryCellSize()
		m.viewport.SetCellAspectRatio(cellH / cellW)
		m.viewport.SetTerminalSize(msg.Width, msg.Height-1) // -1 for status bar
		m.ready = true
		m.updateFrame()
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tea.MouseMsg:
		return m.handleMouseMsg(msg)
	}

	return m, nil
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keymap.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.keymap.ZoomIn):
		m.vpCtrl.ZoomIn(m.viewport)
	case key.Matches(msg, m.keymap.ZoomOut):
		m.vpCtrl.ZoomOut(m.viewport)
	case key.Matches(msg, m.keymap.PanUp):
		m.vpCtrl.PanUp(m.viewport)
	case key.Matches(msg, m.keymap.PanDown):
		m.vpCtrl.PanDown(m.viewport)
	case key.Matches(msg, m.keymap.PanLeft):
		m.vpCtrl.PanLeft(m.viewport)
	case key.Matches(msg, m.keymap.PanRight):
		m.vpCtrl.PanRight(m.viewport)
	case key.Matches(msg, m.keymap.ResetView):
		m.vpCtrl.ResetView(m.viewport)
	case key.Matches(msg, m.keymap.FitToWindow):
		m.vpCtrl.FitToWindow(m.viewport)
	case key.Matches(msg, m.keymap.ToggleMinimap):
		m.renderFrame.SetMinimapEnabled(!m.renderFrame.MinimapEnabled())
	default:
		return m, nil
	}

	m.updateFrame()
	return m, nil
}

func (m Model) handleMouseMsg(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	switch msg.Button {
	case tea.MouseButtonLeft:
		switch msg.Action {
		case tea.MouseActionPress:
			m.dragging = true
			m.dragStartTermX = msg.X
			m.dragStartTermY = msg.Y
			m.dragStartOffX = m.viewport.OffsetX
			m.dragStartOffY = m.viewport.OffsetY
			return m, nil

		case tea.MouseActionMotion:
			if m.dragging && m.config.Mouse.DragToPan {
				dx := msg.X - m.dragStartTermX
				dy := msg.Y - m.dragStartTermY

				pxPerCellX := m.viewport.VisibleWidth() / float64(m.viewport.TermWidth)
				pxPerCellY := m.viewport.VisibleHeight() / float64(m.viewport.TermHeight)

				m.viewport.OffsetX = m.dragStartOffX - float64(dx)*pxPerCellX
				m.viewport.OffsetY = m.dragStartOffY - float64(dy)*pxPerCellY
				m.viewport.Clamp()
				m.updateFrame()
			}
			return m, nil
		}

	case tea.MouseButtonNone:
		if msg.Action == tea.MouseActionRelease {
			m.dragging = false
			return m, nil
		}

	case tea.MouseButtonWheelUp:
		if m.config.Mouse.ScrollToZoom {
			m.vpCtrl.ZoomAtPoint(m.viewport, m.config.Mouse.ScrollSensitivity, msg.X, msg.Y)
			m.updateFrame()
		}
		return m, nil

	case tea.MouseButtonWheelDown:
		if m.config.Mouse.ScrollToZoom {
			m.vpCtrl.ZoomAtPoint(m.viewport, -m.config.Mouse.ScrollSensitivity, msg.X, msg.Y)
			m.updateFrame()
		}
		return m, nil
	}

	return m, nil
}

func (m *Model) updateFrame() {
	if !m.ready {
		return
	}
	output, err := m.renderFrame.Execute(m.image, m.viewport)
	if err != nil {
		m.err = err
		return
	}
	m.renderedFrame = output
}
