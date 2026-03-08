package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"github.com/flexphere/gaze/internal/domain"
)

type tomlConfig struct {
	KeyBindings tomlKeyBindings `toml:"keybindings"`
	Mouse       tomlMouse       `toml:"mouse"`
	Viewport    tomlViewport    `toml:"viewport"`
	Minimap     tomlMinimap     `toml:"minimap"`
}

type tomlKeyBindings struct {
	PanUp         []string `toml:"pan_up"`
	PanDown       []string `toml:"pan_down"`
	PanLeft       []string `toml:"pan_left"`
	PanRight      []string `toml:"pan_right"`
	ZoomIn        []string `toml:"zoom_in"`
	ZoomOut       []string `toml:"zoom_out"`
	ResetView     []string `toml:"reset_view"`
	FitToWindow   []string `toml:"fit_to_window"`
	ToggleMinimap []string `toml:"toggle_minimap"`
	PlayPause     []string `toml:"play_pause"`
	Quit          []string `toml:"quit"`
}

type tomlMouse struct {
	DragToPan         *bool    `toml:"drag_to_pan"`
	ScrollToZoom      *bool    `toml:"scroll_to_zoom"`
	ScrollSensitivity *float64 `toml:"scroll_sensitivity"`
}

type tomlViewport struct {
	ZoomStep *float64 `toml:"zoom_step"`
	PanStep  *float64 `toml:"pan_step"`
	MinZoom  *float64 `toml:"min_zoom"`
	MaxZoom  *float64 `toml:"max_zoom"`
}

type tomlMinimap struct {
	Enabled     *bool    `toml:"enabled"`
	Size        *float64 `toml:"size"`
	BorderColor *string  `toml:"border_color"`
}

// TOMLLoader loads configuration from a TOML file.
type TOMLLoader struct {
	path string
}

// NewTOMLLoader creates a TOMLLoader that reads from ~/.config/gaze/config.toml.
func NewTOMLLoader() *TOMLLoader {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return &TOMLLoader{path: ""}
		}
		configDir = filepath.Join(home, ".config")
	}
	return &TOMLLoader{
		path: filepath.Join(configDir, "gaze", "config.toml"),
	}
}

// NewTOMLLoaderWithPath creates a TOMLLoader that reads from a specific path.
func NewTOMLLoaderWithPath(path string) *TOMLLoader {
	return &TOMLLoader{path: path}
}

// Load reads and parses the config file, merging with defaults.
func (l *TOMLLoader) Load() (*domain.Config, error) {
	cfg := Defaults()

	if l.path == "" {
		return cfg, nil
	}

	var tc tomlConfig
	_, err := toml.DecodeFile(l.path, &tc)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		// Check if the file simply doesn't exist (different OS error formats)
		if _, statErr := os.Stat(l.path); errors.Is(statErr, os.ErrNotExist) {
			return cfg, nil
		}
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	mergeConfig(cfg, &tc)
	return cfg, nil
}

func mergeConfig(cfg *domain.Config, tc *tomlConfig) {
	// KeyBindings: override only if non-empty
	if len(tc.KeyBindings.PanUp) > 0 {
		cfg.KeyBindings.PanUp = tc.KeyBindings.PanUp
	}
	if len(tc.KeyBindings.PanDown) > 0 {
		cfg.KeyBindings.PanDown = tc.KeyBindings.PanDown
	}
	if len(tc.KeyBindings.PanLeft) > 0 {
		cfg.KeyBindings.PanLeft = tc.KeyBindings.PanLeft
	}
	if len(tc.KeyBindings.PanRight) > 0 {
		cfg.KeyBindings.PanRight = tc.KeyBindings.PanRight
	}
	if len(tc.KeyBindings.ZoomIn) > 0 {
		cfg.KeyBindings.ZoomIn = tc.KeyBindings.ZoomIn
	}
	if len(tc.KeyBindings.ZoomOut) > 0 {
		cfg.KeyBindings.ZoomOut = tc.KeyBindings.ZoomOut
	}
	if len(tc.KeyBindings.ResetView) > 0 {
		cfg.KeyBindings.ResetView = tc.KeyBindings.ResetView
	}
	if len(tc.KeyBindings.FitToWindow) > 0 {
		cfg.KeyBindings.FitToWindow = tc.KeyBindings.FitToWindow
	}
	if len(tc.KeyBindings.ToggleMinimap) > 0 {
		cfg.KeyBindings.ToggleMinimap = tc.KeyBindings.ToggleMinimap
	}
	if len(tc.KeyBindings.PlayPause) > 0 {
		cfg.KeyBindings.PlayPause = tc.KeyBindings.PlayPause
	}
	if len(tc.KeyBindings.Quit) > 0 {
		cfg.KeyBindings.Quit = tc.KeyBindings.Quit
	}

	// Mouse: override only if explicitly set
	if tc.Mouse.DragToPan != nil {
		cfg.Mouse.DragToPan = *tc.Mouse.DragToPan
	}
	if tc.Mouse.ScrollToZoom != nil {
		cfg.Mouse.ScrollToZoom = *tc.Mouse.ScrollToZoom
	}
	if tc.Mouse.ScrollSensitivity != nil {
		cfg.Mouse.ScrollSensitivity = *tc.Mouse.ScrollSensitivity
	}

	// Viewport: override only if explicitly set
	if tc.Viewport.ZoomStep != nil {
		cfg.Viewport.ZoomStep = *tc.Viewport.ZoomStep
	}
	if tc.Viewport.PanStep != nil {
		cfg.Viewport.PanStep = *tc.Viewport.PanStep
	}
	if tc.Viewport.MinZoom != nil {
		cfg.Viewport.MinZoom = *tc.Viewport.MinZoom
	}
	if tc.Viewport.MaxZoom != nil {
		cfg.Viewport.MaxZoom = *tc.Viewport.MaxZoom
	}

	// Minimap: override only if explicitly set
	if tc.Minimap.Enabled != nil {
		cfg.Minimap.Enabled = *tc.Minimap.Enabled
	}
	if tc.Minimap.Size != nil {
		size := *tc.Minimap.Size
		if size > 0 && size <= 1 {
			cfg.Minimap.Size = size
		}
	}
	if tc.Minimap.BorderColor != nil {
		cfg.Minimap.BorderColor = *tc.Minimap.BorderColor
	}
}
