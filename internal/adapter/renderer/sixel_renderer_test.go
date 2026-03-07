package renderer

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/flexphere/gaze/internal/domain"
)

func newTestImage(w, h int) *domain.ImageEntity {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 128, A: 255})
		}
	}
	return domain.NewImageEntity(img, "test.png", "png")
}

func newTestViewport() *domain.Viewport {
	vp := domain.NewViewport(domain.ViewportConfig{
		ZoomStep: 0.1, PanStep: 0.05, MinZoom: 0.1, MaxZoom: 20.0,
	})
	return vp
}

func TestSixelRenderer_Upload(t *testing.T) {
	r := NewSixelRenderer()
	img := newTestImage(100, 100)

	if err := r.Upload(img); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.img == nil {
		t.Fatal("image should be stored after Upload")
	}
}

func TestSixelRenderer_Display(t *testing.T) {
	r := NewSixelRenderer()
	img := newTestImage(200, 150)
	if err := r.Upload(img); err != nil {
		t.Fatalf("upload: %v", err)
	}

	vp := newTestViewport()
	vp.ImgWidth = 200
	vp.ImgHeight = 150
	vp.TermWidth = 80
	vp.TermHeight = 24

	output, err := r.Display(vp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should start with cursor home
	if !strings.HasPrefix(output, "\x1b[H") {
		t.Error("output should start with cursor home escape")
	}

	// Should contain Sixel introducer (DCS)
	if !strings.Contains(output, "\x1bP") {
		t.Error("output should contain Sixel DCS introducer")
	}

	// Should contain Sixel string terminator (ST)
	if !strings.Contains(output, "\x1b\\") {
		t.Error("output should contain Sixel string terminator")
	}
}

func TestSixelRenderer_Display_ZoomedIn(t *testing.T) {
	r := NewSixelRenderer()
	img := newTestImage(400, 300)
	if err := r.Upload(img); err != nil {
		t.Fatalf("upload: %v", err)
	}

	vp := newTestViewport()
	vp.ImgWidth = 400
	vp.ImgHeight = 300
	vp.TermWidth = 80
	vp.TermHeight = 24
	vp.ZoomLevel = 2.0 // Shows 200x150

	output, err := r.Display(vp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "\x1bP") {
		t.Error("zoomed output should contain Sixel data")
	}
}

func TestSixelRenderer_Display_ZeroSize(t *testing.T) {
	r := NewSixelRenderer()
	img := newTestImage(100, 100)
	if err := r.Upload(img); err != nil {
		t.Fatalf("upload: %v", err)
	}

	tests := []struct {
		name                     string
		termW, termH, imgW, imgH int
	}{
		{"zero terminal width", 0, 24, 100, 100},
		{"zero terminal height", 80, 0, 100, 100},
		{"zero image via viewport", 80, 24, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vp := newTestViewport()
			vp.TermWidth = tt.termW
			vp.TermHeight = tt.termH
			vp.ImgWidth = tt.imgW
			vp.ImgHeight = tt.imgH

			output, err := r.Display(vp)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if output != "" {
				t.Errorf("expected empty output, got %d bytes", len(output))
			}
		})
	}
}

func TestSixelRenderer_Display_NoUpload(t *testing.T) {
	r := NewSixelRenderer()

	vp := newTestViewport()
	vp.ImgWidth = 100
	vp.ImgHeight = 100
	vp.TermWidth = 80
	vp.TermHeight = 24

	_, err := r.Display(vp)
	if err == nil {
		t.Fatal("expected error when no image uploaded")
	}
}

func TestSixelRenderer_Clear(t *testing.T) {
	r := NewSixelRenderer()
	if err := r.Clear(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
