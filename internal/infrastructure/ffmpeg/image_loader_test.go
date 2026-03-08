package ffmpeg

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/flexphere/gaze/internal/domain"
	"github.com/flexphere/gaze/internal/infrastructure/filesystem"
)

func skipIfFFmpegUnavailable(t *testing.T) {
	t.Helper()
	if !Available() {
		t.Skip("ffmpeg/ffprobe not available")
	}
}

func createTestPNG(t *testing.T, dir, name string, w, h int) string {
	t.Helper()
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("creating test image: %v", err)
	}
	defer f.Close()

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("encoding PNG: %v", err)
	}
	return path
}

func TestImageLoader_Load_FallbackToGoDecoder(t *testing.T) {
	// PNG is supported by Go standard library, so ffmpeg should not be needed.
	dir := t.TempDir()
	path := createTestPNG(t, dir, "test.png", 120, 90)

	fallback := filesystem.NewImageLoader()
	loader := NewImageLoader(fallback)

	entity, err := loader.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entity.Width != 120 {
		t.Errorf("Width = %d, want 120", entity.Width)
	}
	if entity.Height != 90 {
		t.Errorf("Height = %d, want 90", entity.Height)
	}
	if entity.Format != "png" {
		t.Errorf("Format = %q, want %q", entity.Format, "png")
	}
}

func TestImageLoader_Load_FFmpegFallback(t *testing.T) {
	skipIfFFmpegUnavailable(t)

	// Use a format that Go standard library cannot decode.
	// Create a TGA image via ffmpeg from a PNG source.
	dir := t.TempDir()
	srcPath := createTestPNG(t, dir, "src.png", 80, 60)
	tgaPath := filepath.Join(dir, "test.tga")

	convertCmd := exec.Command("ffmpeg",
		"-i", srcPath,
		"-y",
		tgaPath,
	)
	if out, err := convertCmd.CombinedOutput(); err != nil {
		t.Fatalf("converting to TGA: %v: %s", err, out)
	}

	// Use a fallback that always fails, to prove ffmpeg path is taken.
	failLoader := &failingLoader{}
	loader := NewImageLoader(failLoader)

	entity, err := loader.Load(tgaPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entity.Width != 80 {
		t.Errorf("Width = %d, want 80", entity.Width)
	}
	if entity.Height != 60 {
		t.Errorf("Height = %d, want 60", entity.Height)
	}
}

func TestImageLoader_Load_FileNotFound(t *testing.T) {
	fallback := filesystem.NewImageLoader()
	loader := NewImageLoader(fallback)

	_, err := loader.Load("/nonexistent/path/image.avif")
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

	fallback := filesystem.NewImageLoader()
	loader := NewImageLoader(fallback)

	_, err := loader.Load(path)
	if err == nil {
		t.Fatal("expected error for non-image file, got nil")
	}
}

func TestProbe(t *testing.T) {
	skipIfFFmpegUnavailable(t)

	dir := t.TempDir()
	path := createTestPNG(t, dir, "probe.png", 200, 150)

	w, h, codec, err := probe(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w != 200 {
		t.Errorf("width = %d, want 200", w)
	}
	if h != 150 {
		t.Errorf("height = %d, want 150", h)
	}
	if codec != "png" {
		t.Errorf("codec = %q, want %q", codec, "png")
	}
}

func TestProbe_FileNotFound(t *testing.T) {
	skipIfFFmpegUnavailable(t)

	_, _, _, err := probe("/nonexistent/file.png")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// failingLoader always returns an error to force the ffmpeg fallback path.
type failingLoader struct{}

func (l *failingLoader) Load(_ string) (*domain.ImageEntity, error) {
	return nil, fmt.Errorf("unsupported format")
}
