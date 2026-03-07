package usecase

import "github.com/flexphere/gaze/internal/domain"

// ImageLoaderPort loads and decodes an image from a path.
type ImageLoaderPort interface {
	Load(path string) (*domain.ImageEntity, error)
}

// RendererPort abstracts image rendering to the terminal.
// Kitty implementation: Upload() sends image data once, Display() specifies source rect.
// Future Sixel implementation: Upload() holds image in memory, Display() crops+encodes+sends.
type RendererPort interface {
	Upload(img *domain.ImageEntity) error
	Display(vp *domain.Viewport) (string, error)
	Clear() error
}

// ConfigLoaderPort reads configuration from persistent storage.
type ConfigLoaderPort interface {
	Load() (*domain.Config, error)
}
