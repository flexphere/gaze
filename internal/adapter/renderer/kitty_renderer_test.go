package renderer

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"io"
	"strings"
	"testing"
	"time"

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
	if !strings.Contains(output, "p=1") {
		t.Error("output should contain placement ID for consistent cross-terminal behavior")
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
	// Verify p=1 within the placement command only (not in base64 payload)
	aPos := strings.Index(output, "a=p")
	if aPos == -1 {
		t.Fatal("output should contain action=place (a=p) command")
	}
	termPosRel := strings.Index(output[aPos:], "\x1b\\")
	if termPosRel == -1 {
		t.Fatalf("output should contain kitty graphics terminator after placement command: %q", output)
	}
	placementCmd := output[aPos : aPos+termPosRel]
	if !strings.Contains(placementCmd, "p=1") {
		t.Errorf("placement command should contain placement ID p=1 for consistent cross-terminal behavior, got: %q", placementCmd)
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
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	// Set a pixel to verify data is included
	img.SetRGBA(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	output := buildRGBAUploadSequence(42, img)

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
	// Create a larger image and take a sub-image (non-zero origin, stride != 4*w)
	full := image.NewRGBA(image.Rect(0, 0, 8, 8))
	full.SetRGBA(2, 2, color.RGBA{R: 42, G: 0, B: 0, A: 255})
	sub := full.SubImage(image.Rect(2, 2, 6, 6)).(*image.RGBA)

	output := buildRGBAUploadSequence(10, sub)

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

func TestEncodeFramePayload(t *testing.T) {
	flat := image.NewRGBA(image.Rect(0, 0, 40, 30))
	for i := range flat.Pix {
		flat.Pix[i] = 0x80
	}

	tests := []struct {
		name       string
		src        image.Image
		wantFormat string
		wantPix    []byte // expected decompressed pixels, nil to skip
	}{
		{
			name:       "RGBA frame is sent as zlib-compressed raw pixels",
			src:        flat,
			wantFormat: "f=32,s=40,v=30,o=z",
			wantPix:    flat.Pix,
		},
		{
			name:       "non-RGBA frame falls back to PNG",
			src:        image.NewGray(image.Rect(0, 0, 4, 3)),
			wantFormat: "f=100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := encodeFramePayload(tt.src)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p.format != tt.wantFormat {
				t.Errorf("format = %q, want %q", p.format, tt.wantFormat)
			}
			if tt.wantPix == nil {
				return
			}
			raw, err := base64.StdEncoding.DecodeString(p.encoded)
			if err != nil {
				t.Fatalf("decoding base64: %v", err)
			}
			if len(raw) >= len(tt.wantPix) {
				t.Errorf("compressed size %d should be smaller than raw size %d", len(raw), len(tt.wantPix))
			}
			zr, err := zlib.NewReader(bytes.NewReader(raw))
			if err != nil {
				t.Fatalf("opening zlib stream: %v", err)
			}
			got, err := io.ReadAll(zr)
			if err != nil {
				t.Fatalf("decompressing: %v", err)
			}
			if !bytes.Equal(got, tt.wantPix) {
				t.Error("decompressed pixels do not match the frame")
			}
		})
	}
}

func TestBuildChunkedUploadSequence(t *testing.T) {
	tests := []struct {
		name       string
		encodedLen int
		wantChunks int
	}{
		{name: "single chunk", encodedLen: 100, wantChunks: 1},
		{name: "exactly one chunk", encodedLen: 4096, wantChunks: 1},
		{name: "multiple chunks", encodedLen: 4096*2 + 1, wantChunks: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := framePayload{format: "f=32,s=1,v=1,o=z", encoded: strings.Repeat("A", tt.encodedLen)}
			got := buildChunkedUploadSequence(9, p)

			if !strings.HasPrefix(got, "\x1b_Gi=9,f=32,s=1,v=1,o=z,a=t,t=d,q=2,") {
				t.Errorf("first chunk header is wrong: %q", got[:min(len(got), 60)])
			}
			if n := strings.Count(got, "\x1b_G"); n != tt.wantChunks {
				t.Errorf("chunk count = %d, want %d", n, tt.wantChunks)
			}
			if n := strings.Count(got, "m=0;"); n != 1 {
				t.Errorf("exactly one final chunk (m=0) expected, got %d", n)
			}
		})
	}
}

func TestKittyRenderer_FrameDoubleBuffer(t *testing.T) {
	vp := domain.NewViewport(domain.ViewportConfig{
		ZoomStep: 0.1, PanStep: 0.05, MinZoom: 0.1, MaxZoom: 20.0,
	})
	vp.SetImageSize(4, 3)
	vp.SetTerminalSize(80, 24)
	frame := image.NewRGBA(image.Rect(0, 0, 4, 3))

	clock := time.Unix(0, 0)
	r := NewKittyRenderer()
	r.now = func() time.Time { return clock }
	r.imageID = 100 // as if Upload had been called

	// Static image: no back buffer, so Display must not delete anything.
	out, err := r.Display(vp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out, "a=d") {
		t.Errorf("static display should not delete placements, got %q", out)
	}

	front := r.imageID
	for i := 0; i < 3; i++ {
		seq, err := r.buildFrameUpload(frame)
		if err != nil {
			t.Fatalf("frame %d: unexpected error: %v", i, err)
		}
		back := r.spareID
		if back != front {
			t.Fatalf("frame %d: spareID = %d, want previous front %d", i, back, front)
		}
		if !strings.HasPrefix(seq, fmt.Sprintf("\x1b_Gi=%d,", r.imageID)) {
			t.Errorf("frame %d: upload should target new front %d, got prefix %q", i, r.imageID, seq[:20])
		}

		out, err := r.Display(vp)
		if err != nil {
			t.Fatalf("frame %d: unexpected error: %v", i, err)
		}
		place := strings.Index(out, fmt.Sprintf("a=p,i=%d,", r.imageID))
		del := strings.Index(out, fmt.Sprintf("a=d,d=i,i=%d,", back))
		if place < 0 || del < 0 || del < place {
			t.Errorf("frame %d: want placement of %d followed by deletion of %d, got %q", i, r.imageID, back, out)
		}
		front = r.imageID
		clock = clock.Add(time.Hour) // stay within the transfer budget
	}
}

func TestKittyRenderer_FrameTransferBudget(t *testing.T) {
	// Noise does not compress, so each frame costs roughly its raw size.
	noise := image.NewRGBA(image.Rect(0, 0, 256, 256))
	for i := range noise.Pix {
		noise.Pix[i] = byte(i*131 + i/7)
	}
	frames := []image.Image{noise, image.NewRGBA(image.Rect(0, 0, 256, 256))}

	clock := time.Unix(0, 0)
	r := NewKittyRenderer()
	r.now = func() time.Time { return clock }
	r.imageID = 100

	seq, err := r.buildFrameUpload(frames[0])
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if seq == "" {
		t.Fatal("first frame should be sent")
	}
	wait := time.Duration(len(seq)) * time.Second / maxFrameBytesPerSecond
	front := r.imageID

	tests := []struct {
		name     string
		elapsed  time.Duration
		wantSent bool
	}{
		{name: "skipped right after a large frame", elapsed: 0, wantSent: false},
		{name: "skipped just before the budget recovers", elapsed: wait - time.Millisecond, wantSent: false},
		{name: "sent once the budget recovers", elapsed: wait, wantSent: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clock = time.Unix(0, 0).Add(tt.elapsed)
			got, err := r.buildFrameUpload(frames[1])
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if sent := got != ""; sent != tt.wantSent {
				t.Errorf("sent = %v, want %v", sent, tt.wantSent)
			}
			if !tt.wantSent && r.imageID != front {
				t.Errorf("skipped frame must not swap image IDs: imageID = %d, want %d", r.imageID, front)
			}
		})
	}
}

func TestKittyRenderer_FramePayloadCache(t *testing.T) {
	frame := image.NewRGBA(image.Rect(0, 0, 8, 8))
	r := NewKittyRenderer()

	first, err := r.framePayload(frame)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Mutating the pixels must not change the payload: the cached encoding
	// is reused instead of compressing the frame again.
	frame.Pix[0] = 0xFF
	second, err := r.framePayload(frame)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if first != second {
		t.Error("payload should be served from the cache on the second call")
	}
}
