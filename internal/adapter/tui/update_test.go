package tui

import (
	"image"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/flexphere/gaze/internal/domain"
	"github.com/flexphere/gaze/internal/usecase"
)

type mockRenderFrame struct {
	output         string
	minimapEnabled bool
	frame          int
	setFrameCnt    int
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

func (m *mockRenderFrame) SetFrame(index int) {
	m.frame = index
	m.setFrameCnt++
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

	// FitToWindow calculates zoom based on terminal and image size
	// With 800x600 image, 80x24 terminal, cellAspect=2.0:
	// zoomH = 800*24*2/(80*600) = 0.8, so zoom = 0.8
	if um.viewport.ZoomLevel > 1.001 {
		t.Errorf("FitToWindow should set zoom to fit level (<=1.0), got %f", um.viewport.ZoomLevel)
	}
	if um.viewport.IsZoomed() {
		t.Error("FitToWindow should not leave viewport in zoomed state")
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

func newAnimatedTestModel(frameCount int) Model {
	cfg := domain.DefaultConfig()
	frames := make([]domain.Frame, frameCount)
	for i := range frames {
		frames[i] = domain.Frame{
			Image: image.NewRGBA(image.Rect(0, 0, 800, 600)),
			Delay: 100 * time.Millisecond,
		}
	}
	img := domain.NewAnimatedImageEntity(frames, "anim.gif", "gif")

	vpCtrl := usecase.NewViewportControlUseCase()
	renderer := &mockRenderFrame{output: "rendered", minimapEnabled: cfg.Minimap.Enabled}

	m := NewModel(img, cfg, vpCtrl, renderer)
	m.viewport.SetTerminalSize(80, 24)
	m.ready = true
	return m
}

var spaceKey = tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}

func TestInit_Animation(t *testing.T) {
	tests := []struct {
		name    string
		model   Model
		wantCmd bool
	}{
		{name: "static image does not tick", model: newTestModel(), wantCmd: false},
		{name: "animated image starts ticking", model: newAnimatedTestModel(3), wantCmd: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.model.Init() != nil; got != tt.wantCmd {
				t.Errorf("Init() returned cmd = %v, want %v", got, tt.wantCmd)
			}
		})
	}
}

func TestUpdate_FrameTick(t *testing.T) {
	tests := []struct {
		name          string
		startFrame    int
		paused        bool
		msgGen        int
		wantFrame     int
		wantCmd       bool
		wantSetFrames int
	}{
		{name: "advances to next frame", startFrame: 0, wantFrame: 1, wantCmd: true, wantSetFrames: 1},
		{name: "wraps around after last frame", startFrame: 2, wantFrame: 0, wantCmd: true, wantSetFrames: 1},
		{name: "ignored while paused", startFrame: 1, paused: true, wantFrame: 1, wantCmd: false},
		{name: "stale generation is ignored", startFrame: 1, msgGen: -1, wantFrame: 1, wantCmd: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newAnimatedTestModel(3)
			m.frameIndex = tt.startFrame
			m.paused = tt.paused

			updated, cmd := m.Update(frameTickMsg{gen: m.animGen + tt.msgGen})
			um := updated.(Model)
			renderer := um.renderFrame.(*mockRenderFrame)

			if um.frameIndex != tt.wantFrame {
				t.Errorf("frameIndex = %d, want %d", um.frameIndex, tt.wantFrame)
			}
			if (cmd != nil) != tt.wantCmd {
				t.Errorf("returned cmd = %v, want %v", cmd != nil, tt.wantCmd)
			}
			if renderer.setFrameCnt != tt.wantSetFrames {
				t.Errorf("SetFrame called %d times, want %d", renderer.setFrameCnt, tt.wantSetFrames)
			}
			if tt.wantSetFrames > 0 && renderer.frame != tt.wantFrame {
				t.Errorf("SetFrame(%d), want SetFrame(%d)", renderer.frame, tt.wantFrame)
			}
		})
	}
}

func TestUpdate_ToggleAnimation(t *testing.T) {
	m := newAnimatedTestModel(3)

	// Pause
	updated, cmd := m.Update(spaceKey)
	um := updated.(Model)
	if !um.paused {
		t.Fatal("should be paused after first toggle")
	}
	if cmd != nil {
		t.Error("pausing should not schedule a tick")
	}
	if !strings.Contains(um.statusBar(), "PAUSED") {
		t.Error("status bar should show PAUSED while paused")
	}

	// Resume
	updated, cmd = um.Update(spaceKey)
	um = updated.(Model)
	if um.paused {
		t.Fatal("should be resumed after second toggle")
	}
	if cmd == nil {
		t.Error("resuming should schedule a tick")
	}
	if strings.Contains(um.statusBar(), "PAUSED") {
		t.Error("status bar should not show PAUSED while playing")
	}

	// The tick scheduled before pausing belongs to an old generation and must
	// not start a second tick chain.
	updated, cmd = um.Update(frameTickMsg{gen: m.animGen})
	um = updated.(Model)
	if cmd != nil {
		t.Error("tick from before the toggle should be ignored")
	}
	if um.frameIndex != 0 {
		t.Errorf("frameIndex = %d, want 0", um.frameIndex)
	}
}

func TestUpdate_ToggleAnimation_StaticImage(t *testing.T) {
	m := newTestModel()

	updated, cmd := m.Update(spaceKey)
	um := updated.(Model)
	if um.paused {
		t.Error("static image should not be paused")
	}
	if cmd != nil {
		t.Error("static image should not schedule a tick")
	}
}
