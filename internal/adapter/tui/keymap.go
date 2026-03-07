package tui

import (
	"github.com/charmbracelet/bubbles/key"

	"github.com/flexphere/gaze/internal/domain"
)

// KeyMap defines all key bindings for the application.
type KeyMap struct {
	PanUp       key.Binding
	PanDown     key.Binding
	PanLeft     key.Binding
	PanRight    key.Binding
	ZoomIn      key.Binding
	ZoomOut     key.Binding
	ResetView   key.Binding
	FitToWindow key.Binding
	Quit        key.Binding
}

// NewKeyMap creates a KeyMap from configuration.
func NewKeyMap(cfg domain.KeyBindingConfig) KeyMap {
	return KeyMap{
		PanUp: key.NewBinding(
			key.WithKeys(cfg.PanUp...),
			key.WithHelp("↑/k", "pan up"),
		),
		PanDown: key.NewBinding(
			key.WithKeys(cfg.PanDown...),
			key.WithHelp("↓/j", "pan down"),
		),
		PanLeft: key.NewBinding(
			key.WithKeys(cfg.PanLeft...),
			key.WithHelp("←/h", "pan left"),
		),
		PanRight: key.NewBinding(
			key.WithKeys(cfg.PanRight...),
			key.WithHelp("→/l", "pan right"),
		),
		ZoomIn: key.NewBinding(
			key.WithKeys(cfg.ZoomIn...),
			key.WithHelp("+/=", "zoom in"),
		),
		ZoomOut: key.NewBinding(
			key.WithKeys(cfg.ZoomOut...),
			key.WithHelp("-/_", "zoom out"),
		),
		ResetView: key.NewBinding(
			key.WithKeys(cfg.ResetView...),
			key.WithHelp("0/r", "reset view"),
		),
		FitToWindow: key.NewBinding(
			key.WithKeys(cfg.FitToWindow...),
			key.WithHelp("f", "fit to window"),
		),
		Quit: key.NewBinding(
			key.WithKeys(cfg.Quit...),
			key.WithHelp("q", "quit"),
		),
	}
}
