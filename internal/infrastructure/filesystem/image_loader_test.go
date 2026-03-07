package filesystem

import (
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
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
	case "bmp":
		if err := bmp.Encode(f, img); err != nil {
			t.Fatalf("encoding BMP: %v", err)
		}
	case "tiff":
		if err := tiff.Encode(f, img, nil); err != nil {
			t.Fatalf("encoding TIFF: %v", err)
		}
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

func TestImageLoader_Load_BMP(t *testing.T) {
	dir := t.TempDir()
	path := createTestImage(t, dir, "test.bmp", "bmp", 60, 40)

	loader := NewImageLoader()
	entity, err := loader.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entity.Width != 60 {
		t.Errorf("Width = %d, want 60", entity.Width)
	}
	if entity.Height != 40 {
		t.Errorf("Height = %d, want 40", entity.Height)
	}
	if entity.Format != "bmp" {
		t.Errorf("Format = %q, want %q", entity.Format, "bmp")
	}
}

func TestImageLoader_Load_TIFF(t *testing.T) {
	dir := t.TempDir()
	path := createTestImage(t, dir, "test.tiff", "tiff", 70, 55)

	loader := NewImageLoader()
	entity, err := loader.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entity.Width != 70 {
		t.Errorf("Width = %d, want 70", entity.Width)
	}
	if entity.Height != 55 {
		t.Errorf("Height = %d, want 55", entity.Height)
	}
	if entity.Format != "tiff" {
		t.Errorf("Format = %q, want %q", entity.Format, "tiff")
	}
}

func TestImageLoader_Load_WebP(t *testing.T) {
	loader := NewImageLoader()
	entity, err := loader.Load("../../../testdata/test.webp")
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
