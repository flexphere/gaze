package renderer

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/flexphere/gaze/internal/domain"
)

func newTestSixelRenderer() *SixelRenderer {
	return NewSixelRenderer(8.0, 16.0)
}

func newTestViewport() *domain.Viewport {
	return domain.NewViewport(domain.ViewportConfig{
		ZoomStep: 0.1, PanStep: 0.05, MinZoom: 0.1, MaxZoom: 20.0,
	})
}

func TestSixelRenderer_Display(t *testing.T) {
	r := newTestSixelRenderer()

	img := domain.NewImageEntity(
		image.NewRGBA(image.Rect(0, 0, 800, 600)),
		"test.png", "png",
	)
	if err := r.Upload(img); err != nil {
		t.Fatalf("unexpected upload error: %v", err)
	}

	vp := newTestViewport()
	vp.ImgWidth = 800
	vp.ImgHeight = 600
	vp.TermWidth = 80
	vp.TermHeight = 24
	vp.CellAspectRatio = 2.0

	output, err := r.Display(vp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Display writes Sixel to stdout directly and returns only cursor-home
	// for Bubbletea compatibility
	if output != "\x1b[H" {
		t.Errorf("output should be cursor-home only, got: %q", output)
	}
}

func TestSixelRenderer_Display_ZeroSize(t *testing.T) {
	r := newTestSixelRenderer()

	img := domain.NewImageEntity(
		image.NewRGBA(image.Rect(0, 0, 100, 100)),
		"test.png", "png",
	)
	if err := r.Upload(img); err != nil {
		t.Fatalf("unexpected upload error: %v", err)
	}

	vp := newTestViewport()
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

func TestSixelRenderer_Display_NoUpload(t *testing.T) {
	r := newTestSixelRenderer()

	vp := newTestViewport()
	vp.ImgWidth = 800
	vp.ImgHeight = 600
	vp.TermWidth = 80
	vp.TermHeight = 24

	output, err := r.Display(vp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output != "" {
		t.Errorf("expected empty output without upload, got: %q", output)
	}
}

func TestSixelRenderer_Display_ZoomedIn(t *testing.T) {
	r := newTestSixelRenderer()

	img := domain.NewImageEntity(
		image.NewRGBA(image.Rect(0, 0, 800, 600)),
		"test.png", "png",
	)
	if err := r.Upload(img); err != nil {
		t.Fatalf("unexpected upload error: %v", err)
	}

	vp := newTestViewport()
	vp.ImgWidth = 800
	vp.ImgHeight = 600
	vp.TermWidth = 80
	vp.TermHeight = 24
	vp.CellAspectRatio = 2.0
	vp.ZoomLevel = 2.0

	output, err := r.Display(vp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Sixel data is written to stdout; return value is cursor-home only
	if output != "\x1b[H" {
		t.Errorf("zoomed output should be cursor-home only, got: %q", output)
	}
}

func TestSixelRenderer_Upload(t *testing.T) {
	r := newTestSixelRenderer()

	rgba := image.NewRGBA(image.Rect(0, 0, 100, 50))
	img := domain.NewImageEntity(rgba, "test.png", "png")

	if err := r.Upload(img); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if r.imgW != 100 {
		t.Errorf("imgW = %d, want 100", r.imgW)
	}
	if r.imgH != 50 {
		t.Errorf("imgH = %d, want 50", r.imgH)
	}
	if r.img == nil {
		t.Error("img should be stored after upload")
	}
}

func TestSixelRenderer_Clear(t *testing.T) {
	r := newTestSixelRenderer()
	// Clear is a no-op for Sixel; just ensure it doesn't error
	if err := r.Clear(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func setupSixelMinimapRenderer() *SixelRenderer {
	r := newTestSixelRenderer()
	r.minimapW = 128
	r.minimapH = 96
	r.minimapBase = image.NewRGBA(image.Rect(0, 0, 128, 96))
	r.minimapFrame = image.NewRGBA(image.Rect(0, 0, 128, 96))
	return r
}

func TestSixelRenderer_UploadMinimap(t *testing.T) {
	r := newTestSixelRenderer()

	img := domain.NewImageEntity(
		image.NewRGBA(image.Rect(0, 0, 800, 600)),
		"test.png", "png",
	)

	err := r.UploadMinimap(img, 16, 6, 2.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if r.minimapW <= 0 || r.minimapH <= 0 {
		t.Errorf("minimap dimensions should be positive, got %dx%d", r.minimapW, r.minimapH)
	}
	if r.minimapBase == nil {
		t.Error("minimapBase should be set after upload")
	}
	if r.minimapFrame == nil {
		t.Error("minimapFrame should be set after upload")
	}
}

func TestSixelRenderer_DisplayMinimap(t *testing.T) {
	r := setupSixelMinimapRenderer()

	vp := newTestViewport()
	vp.ImgWidth = 800
	vp.ImgHeight = 600
	vp.TermWidth = 80
	vp.TermHeight = 24
	vp.ZoomLevel = 2.0

	// DisplayMinimap writes Sixel to stdout and returns empty string
	output, err := r.DisplayMinimap(vp, 16, 6, "#FFFFFF")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output != "" {
		t.Errorf("expected empty return (Sixel written to stdout), got length: %d", len(output))
	}

	// Verify cache was populated
	if !r.prevCached {
		t.Error("cache should be populated after first call")
	}
	if r.prevSixel == "" {
		t.Error("cached sixel data should not be empty")
	}
}

func TestSixelRenderer_DisplayMinimap_ZeroSize(t *testing.T) {
	r := newTestSixelRenderer()

	vp := newTestViewport()

	output, err := r.DisplayMinimap(vp, 0, 0, "#FFFFFF")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output != "" {
		t.Errorf("expected empty output for zero size, got length: %d", len(output))
	}
}

func TestSixelRenderer_DisplayMinimap_NoBase(t *testing.T) {
	r := newTestSixelRenderer()

	vp := newTestViewport()
	vp.ImgWidth = 800
	vp.ImgHeight = 600
	vp.TermWidth = 80
	vp.TermHeight = 24

	output, err := r.DisplayMinimap(vp, 16, 6, "#FFFFFF")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output != "" {
		t.Errorf("expected empty output without minimap base, got length: %d", len(output))
	}
}

func TestSixelRenderer_DisplayMinimap_CacheHit(t *testing.T) {
	r := setupSixelMinimapRenderer()

	vp := newTestViewport()
	vp.ImgWidth = 800
	vp.ImgHeight = 600
	vp.TermWidth = 80
	vp.TermHeight = 24
	vp.ZoomLevel = 2.0

	// First call — full encode
	_, err := r.DisplayMinimap(vp, 16, 6, "#FFFFFF")
	if err != nil {
		t.Fatalf("unexpected error on first call: %v", err)
	}

	cachedSixel := r.prevSixel
	if cachedSixel == "" {
		t.Fatal("cached sixel should not be empty after first call")
	}

	// Second call with same viewport — cache hit
	_, err = r.DisplayMinimap(vp, 16, 6, "#FFFFFF")
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}

	// Cache should remain the same
	if r.prevSixel != cachedSixel {
		t.Error("cached sixel should not change on cache hit")
	}
}

func TestSixelRenderer_DisplayMinimap_CacheInvalidatedByClear(t *testing.T) {
	r := setupSixelMinimapRenderer()

	vp := newTestViewport()
	vp.ImgWidth = 800
	vp.ImgHeight = 600
	vp.TermWidth = 80
	vp.TermHeight = 24
	vp.ZoomLevel = 2.0

	// First call
	_, err := r.DisplayMinimap(vp, 16, 6, "#FFFFFF")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !r.prevCached {
		t.Error("cache should be valid after first call")
	}

	// Clear minimap
	if err := r.ClearMinimap(); err != nil {
		t.Fatalf("unexpected error on clear: %v", err)
	}

	if r.prevCached {
		t.Error("cache should be invalidated after ClearMinimap")
	}
}

func TestSixelRenderer_DisplayMinimap_CacheInvalidatedByColorChange(t *testing.T) {
	r := setupSixelMinimapRenderer()

	vp := newTestViewport()
	vp.ImgWidth = 800
	vp.ImgHeight = 600
	vp.TermWidth = 80
	vp.TermHeight = 24
	vp.ZoomLevel = 2.0

	// First call with white border
	_, err := r.DisplayMinimap(vp, 16, 6, "#FFFFFF")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	firstSixel := r.prevSixel

	// Second call with different border color — should re-encode
	_, err = r.DisplayMinimap(vp, 16, 6, "#FF0000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if r.prevSixel == firstSixel {
		t.Error("cached sixel should change when border color changes")
	}
}

func TestEncodeSixel(t *testing.T) {
	// Small test image with known colors
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	// Set all pixels to red
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	output := encodeSixel(img)

	if !strings.HasPrefix(output, "\x1bPq") {
		t.Error("should start with DCS q")
	}
	if !strings.HasSuffix(output, "\x1b\\") {
		t.Error("should end with ST")
	}
	if !strings.Contains(output, "\"1;1;4;4") {
		t.Error("should contain raster attributes with width=4, height=4")
	}
	// Should have at least one palette definition
	if !strings.Contains(output, ";2;") {
		t.Error("should contain palette color definition")
	}
}

func TestEncodeSixel_EmptyImage(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 0, 0))

	output := encodeSixel(img)
	if output != "" {
		t.Errorf("expected empty output for empty image, got length: %d", len(output))
	}
}

func TestBuildSixelPalette(t *testing.T) {
	palette := buildSixelPalette()

	if len(palette) != 216 {
		t.Errorf("palette should have 216 colors (6^3), got %d", len(palette))
	}

	// First color should be black
	first := palette[0].(color.RGBA)
	if first.R != 0 || first.G != 0 || first.B != 0 {
		t.Errorf("first palette color should be black, got %v", first)
	}

	// Last color should be white
	last := palette[len(palette)-1].(color.RGBA)
	if last.R != 255 || last.G != 255 || last.B != 255 {
		t.Errorf("last palette color should be white, got %v", last)
	}
}

func TestScaleRegion(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 100, 100))
	src.SetRGBA(50, 50, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	// Scale a 50x50 region to 10x10
	rect := image.Rect(25, 25, 75, 75)
	dst := scaleRegion(src, rect, 10, 10)

	bounds := dst.Bounds()
	if bounds.Dx() != 10 || bounds.Dy() != 10 {
		t.Errorf("expected 10x10, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestWriteSixelRLE(t *testing.T) {
	tests := []struct {
		name  string
		ch    byte
		count int
		want  string
	}{
		{"single", '?', 1, "?"},
		{"double", '?', 2, "??"},
		{"triple", '?', 3, "???"},
		{"rle 4", '?', 4, "!4?"},
		{"rle 100", 'A', 100, "!100A"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out strings.Builder
			writeSixelRLE(&out, tt.ch, tt.count)
			got := out.String()
			if got != tt.want {
				t.Errorf("writeSixelRLE(%q, %d) = %q, want %q", tt.ch, tt.count, got, tt.want)
			}
		})
	}
}
