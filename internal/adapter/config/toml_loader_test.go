package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTOMLLoader_Load_MissingFile(t *testing.T) {
	loader := NewTOMLLoaderWithPath("/nonexistent/config.toml")
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return defaults
	if len(cfg.KeyBindings.PanUp) == 0 {
		t.Error("PanUp should have default keys")
	}
	if cfg.Viewport.ZoomStep != 0.1 {
		t.Errorf("ZoomStep = %f, want 0.1", cfg.Viewport.ZoomStep)
	}
}

func TestTOMLLoader_Load_FullConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	content := `
[keybindings]
pan_up = ["w"]
pan_down = ["s"]
zoom_in = ["z"]
quit = ["x"]

[mouse]
drag_to_pan = false
scroll_sensitivity = 0.25

[viewport]
zoom_step = 0.2
pan_step = 0.1
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing config: %v", err)
	}

	loader := NewTOMLLoaderWithPath(path)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Custom keys
	if len(cfg.KeyBindings.PanUp) != 1 || cfg.KeyBindings.PanUp[0] != "w" {
		t.Errorf("PanUp = %v, want [w]", cfg.KeyBindings.PanUp)
	}
	if len(cfg.KeyBindings.Quit) != 1 || cfg.KeyBindings.Quit[0] != "x" {
		t.Errorf("Quit = %v, want [x]", cfg.KeyBindings.Quit)
	}

	// Defaults preserved for unset keys
	if len(cfg.KeyBindings.PanLeft) == 0 {
		t.Error("PanLeft should retain defaults")
	}

	// Mouse overrides
	if cfg.Mouse.DragToPan {
		t.Error("DragToPan should be false")
	}
	if cfg.Mouse.ScrollSensitivity != 0.25 {
		t.Errorf("ScrollSensitivity = %f, want 0.25", cfg.Mouse.ScrollSensitivity)
	}
	// ScrollToZoom should retain default
	if !cfg.Mouse.ScrollToZoom {
		t.Error("ScrollToZoom should retain default true")
	}

	// Viewport overrides
	if cfg.Viewport.ZoomStep != 0.2 {
		t.Errorf("ZoomStep = %f, want 0.2", cfg.Viewport.ZoomStep)
	}
	if cfg.Viewport.PanStep != 0.1 {
		t.Errorf("PanStep = %f, want 0.1", cfg.Viewport.PanStep)
	}
}

func TestTOMLLoader_Load_PartialConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	content := `
[keybindings]
quit = ["q"]
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing config: %v", err)
	}

	loader := NewTOMLLoaderWithPath(path)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only quit should be overridden
	if len(cfg.KeyBindings.Quit) != 1 || cfg.KeyBindings.Quit[0] != "q" {
		t.Errorf("Quit = %v, want [q]", cfg.KeyBindings.Quit)
	}

	// Others should be defaults
	if len(cfg.KeyBindings.PanUp) != 2 {
		t.Errorf("PanUp should have 2 default keys, got %d", len(cfg.KeyBindings.PanUp))
	}
}

func TestTOMLLoader_Load_InvalidTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	if err := os.WriteFile(path, []byte("invalid = [[["), 0o644); err != nil {
		t.Fatalf("writing config: %v", err)
	}

	loader := NewTOMLLoaderWithPath(path)
	_, err := loader.Load()
	if err == nil {
		t.Fatal("expected error for invalid TOML, got nil")
	}
}

func TestTOMLLoader_Load_EmptyPath(t *testing.T) {
	loader := &TOMLLoader{path: ""}
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("config should not be nil")
	}
}

func TestTOMLLoader_Load_MinimapConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	content := `
[minimap]
enabled = false
size = 0.3
border_color = "#00FF00"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing config: %v", err)
	}

	loader := NewTOMLLoaderWithPath(path)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Minimap.Enabled {
		t.Error("Minimap.Enabled should be false")
	}
	if cfg.Minimap.Size != 0.3 {
		t.Errorf("Minimap.Size = %f, want 0.3", cfg.Minimap.Size)
	}
	if cfg.Minimap.BorderColor != "#00FF00" {
		t.Errorf("Minimap.BorderColor = %q, want %q", cfg.Minimap.BorderColor, "#00FF00")
	}
}

func TestTOMLLoader_Load_MinimapSizeValidation(t *testing.T) {
	tests := []struct {
		name     string
		size     string
		wantSize float64
	}{
		{"valid", "0.3", 0.3},
		{"zero ignored", "0.0", 0.2},      // default
		{"negative ignored", "-0.5", 0.2}, // default
		{"too large ignored", "1.5", 0.2}, // default
		{"max valid", "1.0", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "config.toml")

			content := "[minimap]\nsize = " + tt.size + "\n"
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				t.Fatalf("writing config: %v", err)
			}

			loader := NewTOMLLoaderWithPath(path)
			cfg, err := loader.Load()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Minimap.Size != tt.wantSize {
				t.Errorf("Minimap.Size = %f, want %f", cfg.Minimap.Size, tt.wantSize)
			}
		})
	}
}
