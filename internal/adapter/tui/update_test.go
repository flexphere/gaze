package tui

import (
	"image"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/flexphere/gaze/internal/domain"
	"github.com/flexphere/gaze/internal/usecase"
)

type mockRenderFrame struct {
	output         string
	minimapEnabled bool
}

func (m *mockRenderFrame) Execute(_ *domain.ImageEntity, _ *domain.Viewport) (string, error) {
	return m.output, nil
}

func (m *mockRenderFrame) SetMinimapEnabled(enabled bool) {
	m.minimapEnabled = enabled
}

func (m *mockRenderFrame) MinimapEnabled() bool {
	return m.minimapEnabled
}

func newTestModel() Model {
	cfg := domain.DefaultConfig()
	img := domain.NewImageEntity(
		image.NewRGBA(image.Rect(0, 0, 800, 600)),
		"test.png",
		"png",
	)

	vpCtrl := usecase.NewViewportControlUseCase()
	renderer := &mockRenderFrame{output: "rendered", minimapEnabled: cfg.Minimap.Enabled}

	m := NewModel(img, cfg, vpCtrl, renderer)
	// Simulate window size message
	m.viewport.SetTerminalSize(80, 24)
	m.ready = true
	return m
}

func TestUpdate_QuitKey(t *testing.T) {
	m := newTestModel()
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}

	_, cmd := m.Update(msg)
	if cmd == nil {
		t.Error("expected tea.Quit command")
	}
}

func TestUpdate_ZoomIn(t *testing.T) {
	m := newTestModel()
	before := m.viewport.ZoomLevel
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}}

	updated, _ := m.Update(msg)
	um := updated.(Model)

	if um.viewport.ZoomLevel <= before {
		t.Errorf("ZoomIn should increase zoom, got %f", um.viewport.ZoomLevel)
	}
}

func TestUpdate_ZoomOut(t *testing.T) {
	m := newTestModel()
	m.viewport.ZoomLevel = 5.0
	before := m.viewport.ZoomLevel
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'-'}}

	updated, _ := m.Update(msg)
	um := updated.(Model)

	if um.viewport.ZoomLevel >= before {
		t.Errorf("ZoomOut should decrease zoom, got %f", um.viewport.ZoomLevel)
	}
}

func TestUpdate_PanDirections(t *testing.T) {
	tests := []struct {
		name string
		key  rune
	}{
		{"PanUp", 'k'},
		{"PanDown", 'j'},
		{"PanLeft", 'h'},
		{"PanRight", 'l'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel()
			m.viewport.ZoomLevel = 2.0
			m.viewport.OffsetX = 200
			m.viewport.OffsetY = 150
			m.viewport.Clamp()

			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{tt.key}}
			_, _ = m.Update(msg)
			// No panic = success
		})
	}
}

func TestUpdate_FitToWindow(t *testing.T) {
	m := newTestModel()
	m.viewport.ZoomLevel = 5.0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}}
	updated, _ := m.Update(msg)
	um := updated.(Model)

	if um.viewport.ZoomLevel != 1.0 {
		t.Errorf("FitToWindow should set zoom to 1.0, got %f", um.viewport.ZoomLevel)
	}
}

func TestUpdate_WindowSizeMsg(t *testing.T) {
	m := newTestModel()
	m.ready = false

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(msg)
	um := updated.(Model)

	if um.viewport.TermWidth != 120 {
		t.Errorf("TermWidth = %d, want 120", um.viewport.TermWidth)
	}
	if um.viewport.TermHeight != 39 { // -1 for status bar
		t.Errorf("TermHeight = %d, want 39", um.viewport.TermHeight)
	}
	if !um.ready {
		t.Error("should be ready after WindowSizeMsg")
	}
}

func TestUpdate_MouseWheelZoom(t *testing.T) {
	m := newTestModel()
	before := m.viewport.ZoomLevel

	msg := tea.MouseMsg{
		X:      40,
		Y:      12,
		Button: tea.MouseButtonWheelUp,
		Action: tea.MouseActionPress,
	}

	updated, _ := m.Update(msg)
	um := updated.(Model)

	if um.viewport.ZoomLevel <= before {
		t.Errorf("Wheel up should zoom in, got %f", um.viewport.ZoomLevel)
	}
}

func TestUpdate_MouseDrag(t *testing.T) {
	m := newTestModel()
	m.viewport.ZoomLevel = 2.0
	m.viewport.OffsetX = 200
	m.viewport.OffsetY = 150
	m.viewport.Clamp()

	// Press
	press := tea.MouseMsg{
		X: 40, Y: 12,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	}
	updated, _ := m.Update(press)
	um := updated.(Model)

	if !um.dragging {
		t.Error("should be dragging after press")
	}

	// Motion
	motion := tea.MouseMsg{
		X: 45, Y: 15,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionMotion,
	}
	updated, _ = um.Update(motion)
	um = updated.(Model)

	// Release
	release := tea.MouseMsg{
		X: 45, Y: 15,
		Button: tea.MouseButtonNone,
		Action: tea.MouseActionRelease,
	}
	updated, _ = um.Update(release)
	um = updated.(Model)

	if um.dragging {
		t.Error("should not be dragging after release")
	}
}

func TestUpdate_ToggleMinimap(t *testing.T) {
	m := newTestModel()
	renderer := m.renderFrame.(*mockRenderFrame)

	if !renderer.minimapEnabled {
		t.Fatal("minimap should be enabled by default")
	}

	// Toggle off
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	updated, _ := m.Update(msg)
	um := updated.(Model)
	renderer = um.renderFrame.(*mockRenderFrame)

	if renderer.minimapEnabled {
		t.Error("minimap should be disabled after toggle")
	}

	// Toggle on
	updated, _ = um.Update(msg)
	um = updated.(Model)
	renderer = um.renderFrame.(*mockRenderFrame)

	if !renderer.minimapEnabled {
		t.Error("minimap should be re-enabled after second toggle")
	}
}

func TestKeyMap_NewKeyMap(t *testing.T) {
	cfg := domain.DefaultConfig().KeyBindings
	km := NewKeyMap(cfg)

	if !key.Matches(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}, km.Quit) {
		t.Error("q should match Quit binding")
	}
	if !key.Matches(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}}, km.ZoomIn) {
		t.Error("+ should match ZoomIn binding")
	}
}
