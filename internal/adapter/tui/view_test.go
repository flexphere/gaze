package tui

import (
	"errors"
	"strings"
	"testing"
)

func TestView_StatusBarOnLastRow(t *testing.T) {
	tests := []struct {
		name       string
		frame      string
		termWidth  int
		termHeight int
	}{
		{"empty frame", "", 80, 23},
		{"image placement", "\x1b[H\x1b_Ga=p,i=1,c=80,r=5,q=2\x1b\\", 80, 23},
		{"image with minimap cursor move", "\x1b[H\x1b_Ga=p,i=1,q=2\x1b\\\x1b[18;60H\x1b_Ga=p,i=2,q=2\x1b\\", 80, 23},
		{"minimum terminal size", "\x1b[H", minTermWidth, minTermHeight},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel()
			m.viewport.SetTerminalSize(tt.termWidth, tt.termHeight)
			m.renderedFrame = tt.frame

			lines := strings.Split(m.View(), "\n")

			if got, want := len(lines), tt.termHeight+1; got != want {
				t.Fatalf("line count = %d, want %d", got, want)
			}
			if got, want := lines[0], saveCursor+tt.frame+restoreCursor; got != want {
				t.Errorf("first line = %q, want %q", got, want)
			}
			for i, l := range lines[1 : len(lines)-1] {
				if l != "" {
					t.Errorf("line %d = %q, want empty", i+1, l)
				}
			}
			if got, want := lines[len(lines)-1], m.statusBar(); got != want {
				t.Errorf("last line = %q, want status bar %q", got, want)
			}
		})
	}
}

func TestView_NonFrameStates(t *testing.T) {
	tests := []struct {
		name  string
		setup func(m *Model)
		want  string
	}{
		{"error", func(m *Model) { m.err = errors.New("boom") }, "Error: boom\nPress q to quit.\n"},
		{"not ready", func(m *Model) { m.ready = false }, "Loading..."},
		{"terminal too small", func(m *Model) { m.viewport.SetTerminalSize(minTermWidth-1, minTermHeight) }, "Terminal too small. Please resize."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel()
			tt.setup(&m)
			if got := m.View(); got != tt.want {
				t.Errorf("View() = %q, want %q", got, tt.want)
			}
		})
	}
}
