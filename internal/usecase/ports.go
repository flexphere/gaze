package usecase

import (
	"image"
	"time"

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
	Display(vp *domain.Viewport) (string, error)
	Clear() error

	// Minimap methods
	UploadMinimap(img *domain.ImageEntity, cols, rows int) error
	DisplayMinimap(vp *domain.Viewport, cols, rows int, borderColor string) (string, error)
	ClearMinimap() error
}

// VideoDecoderPort decodes video frames sequentially.
type VideoDecoderPort interface {
	Open(path string) (*domain.VideoInfo, error)
	NextFrame() (image.Image, error) // returns io.EOF at end of video
	Seek(pos time.Duration) error
	Close() error
}

// ConfigLoaderPort reads configuration from persistent storage.
type ConfigLoaderPort interface {
	Load() (*domain.Config, error)
}
