package domain

// KeyBindingConfig holds the key strings for each action.
type KeyBindingConfig struct {
	PanUp         []string
	PanDown       []string
	PanLeft       []string
	PanRight      []string
	ZoomIn        []string
	ZoomOut       []string
	ResetView     []string
	FitToWindow   []string
	ToggleMinimap []string
	Quit          []string
}

// MouseConfig holds mouse behavior settings.
type MouseConfig struct {
	DragToPan         bool
	ScrollToZoom      bool
	ScrollSensitivity float64
}

// ViewportConfig holds viewport behavior settings.
type ViewportConfig struct {
	ZoomStep float64
	PanStep  float64
	MinZoom  float64
	MaxZoom  float64
}

// MinimapConfig holds minimap display settings.
type MinimapConfig struct {
	Enabled     bool
	Size        float64 // fraction of terminal width (0.0–1.0)
	BorderColor string  // hex color for indicator border (e.g. "#FFFFFF")
}

// Config is the complete application configuration.
type Config struct {
	KeyBindings KeyBindingConfig
	Mouse       MouseConfig
	Viewport    ViewportConfig
	Minimap     MinimapConfig
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		KeyBindings: KeyBindingConfig{
			PanUp:         []string{"k", "up"},
			PanDown:       []string{"j", "down"},
			PanLeft:       []string{"h", "left"},
			PanRight:      []string{"l", "right"},
			ZoomIn:        []string{"+", "="},
			ZoomOut:       []string{"-", "_"},
			ResetView:     []string{"0", "r"},
			FitToWindow:   []string{"f"},
			ToggleMinimap: []string{"m"},
			Quit:          []string{"q", "ctrl+c", "esc"},
		},
		Mouse: MouseConfig{
			DragToPan:         true,
			ScrollToZoom:      true,
			ScrollSensitivity: 0.15,
		},
		Viewport: ViewportConfig{
			ZoomStep: 0.1,
			PanStep:  0.05,
			MinZoom:  0.1,
			MaxZoom:  20.0,
		},
		Minimap: MinimapConfig{
			Enabled:     true,
			Size:        0.2,
			BorderColor: "#FFFFFF",
		},
	}
}
