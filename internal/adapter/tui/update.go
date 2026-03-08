package tui

import (
	"errors"
	"io"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/flexphere/gaze/internal/domain"
)

const seekStep = 5 * time.Second

type videoTickMsg struct{}

// Init returns the initial command.
func (m Model) Init() tea.Cmd {
	if m.isVideoMode() && m.playing {
		return m.tickCmd()
	}
	return nil
}

// Update handles messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.SetTerminalSize(msg.Width, msg.Height-1) // -1 for status bar
		m.ready = true
		m.updateFrame()
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tea.MouseMsg:
		if !m.isVideoMode() {
			return m.handleMouseMsg(msg)
		}
		return m, nil

	case videoTickMsg:
		return m.handleVideoTick()
	}

	return m, nil
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, m.keymap.Quit) {
		return m, tea.Quit
	}

	if m.isVideoMode() {
		return m.handleVideoKeyMsg(msg)
	}
	return m.handleImageKeyMsg(msg)
}

func (m Model) handleImageKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
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

func (m Model) handleVideoKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keymap.PlayPause):
		m.playing = !m.playing
		if m.playing {
			return m, m.tickCmd()
		}
		return m, nil

	case key.Matches(msg, m.keymap.PanLeft):
		return m.seekVideo(-seekStep)

	case key.Matches(msg, m.keymap.PanRight):
		return m.seekVideo(seekStep)
	}

	return m, nil
}

func (m Model) seekVideo(delta time.Duration) (tea.Model, tea.Cmd) {
	newPos := m.position + delta
	if newPos < 0 {
		newPos = 0
	}
	if m.videoInfo.Duration > 0 && newPos > m.videoInfo.Duration {
		newPos = m.videoInfo.Duration
	}

	if err := m.videoDecoder.Seek(newPos); err != nil {
		m.err = err
		return m, nil
	}
	m.position = newPos

	frame, err := m.videoDecoder.NextFrame()
	if err != nil {
		if !errors.Is(err, io.EOF) {
			m.err = err
		}
		m.playing = false
		return m, nil
	}

	m.image = domain.NewImageEntity(frame, m.image.Path, "video")
	m.updateFrame()

	if m.playing {
		return m, m.tickCmd()
	}
	return m, nil
}

func (m Model) handleVideoTick() (tea.Model, tea.Cmd) {
	if !m.playing || m.videoDecoder == nil {
		return m, nil
	}

	frame, err := m.videoDecoder.NextFrame()
	if err != nil {
		if !errors.Is(err, io.EOF) {
			m.err = err
		}
		m.playing = false
		return m, nil
	}

	m.image = domain.NewImageEntity(frame, m.image.Path, "video")
	m.position += time.Duration(float64(time.Second) / m.videoInfo.FrameRate)
	m.updateFrame()

	return m, m.tickCmd()
}

func (m Model) tickCmd() tea.Cmd {
	interval := time.Duration(float64(time.Second) / m.videoInfo.FrameRate)
	return tea.Tick(interval, func(time.Time) tea.Msg {
		return videoTickMsg{}
	})
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
