package renderer

import (
	"strings"
	"testing"

	"github.com/flexphere/gaze/internal/domain"
)

func TestKittyRenderer_Display(t *testing.T) {
	r := NewKittyRenderer()
	r.imageID = 1
	r.imgW = 800
	r.imgH = 600

	vp := domain.NewViewport(domain.ViewportConfig{
		ZoomStep: 0.1, PanStep: 0.05, MinZoom: 0.1, MaxZoom: 20.0,
	})
	vp.ImgWidth = 800
	vp.ImgHeight = 600
	vp.TermWidth = 80
	vp.TermHeight = 24

	output, err := r.Display(vp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "i=1") {
		t.Error("output should contain image ID")
	}
	if !strings.Contains(output, "a=p") {
		t.Error("output should contain action=place")
	}
	if !strings.Contains(output, "w=800") {
		t.Error("output should contain source width")
	}
	if !strings.Contains(output, "h=600") {
		t.Error("output should contain source height")
	}
}

func TestKittyRenderer_Display_ZoomedIn(t *testing.T) {
	r := NewKittyRenderer()
	r.imageID = 1
	r.imgW = 800
	r.imgH = 600

	vp := domain.NewViewport(domain.ViewportConfig{
		ZoomStep: 0.1, PanStep: 0.05, MinZoom: 0.1, MaxZoom: 20.0,
	})
	vp.ImgWidth = 800
	vp.ImgHeight = 600
	vp.TermWidth = 80
	vp.TermHeight = 24
	vp.ZoomLevel = 2.0 // Shows 400x300

	output, err := r.Display(vp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "w=400") {
		t.Errorf("output should contain w=400 for 2x zoom, got: %s", output)
	}
	if !strings.Contains(output, "h=300") {
		t.Errorf("output should contain h=300 for 2x zoom, got: %s", output)
	}
}

func TestKittyRenderer_Display_ZeroSize(t *testing.T) {
	r := NewKittyRenderer()
	r.imageID = 1

	vp := domain.NewViewport(domain.ViewportConfig{
		ZoomStep: 0.1, PanStep: 0.05, MinZoom: 0.1, MaxZoom: 20.0,
	})
	vp.TermWidth = 0
	vp.TermHeight = 0

	output, err := r.Display(vp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output != "" {
		t.Errorf("expected empty output for zero terminal size, got: %q", output)
	}
}

func TestKittyRenderer_Display_AspectRatio(t *testing.T) {
	r := NewKittyRenderer()
	r.imageID = 1
	r.imgW = 1920
	r.imgH = 1080

	vp := domain.NewViewport(domain.ViewportConfig{
		ZoomStep: 0.1, PanStep: 0.05, MinZoom: 0.1, MaxZoom: 20.0,
	})
	vp.ImgWidth = 1920
	vp.ImgHeight = 1080
	vp.TermWidth = 80
	vp.TermHeight = 24

	output, err := r.Display(vp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should contain display columns and rows
	if !strings.Contains(output, "c=") {
		t.Error("output should contain display columns")
	}
	if !strings.Contains(output, "r=") {
		t.Error("output should contain display rows")
	}
}
