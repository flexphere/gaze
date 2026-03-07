package filesystem

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/flexphere/gaze/internal/domain"
)

// ImageLoader loads images from the filesystem.
type ImageLoader struct{}

// NewImageLoader creates a new ImageLoader.
func NewImageLoader() *ImageLoader {
	return &ImageLoader{}
}

// Load reads and decodes an image file.
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
