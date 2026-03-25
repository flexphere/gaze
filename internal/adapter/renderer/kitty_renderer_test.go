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
	vp.CellAspectRatio = 2.0

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
	// VH = 800*24*2/80 = 480
	if !strings.Contains(output, "h=480") {
		t.Errorf("output should contain h=480 (visible height with cell aspect), got: %s", output)
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
	vp.CellAspectRatio = 2.0
	vp.ZoomLevel = 2.0 // VW=400, VH=400*24*2/80=240

	output, err := r.Display(vp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "w=400") {
		t.Errorf("output should contain w=400 for 2x zoom, got: %s", output)
	}
	if !strings.Contains(output, "h=240") {
		t.Errorf("output should contain h=240 for 2x zoom with cell aspect, got: %s", output)
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
	vp.CellAspectRatio = 2.0

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
	r.minimapFrame = image.NewRGBA(image.Rect(0, 0, 128, 96))
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

	// Should contain placement and upload commands
	if !strings.Contains(output, "a=p") {
		t.Error("output should contain action=place")
	}
	// Uses raw RGBA format
	if !strings.Contains(output, "f=32") {
		t.Error("output should use raw RGBA format (f=32)")
	}
	// Reuses same minimap ID
	if !strings.Contains(output, "i=2") {
		t.Error("output should contain minimap image ID")
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

func TestKittyRenderer_DisplayMinimap_CacheHit(t *testing.T) {
	r := setupMinimapRenderer()

	vp := domain.NewViewport(domain.ViewportConfig{
		ZoomStep: 0.1, PanStep: 0.05, MinZoom: 0.1, MaxZoom: 20.0,
	})
	vp.ImgWidth = 800
	vp.ImgHeight = 600
	vp.TermWidth = 80
	vp.TermHeight = 24
	vp.ZoomLevel = 2.0

	// First call — full upload
	output1, err := r.DisplayMinimap(vp, 16, 6, "#FFFFFF")
	if err != nil {
		t.Fatalf("unexpected error on first call: %v", err)
	}
	if !strings.Contains(output1, "f=32") {
		t.Error("first call should upload with f=32")
	}

	// Second call with same viewport — cache hit, only placement
	output2, err := r.DisplayMinimap(vp, 16, 6, "#FFFFFF")
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}
	if strings.Contains(output2, "f=32") {
		t.Error("second call should not re-upload (cache hit)")
	}
	if !strings.Contains(output2, "a=p") {
		t.Error("second call should still contain placement command")
	}
}

func TestBuildRGBAUploadSequence(t *testing.T) {
	r := NewKittyRenderer()
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	// Set a pixel to verify data is included
	img.SetRGBA(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	output := r.buildRGBAUploadSequence(42, img)

	if !strings.Contains(output, "i=42") {
		t.Error("output should contain image ID")
	}
	if !strings.Contains(output, "f=32") {
		t.Error("output should specify raw RGBA format")
	}
	if !strings.Contains(output, "s=4") {
		t.Error("output should contain image width")
	}
	if !strings.Contains(output, "v=4") {
		t.Error("output should contain image height")
	}
	if !strings.Contains(output, "a=t") {
		t.Error("output should contain transmit action")
	}
	if !strings.Contains(output, "q=2") {
		t.Error("output should contain quiet mode flag")
	}
}

func TestBuildRGBAUploadSequence_SubImage(t *testing.T) {
	r := NewKittyRenderer()
	// Create a larger image and take a sub-image (non-zero origin, stride != 4*w)
	full := image.NewRGBA(image.Rect(0, 0, 8, 8))
	full.SetRGBA(2, 2, color.RGBA{R: 42, G: 0, B: 0, A: 255})
	sub := full.SubImage(image.Rect(2, 2, 6, 6)).(*image.RGBA)

	output := r.buildRGBAUploadSequence(10, sub)

	if !strings.Contains(output, "s=4") {
		t.Error("sub-image should report width=4")
	}
	if !strings.Contains(output, "v=4") {
		t.Error("sub-image should report height=4")
	}
}

func TestKittyRenderer_DisplayMinimap_CacheInvalidatedByClear(t *testing.T) {
	r := setupMinimapRenderer()

	vp := domain.NewViewport(domain.ViewportConfig{
		ZoomStep: 0.1, PanStep: 0.05, MinZoom: 0.1, MaxZoom: 20.0,
	})
	vp.ImgWidth = 800
	vp.ImgHeight = 600
	vp.TermWidth = 80
	vp.TermHeight = 24
	vp.ZoomLevel = 2.0

	// First call — full upload
	_, err := r.DisplayMinimap(vp, 16, 6, "#FFFFFF")
	if err != nil {
		t.Fatalf("unexpected error on first call: %v", err)
	}

	// Clear minimap (removes from terminal)
	_ = r.ClearMinimap()

	// Third call — should re-upload since clear invalidated cache
	output, err := r.DisplayMinimap(vp, 16, 6, "#FFFFFF")
	if err != nil {
		t.Fatalf("unexpected error after clear: %v", err)
	}
	if !strings.Contains(output, "f=32") {
		t.Error("should re-upload after ClearMinimap (cache should be invalidated)")
	}
}

func TestKittyRenderer_DisplayMinimap_CacheInvalidatedByColorChange(t *testing.T) {
	r := setupMinimapRenderer()

	vp := domain.NewViewport(domain.ViewportConfig{
		ZoomStep: 0.1, PanStep: 0.05, MinZoom: 0.1, MaxZoom: 20.0,
	})
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

	// Second call with different border color — should re-upload
	output, err := r.DisplayMinimap(vp, 16, 6, "#FF0000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "f=32") {
		t.Error("should re-upload when border color changes")
	}
}

func TestKittyRenderer_UploadMinimap(t *testing.T) {
	r := NewKittyRenderer()

	img := domain.NewImageEntity(
		image.NewRGBA(image.Rect(0, 0, 800, 600)),
		"test.png",
		"png",
	)

	err := r.UploadMinimap(img, 16, 6, 2.0)
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

func TestWrapDCSPassthrough(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			"simple APC",
			"\x1b_Ga=d,d=i,i=1\x1b\\",
			"\x1bPtmux;\x1b\x1b_Ga=d,d=i,i=1\x1b\x1b\\\x1b\\",
		},
		{
			"no ESC chars",
			"hello",
			"\x1bPtmux;hello\x1b\\",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapDCSPassthrough(tt.input)
			if got != tt.want {
				t.Errorf("wrapDCSPassthrough(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestKittyRenderer_TmuxMode_Display(t *testing.T) {
	r := NewKittyRenderer(WithTmuxMode(true))
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
	vp.CellAspectRatio = 2.0

	output, err := r.Display(vp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should start with cursor move (NOT wrapped)
	if !strings.HasPrefix(output, "\x1b[H") {
		t.Error("output should start with cursor move \\x1b[H")
	}

	// APC should be wrapped in DCS passthrough
	if !strings.Contains(output, "\x1bPtmux;") {
		t.Error("tmux mode output should contain DCS passthrough prefix")
	}
	// The wrapped content should still contain the placement params
	if !strings.Contains(output, "a=p") {
		t.Error("tmux mode output should contain action=place")
	}
	if !strings.Contains(output, "i=1") {
		t.Error("tmux mode output should contain image ID")
	}
}

func TestKittyRenderer_TmuxMode_DisplayMinimap(t *testing.T) {
	r := NewKittyRenderer(WithTmuxMode(true))
	r.minimapID = 2
	r.minimapW = 128
	r.minimapH = 96
	r.minimapBase = image.NewRGBA(image.Rect(0, 0, 128, 96))
	r.minimapFrame = image.NewRGBA(image.Rect(0, 0, 128, 96))

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

	// Upload chunks should be wrapped
	if !strings.Contains(output, "\x1bPtmux;") {
		t.Error("tmux mode minimap should contain DCS passthrough")
	}
	// Placement APC should also be wrapped
	if count := strings.Count(output, "\x1bPtmux;"); count < 2 {
		t.Errorf("expected at least 2 DCS wrappers (upload + placement), got %d", count)
	}
}

func TestKittyRenderer_TmuxMode_DisplayMinimap_CacheHit(t *testing.T) {
	r := NewKittyRenderer(WithTmuxMode(true))
	r.minimapID = 2
	r.minimapW = 128
	r.minimapH = 96
	r.minimapBase = image.NewRGBA(image.Rect(0, 0, 128, 96))
	r.minimapFrame = image.NewRGBA(image.Rect(0, 0, 128, 96))

	vp := domain.NewViewport(domain.ViewportConfig{
		ZoomStep: 0.1, PanStep: 0.05, MinZoom: 0.1, MaxZoom: 20.0,
	})
	vp.ImgWidth = 800
	vp.ImgHeight = 600
	vp.TermWidth = 80
	vp.TermHeight = 24
	vp.ZoomLevel = 2.0

	// First call — full upload
	_, err := r.DisplayMinimap(vp, 16, 6, "#FFFFFF")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Second call — cache hit, placement only
	output, err := r.DisplayMinimap(vp, 16, 6, "#FFFFFF")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Placement should still be wrapped in tmux mode
	if !strings.Contains(output, "\x1bPtmux;") {
		t.Error("cached placement should still be wrapped in tmux mode")
	}
}

func TestKittyRenderer_TmuxMode_BuildRGBAUploadSequence(t *testing.T) {
	r := NewKittyRenderer(WithTmuxMode(true))
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))

	output := r.buildRGBAUploadSequence(42, img)

	// Each chunk should be individually wrapped
	if !strings.Contains(output, "\x1bPtmux;") {
		t.Error("tmux mode upload should wrap chunks in DCS passthrough")
	}
	// Should NOT contain raw (unwrapped) APC start
	// In wrapped form, \x1b_ becomes \x1b\x1b_ inside the DCS payload
	if strings.Contains(output, "\x1b_G") && !strings.Contains(output, "\x1b\x1b_G") {
		t.Error("tmux mode should not contain unwrapped APC sequences")
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
