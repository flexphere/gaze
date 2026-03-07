package main

import (
	"testing"

	"github.com/flexphere/gaze/internal/adapter/renderer"
)

func TestSelectRenderer(t *testing.T) {
	tests := []struct {
		name     string
		flag     string
		wantType string
	}{
		{"kitty flag", "kitty", "*renderer.KittyRenderer"},
		{"sixel flag", "sixel", "*renderer.SixelRenderer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := selectRenderer(tt.flag)
			switch tt.flag {
			case "kitty":
				if _, ok := r.(*renderer.KittyRenderer); !ok {
					t.Errorf("expected KittyRenderer, got %T", r)
				}
			case "sixel":
				if _, ok := r.(*renderer.SixelRenderer); !ok {
					t.Errorf("expected SixelRenderer, got %T", r)
				}
			}
		})
	}
}

func TestIsTmux(t *testing.T) {
	tests := []struct {
		name   string
		tmux   string
		term   string
		expect bool
	}{
		{"tmux env set", "/tmp/tmux-1000/default,12345,0", "xterm-256color", true},
		{"screen term", "", "screen", true},
		{"screen-256color", "", "screen-256color", true},
		{"tmux term", "", "tmux-256color", true},
		{"kitty term", "", "xterm-kitty", false},
		{"no tmux", "", "xterm-256color", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TMUX", tt.tmux)
			t.Setenv("TERM", tt.term)

			got := isTmux()
			if got != tt.expect {
				t.Errorf("isTmux() = %v, want %v", got, tt.expect)
			}
		})
	}
}
