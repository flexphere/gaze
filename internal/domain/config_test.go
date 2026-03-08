package domain

import "testing"

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	t.Run("keybindings are populated", func(t *testing.T) {
		if len(cfg.KeyBindings.PanUp) == 0 {
			t.Error("PanUp keys should not be empty")
		}
		if len(cfg.KeyBindings.Quit) == 0 {
			t.Error("Quit keys should not be empty")
		}
		if len(cfg.KeyBindings.ZoomIn) == 0 {
			t.Error("ZoomIn keys should not be empty")
		}
	})

	t.Run("mouse defaults", func(t *testing.T) {
		if !cfg.Mouse.DragToPan {
			t.Error("DragToPan should be true by default")
		}
		if !cfg.Mouse.ScrollToZoom {
			t.Error("ScrollToZoom should be true by default")
		}
		if cfg.Mouse.ScrollSensitivity <= 0 {
			t.Error("ScrollSensitivity should be positive")
		}
	})

	t.Run("viewport defaults", func(t *testing.T) {
		if cfg.Viewport.ZoomStep <= 0 {
			t.Error("ZoomStep should be positive")
		}
		if cfg.Viewport.PanStep <= 0 {
			t.Error("PanStep should be positive")
		}
		if cfg.Viewport.MinZoom <= 0 {
			t.Error("MinZoom should be positive")
		}
		if cfg.Viewport.MaxZoom <= cfg.Viewport.MinZoom {
			t.Error("MaxZoom should be greater than MinZoom")
		}
	})

	t.Run("minimap defaults", func(t *testing.T) {
		if !cfg.Minimap.Enabled {
			t.Error("Minimap should be enabled by default")
		}
		if cfg.Minimap.Size <= 0 || cfg.Minimap.Size > 1.0 {
			t.Errorf("Minimap.Size should be between 0 and 1, got %f", cfg.Minimap.Size)
		}
		if cfg.Minimap.BorderColor == "" {
			t.Error("Minimap.BorderColor should have a default value")
		}
	})
}
