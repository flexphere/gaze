package domain

// KeyBindingConfig holds the key strings for each action.
type KeyBindingConfig struct {
	PanUp       []string
	PanDown     []string
	PanLeft     []string
	PanRight    []string
	ZoomIn      []string
	ZoomOut     []string
	ResetView   []string
	FitToWindow []string
	Quit        []string
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

// Config is the complete application configuration.
type Config struct {
	KeyBindings KeyBindingConfig
	Mouse       MouseConfig
	Viewport    ViewportConfig
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		KeyBindings: KeyBindingConfig{
			PanUp:       []string{"k", "up"},
			PanDown:     []string{"j", "down"},
			PanLeft:     []string{"h", "left"},
			PanRight:    []string{"l", "right"},
			ZoomIn:      []string{"+", "="},
			ZoomOut:     []string{"-", "_"},
			ResetView:   []string{"0", "r"},
			FitToWindow: []string{"f"},
			Quit:        []string{"q", "ctrl+c", "esc"},
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
	}
}
