package filesystem

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func createTestImage(t *testing.T, dir, name, format string, w, h int) string {
	t.Helper()
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("creating test image: %v", err)
	}
	defer f.Close()

	img := image.NewRGBA(image.Rect(0, 0, w, h))

	switch format {
	case "png":
		if err := png.Encode(f, img); err != nil {
			t.Fatalf("encoding PNG: %v", err)
		}
	case "jpeg":
		if err := jpeg.Encode(f, img, nil); err != nil {
			t.Fatalf("encoding JPEG: %v", err)
		}
	case "gif":
		if err := gif.Encode(f, img, nil); err != nil {
			t.Fatalf("encoding GIF: %v", err)
		}
	default:
		t.Fatalf("unsupported image format: %q", format)
	}

	return path
}

func TestImageLoader_Load_PNG(t *testing.T) {
	dir := t.TempDir()
	path := createTestImage(t, dir, "test.png", "png", 100, 80)

	loader := NewImageLoader()
	entity, err := loader.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entity.Width != 100 {
		t.Errorf("Width = %d, want 100", entity.Width)
	}
	if entity.Height != 80 {
		t.Errorf("Height = %d, want 80", entity.Height)
	}
	if entity.Format != "png" {
		t.Errorf("Format = %q, want %q", entity.Format, "png")
	}
	if entity.Path != path {
		t.Errorf("Path = %q, want %q", entity.Path, path)
	}
}

func TestImageLoader_Load_JPEG(t *testing.T) {
	dir := t.TempDir()
	path := createTestImage(t, dir, "test.jpg", "jpeg", 200, 150)

	loader := NewImageLoader()
	entity, err := loader.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entity.Width != 200 {
		t.Errorf("Width = %d, want 200", entity.Width)
	}
	if entity.Format != "jpeg" {
		t.Errorf("Format = %q, want %q", entity.Format, "jpeg")
	}
}

func TestImageLoader_Load_GIF(t *testing.T) {
	dir := t.TempDir()
	path := createTestImage(t, dir, "test.gif", "gif", 50, 50)

	loader := NewImageLoader()
	entity, err := loader.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entity.Width != 50 {
		t.Errorf("Width = %d, want 50", entity.Width)
	}
	if entity.Format != "gif" {
		t.Errorf("Format = %q, want %q", entity.Format, "gif")
	}
}

var (
	testRed   = color.RGBA{R: 255, A: 255}
	testGreen = color.RGBA{G: 255, A: 255}
	testBlue  = color.RGBA{B: 255, A: 255}
)

// testGIFFrame describes one frame of a generated animated GIF.
type testGIFFrame struct {
	rect     image.Rectangle
	c        color.RGBA
	delay    int
	disposal byte
}

func encodeTestGIF(t *testing.T, w, h int, frames []testGIFFrame) []byte {
	t.Helper()
	palette := color.Palette{color.Transparent, testRed, testGreen, testBlue}
	g := &gif.GIF{Config: image.Config{Width: w, Height: h, ColorModel: palette}}
	for _, f := range frames {
		img := image.NewPaletted(f.rect, palette)
		draw.Draw(img, f.rect, image.NewUniform(f.c), image.Point{}, draw.Src)
		g.Image = append(g.Image, img)
		g.Delay = append(g.Delay, f.delay)
		g.Disposal = append(g.Disposal, f.disposal)
	}
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, g); err != nil {
		t.Fatalf("encoding animated GIF: %v", err)
	}
	return buf.Bytes()
}

func writeTestFile(t *testing.T, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "anim.gif")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}
	return path
}

func TestImageLoader_Load_AnimatedGIF(t *testing.T) {
	full := image.Rect(0, 0, 4, 4)
	topLeft := image.Rect(0, 0, 2, 2)
	bottomRight := image.Rect(2, 2, 4, 4)

	// pixel is a sample point checked on a specific composited frame.
	type pixel struct {
		frame int
		x, y  int
		want  color.RGBA
	}

	tests := []struct {
		name       string
		frames     []testGIFFrame
		wantDelays []time.Duration
		wantPixels []pixel
	}{
		{
			name: "delays are converted and tiny delays fall back to 100ms",
			frames: []testGIFFrame{
				{rect: full, c: testRed, delay: 0},
				{rect: full, c: testGreen, delay: 1},
				{rect: full, c: testBlue, delay: 5},
			},
			wantDelays: []time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 50 * time.Millisecond},
			wantPixels: []pixel{
				{frame: 0, x: 0, y: 0, want: testRed},
				{frame: 1, x: 0, y: 0, want: testGreen},
				{frame: 2, x: 0, y: 0, want: testBlue},
			},
		},
		{
			name: "disposal none keeps previous pixels outside the next frame",
			frames: []testGIFFrame{
				{rect: full, c: testRed, delay: 10, disposal: gif.DisposalNone},
				{rect: topLeft, c: testGreen, delay: 10, disposal: gif.DisposalNone},
			},
			wantDelays: []time.Duration{100 * time.Millisecond, 100 * time.Millisecond},
			wantPixels: []pixel{
				{frame: 1, x: 0, y: 0, want: testGreen},
				{frame: 1, x: 3, y: 3, want: testRed},
			},
		},
		{
			name: "disposal background clears the frame area to transparent",
			frames: []testGIFFrame{
				{rect: full, c: testRed, delay: 10, disposal: gif.DisposalBackground},
				{rect: topLeft, c: testGreen, delay: 10, disposal: gif.DisposalNone},
			},
			wantDelays: []time.Duration{100 * time.Millisecond, 100 * time.Millisecond},
			wantPixels: []pixel{
				{frame: 1, x: 0, y: 0, want: testGreen},
				{frame: 1, x: 3, y: 3, want: color.RGBA{}},
			},
		},
		{
			name: "disposal previous restores the canvas before the frame",
			frames: []testGIFFrame{
				{rect: full, c: testRed, delay: 10, disposal: gif.DisposalNone},
				{rect: topLeft, c: testGreen, delay: 10, disposal: gif.DisposalPrevious},
				{rect: bottomRight, c: testBlue, delay: 10, disposal: gif.DisposalNone},
			},
			wantDelays: []time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
			wantPixels: []pixel{
				{frame: 1, x: 0, y: 0, want: testGreen},
				{frame: 2, x: 0, y: 0, want: testRed},
				{frame: 2, x: 3, y: 3, want: testBlue},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeTestFile(t, encodeTestGIF(t, 4, 4, tt.frames))

			entity, err := NewImageLoader().Load(path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !entity.IsAnimated() {
				t.Fatal("IsAnimated() = false, want true")
			}
			if entity.Width != 4 || entity.Height != 4 {
				t.Errorf("size = %dx%d, want 4x4", entity.Width, entity.Height)
			}
			if len(entity.Frames) != len(tt.wantDelays) {
				t.Fatalf("len(Frames) = %d, want %d", len(entity.Frames), len(tt.wantDelays))
			}
			for i, want := range tt.wantDelays {
				if got := entity.Frames[i].Delay; got != want {
					t.Errorf("Frames[%d].Delay = %v, want %v", i, got, want)
				}
			}
			for _, p := range tt.wantPixels {
				got := color.RGBAModel.Convert(entity.Frames[p.frame].Image.At(p.x, p.y)).(color.RGBA)
				if got != p.want {
					t.Errorf("Frames[%d] at (%d,%d) = %v, want %v", p.frame, p.x, p.y, got, p.want)
				}
			}
		})
	}
}

func TestImageLoader_Load_TruncatedAnimatedGIF(t *testing.T) {
	data := encodeTestGIF(t, 4, 4, []testGIFFrame{
		{rect: image.Rect(0, 0, 4, 4), c: testRed, delay: 10},
		{rect: image.Rect(0, 0, 4, 4), c: testGreen, delay: 10},
	})
	// Cut off the trailer and the end of the second frame so that only the
	// first frame is decodable.
	path := writeTestFile(t, data[:len(data)-4])

	entity, err := NewImageLoader().Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entity.IsAnimated() {
		t.Error("IsAnimated() = true, want false for a GIF whose later frames are broken")
	}
	if entity.Width != 4 || entity.Height != 4 {
		t.Errorf("size = %dx%d, want 4x4", entity.Width, entity.Height)
	}
}

func TestImageLoader_Load_BMP(t *testing.T) {
	path := filepath.Join("..", "..", "..", "testdata", "test.bmp")
	loader := NewImageLoader()
	entity, err := loader.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entity.Width != 100 {
		t.Errorf("Width = %d, want 100", entity.Width)
	}
	if entity.Height != 100 {
		t.Errorf("Height = %d, want 100", entity.Height)
	}
	if entity.Format != "bmp" {
		t.Errorf("Format = %q, want %q", entity.Format, "bmp")
	}
}

func TestImageLoader_Load_TIFF(t *testing.T) {
	path := filepath.Join("..", "..", "..", "testdata", "test.tiff")
	loader := NewImageLoader()
	entity, err := loader.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entity.Width != 100 {
		t.Errorf("Width = %d, want 100", entity.Width)
	}
	if entity.Height != 100 {
		t.Errorf("Height = %d, want 100", entity.Height)
	}
	if entity.Format != "tiff" {
		t.Errorf("Format = %q, want %q", entity.Format, "tiff")
	}
}

func TestImageLoader_Load_WebP(t *testing.T) {
	path := filepath.Join("..", "..", "..", "testdata", "test.webp")
	loader := NewImageLoader()
	entity, err := loader.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entity.Width != 100 {
		t.Errorf("Width = %d, want 100", entity.Width)
	}
	if entity.Height != 100 {
		t.Errorf("Height = %d, want 100", entity.Height)
	}
	if entity.Format != "webp" {
		t.Errorf("Format = %q, want %q", entity.Format, "webp")
	}
}

func TestImageLoader_Load_FileNotFound(t *testing.T) {
	loader := NewImageLoader()
	_, err := loader.Load("/nonexistent/path/image.png")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestImageLoader_Load_NotAnImage(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "notimage.txt")
	if err := os.WriteFile(path, []byte("hello world"), 0o644); err != nil {
		t.Fatalf("creating test file: %v", err)
	}

	loader := NewImageLoader()
	_, err := loader.Load(path)
	if err == nil {
		t.Fatal("expected error for non-image file, got nil")
	}
}
