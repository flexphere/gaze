package filesystem

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"

	"github.com/flexphere/gaze/internal/domain"
)

// ImageLoader loads images from the filesystem.
type ImageLoader struct{}

// NewImageLoader creates a new ImageLoader.
func NewImageLoader() *ImageLoader {
	return &ImageLoader{}
}

// Load reads and decodes an image file.
// Supports PNG, JPEG, GIF, BMP, TIFF, WebP via image.Decode.
func (l *ImageLoader) Load(path string) (*domain.ImageEntity, error) {
	f, err := os.Open(path) //nolint:gosec // path is user-provided CLI argument
	if err != nil {
		return nil, fmt.Errorf("opening image file: %w", err)
	}
	defer func() { _ = f.Close() }()

	img, format, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decoding image: %w", err)
	}

	return domain.NewImageEntity(img, path, format), nil
}
