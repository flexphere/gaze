package renderer

import (
	"image"
	"image/color"
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

func setupMinimapRenderer() *KittyRenderer {
	r := NewKittyRenderer()
	r.minimapID = 2
	r.minimapW = 128
	r.minimapH = 96
	r.minimapBase = image.NewRGBA(image.Rect(0, 0, 128, 96))
	return r
}

func TestKittyRenderer_DisplayMinimap(t *testing.T) {
	r := setupMinimapRenderer()

	vp := domain.NewViewport(domain.ViewportConfig{
		ZoomStep: 0.1, PanStep: 0.05, MinZoom: 0.1, MaxZoom: 20.0,
	})
	vp.ImgWidth = 800
	vp.ImgHeight = 600
	vp.TermWidth = 80
	vp.TermHeight = 24
	vp.ZoomLevel = 2.0

	output, err := r.DisplayMinimap(vp, 16, 6, "#FFFFFF")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should contain minimap image display command
	if !strings.Contains(output, "i=2") {
		t.Error("output should contain minimap image ID")
	}
	if !strings.Contains(output, "a=p") {
		t.Error("output should contain action=place")
	}
}

func TestKittyRenderer_DisplayMinimap_ZeroSize(t *testing.T) {
	r := NewKittyRenderer()
	r.minimapID = 2

	vp := domain.NewViewport(domain.ViewportConfig{
		ZoomStep: 0.1, PanStep: 0.05, MinZoom: 0.1, MaxZoom: 20.0,
	})

	output, err := r.DisplayMinimap(vp, 0, 0, "#FFFFFF")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output != "" {
		t.Errorf("expected empty output for zero size, got: %q", output)
	}
}

func TestKittyRenderer_DisplayMinimap_NoMinimapID(t *testing.T) {
	r := NewKittyRenderer()

	vp := domain.NewViewport(domain.ViewportConfig{
		ZoomStep: 0.1, PanStep: 0.05, MinZoom: 0.1, MaxZoom: 20.0,
	})
	vp.ImgWidth = 800
	vp.ImgHeight = 600
	vp.TermWidth = 80
	vp.TermHeight = 24

	output, err := r.DisplayMinimap(vp, 16, 6, "#FFFFFF")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output != "" {
		t.Errorf("expected empty output without minimap ID, got: %q", output)
	}
}

func TestKittyRenderer_DisplayMinimap_CursorPosition(t *testing.T) {
	r := setupMinimapRenderer()

	vp := domain.NewViewport(domain.ViewportConfig{
		ZoomStep: 0.1, PanStep: 0.05, MinZoom: 0.1, MaxZoom: 20.0,
	})
	vp.ImgWidth = 800
	vp.ImgHeight = 600
	vp.TermWidth = 80
	vp.TermHeight = 24
	vp.ZoomLevel = 2.0

	cols := 16
	rows := 6

	output, err := r.DisplayMinimap(vp, cols, rows, "#FFFFFF")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Minimap should be positioned at bottom-right
	// startRow = 24 - 6 + 1 = 19, startCol = 80 - 16 + 1 = 65
	if !strings.Contains(output, "\x1b[19;65H") {
		t.Errorf("output should position cursor at row 19, col 65, got: %q", output)
	}
}

func TestKittyRenderer_UploadMinimap(t *testing.T) {
	r := NewKittyRenderer()

	img := domain.NewImageEntity(
		image.NewRGBA(image.Rect(0, 0, 800, 600)),
		"test.png",
		"png",
	)

	err := r.UploadMinimap(img, 16, 6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if r.minimapID == 0 {
		t.Error("minimapID should be set after upload")
	}
	if r.minimapW <= 0 || r.minimapH <= 0 {
		t.Errorf("minimap dimensions should be positive, got %dx%d", r.minimapW, r.minimapH)
	}
	if r.minimapBase == nil {
		t.Error("minimapBase should be set after upload")
	}
}

func TestKittyRenderer_DisplayMinimap_NoBase(t *testing.T) {
	r := NewKittyRenderer()
	r.minimapID = 2
	// minimapBase is nil

	vp := domain.NewViewport(domain.ViewportConfig{
		ZoomStep: 0.1, PanStep: 0.05, MinZoom: 0.1, MaxZoom: 20.0,
	})
	vp.ImgWidth = 800
	vp.ImgHeight = 600
	vp.TermWidth = 80
	vp.TermHeight = 24

	output, err := r.DisplayMinimap(vp, 16, 6, "#FFFFFF")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output != "" {
		t.Errorf("expected empty output without minimap base, got: %q", output)
	}
}

func TestClampInt(t *testing.T) {
	tests := []struct {
		name          string
		val, min, max int
		want          int
	}{
		{"within range", 5, 0, 10, 5},
		{"below min", -1, 0, 10, 0},
		{"above max", 15, 0, 10, 10},
		{"at min", 0, 0, 10, 0},
		{"at max", 10, 0, 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clampInt(tt.val, tt.min, tt.max)
			if got != tt.want {
				t.Errorf("clampInt(%d, %d, %d) = %d, want %d", tt.val, tt.min, tt.max, got, tt.want)
			}
		})
	}
}

func TestParseHexColor(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  color.RGBA
	}{
		{"white", "#FFFFFF", color.RGBA{R: 255, G: 255, B: 255, A: 230}},
		{"red", "#FF0000", color.RGBA{R: 255, G: 0, B: 0, A: 230}},
		{"green", "#00FF00", color.RGBA{R: 0, G: 255, B: 0, A: 230}},
		{"blue", "#0000FF", color.RGBA{R: 0, G: 0, B: 255, A: 230}},
		{"lowercase", "#ff8800", color.RGBA{R: 255, G: 136, B: 0, A: 230}},
		{"no hash", "FF0000", color.RGBA{R: 255, G: 0, B: 0, A: 230}},
		{"empty fallback", "", color.RGBA{R: 255, G: 255, B: 255, A: 230}},
		{"invalid fallback", "xyz", color.RGBA{R: 255, G: 255, B: 255, A: 230}},
		{"short fallback", "#FFF", color.RGBA{R: 255, G: 255, B: 255, A: 230}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseHexColor(tt.input)
			if got != tt.want {
				t.Errorf("parseHexColor(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
