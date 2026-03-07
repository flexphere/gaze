package usecase

import (
	"testing"

	"github.com/flexphere/gaze/internal/domain"
)

func newTestViewport() *domain.Viewport {
	vp := domain.NewViewport(domain.ViewportConfig{
		ZoomStep: 0.1,
		PanStep:  0.05,
		MinZoom:  0.1,
		MaxZoom:  20.0,
	})
	vp.TermWidth = 80
	vp.TermHeight = 24
	vp.ImgWidth = 1000
	vp.ImgHeight = 800
	return vp
}

func TestViewportControlUseCase_ZoomIn(t *testing.T) {
	uc := NewViewportControlUseCase()
	vp := newTestViewport()
	before := vp.ZoomLevel

	uc.ZoomIn(vp)

	if vp.ZoomLevel <= before {
		t.Errorf("ZoomIn should increase zoom, got %f (was %f)", vp.ZoomLevel, before)
	}
}

func TestViewportControlUseCase_ZoomOut(t *testing.T) {
	uc := NewViewportControlUseCase()
	vp := newTestViewport()
	vp.ZoomLevel = 5.0
	before := vp.ZoomLevel

	uc.ZoomOut(vp)

	if vp.ZoomLevel >= before {
		t.Errorf("ZoomOut should decrease zoom, got %f (was %f)", vp.ZoomLevel, before)
	}
}

func TestViewportControlUseCase_PanDirections(t *testing.T) {
	uc := NewViewportControlUseCase()

	tests := []struct {
		name     string
		action   func(vp *domain.Viewport)
		checkX   bool
		positive bool
	}{
		{"PanRight", uc.PanRight, true, true},
		{"PanLeft", uc.PanLeft, true, false},
		{"PanDown", uc.PanDown, false, true},
		{"PanUp", uc.PanUp, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vp := newTestViewport()
			vp.ZoomLevel = 2.0
			vp.OffsetX = 200
			vp.OffsetY = 200
			vp.Clamp()

			beforeX := vp.OffsetX
			beforeY := vp.OffsetY

			tt.action(vp)

			if tt.checkX {
				if tt.positive && vp.OffsetX <= beforeX {
					t.Errorf("%s should increase OffsetX", tt.name)
				}
				if !tt.positive && vp.OffsetX >= beforeX {
					t.Errorf("%s should decrease OffsetX", tt.name)
				}
			} else {
				if tt.positive && vp.OffsetY <= beforeY {
					t.Errorf("%s should increase OffsetY", tt.name)
				}
				if !tt.positive && vp.OffsetY >= beforeY {
					t.Errorf("%s should decrease OffsetY", tt.name)
				}
			}
		})
	}
}

func TestViewportControlUseCase_PanByPixels(t *testing.T) {
	uc := NewViewportControlUseCase()
	vp := newTestViewport()
	vp.ZoomLevel = 2.0

	uc.PanByPixels(vp, 50, 30)

	if vp.OffsetX != 50 {
		t.Errorf("OffsetX = %f, want 50", vp.OffsetX)
	}
	if vp.OffsetY != 30 {
		t.Errorf("OffsetY = %f, want 30", vp.OffsetY)
	}
}

func TestViewportControlUseCase_ResetView(t *testing.T) {
	uc := NewViewportControlUseCase()
	vp := newTestViewport()
	vp.ZoomLevel = 5.0
	vp.OffsetX = 300

	uc.ResetView(vp)

	if vp.ZoomLevel != 1.0 {
		t.Errorf("ZoomLevel = %f, want 1.0", vp.ZoomLevel)
	}
}

func TestViewportControlUseCase_FitToWindow(t *testing.T) {
	uc := NewViewportControlUseCase()
	vp := newTestViewport()
	vp.ZoomLevel = 5.0

	uc.FitToWindow(vp)

	if vp.ZoomLevel != 1.0 {
		t.Errorf("ZoomLevel = %f, want 1.0", vp.ZoomLevel)
	}
}

func TestViewportControlUseCase_ZoomAtPoint(t *testing.T) {
	uc := NewViewportControlUseCase()
	vp := newTestViewport()
	before := vp.ZoomLevel

	uc.ZoomAtPoint(vp, 0.1, 40, 12)

	if vp.ZoomLevel <= before {
		t.Errorf("ZoomAtPoint should increase zoom, got %f (was %f)", vp.ZoomLevel, before)
	}
}
