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

	// Should start with cursor-to-home + DCS Sixel start
	if !strings.HasPrefix(output, "\x1b[H\x1bPq") {
		t.Errorf("output should start with cursor-home + DCS q, got prefix: %q", output[:min(len(output), 20)])
	}
	// Should end with string terminator
	if !strings.HasSuffix(output, "\x1b\\") {
		t.Error("output should end with string terminator (ST)")
	}
	// Should contain raster attributes
	if !strings.Contains(output, "\"1;1;") {
		t.Error("output should contain raster attributes")
	}
	// Should contain palette definition
	if !strings.Contains(output, ";2;") {
		t.Error("output should contain palette color definitions")
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
		t.Errorf("expected empty output for zero terminal size, got length: %d", len(output))
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
		t.Errorf("expected empty output without upload, got length: %d", len(output))
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

	if !strings.HasPrefix(output, "\x1b[H\x1bPq") {
		t.Error("zoomed output should start with cursor-home + DCS q")
	}
	if !strings.HasSuffix(output, "\x1b\\") {
		t.Error("zoomed output should end with ST")
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

	output, err := r.DisplayMinimap(vp, 16, 6, "#FFFFFF")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should contain cursor position and Sixel data
	if !strings.Contains(output, "\x1b[") {
		t.Error("output should contain cursor positioning escape")
	}
	if !strings.Contains(output, "\x1bPq") {
		t.Error("output should contain Sixel DCS start")
	}
	if !strings.HasSuffix(output, "\x1b\\") {
		t.Error("output should end with ST")
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

func TestSixelRenderer_DisplayMinimap_CursorPosition(t *testing.T) {
	r := setupSixelMinimapRenderer()

	vp := newTestViewport()
	vp.ImgWidth = 800
	vp.ImgHeight = 600
	vp.TermWidth = 80
	vp.TermHeight = 24
	vp.ZoomLevel = 2.0

	output, err := r.DisplayMinimap(vp, 16, 6, "#FFFFFF")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Minimap at bottom-right: startRow=24-6+1=19, startCol=80-16+1=65
	if !strings.Contains(output, "\x1b[19;65H") {
		t.Errorf("output should position cursor at row 19, col 65, got: %q", output[:min(len(output), 40)])
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
	output1, err := r.DisplayMinimap(vp, 16, 6, "#FFFFFF")
	if err != nil {
		t.Fatalf("unexpected error on first call: %v", err)
	}
	if !strings.Contains(output1, "\x1bPq") {
		t.Error("first call should contain Sixel data")
	}

	// Second call with same viewport — cache hit
	output2, err := r.DisplayMinimap(vp, 16, 6, "#FFFFFF")
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}

	// Cache hit should be shorter (no sixel re-encode, just cursor + cached sixel)
	// Both should contain Sixel data (cached includes it), but output should match
	if output1 != output2 {
		t.Error("second call should return same output from cache")
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

	// Second call with different border color — should re-encode
	output, err := r.DisplayMinimap(vp, 16, 6, "#FF0000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "\x1bPq") {
		t.Error("should re-encode Sixel when border color changes")
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
