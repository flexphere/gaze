package usecase

import (
	"image"

	"github.com/flexphere/gaze/internal/domain"
)

// ImageLoaderPort loads and decodes an image from a path.
type ImageLoaderPort interface {
	Load(path string) (*domain.ImageEntity, error)
}

// RendererPort abstracts image rendering to the terminal.
// Kitty implementation: Upload() sends image data once, Display() specifies source rect.
type RendererPort interface {
	Upload(img *domain.ImageEntity) error
	// UploadFrame replaces the pixel data of the uploaded image with src,
	// keeping the same image ID. Used to advance animation frames.
	UploadFrame(src image.Image) error
	Display(vp *domain.Viewport) (string, error)
	Clear() error

	// Minimap methods
	UploadMinimap(img *domain.ImageEntity, cols, rows int, cellAspect float64) error
	DisplayMinimap(vp *domain.Viewport, cols, rows int, borderColor string) (string, error)
	ClearMinimap() error
}

// ConfigLoaderPort reads configuration from persistent storage.
type ConfigLoaderPort interface {
	Load() (*domain.Config, error)
}
