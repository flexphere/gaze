package domain

import (
	"image"
	"time"
)

// Frame is a single fully-composited frame of an animated image.
type Frame struct {
	Image image.Image
	Delay time.Duration
}

// ImageEntity holds the decoded image and its metadata.
type ImageEntity struct {
	Source image.Image
	Width  int
	Height int
	Path   string
	Format string

	// Frames holds all frames of an animated image. Empty for static images.
	Frames []Frame
}

// NewImageEntity creates an ImageEntity from a decoded image.
func NewImageEntity(src image.Image, path, format string) *ImageEntity {
	bounds := src.Bounds()
	return &ImageEntity{
		Source: src,
		Width:  bounds.Dx(),
		Height: bounds.Dy(),
		Path:   path,
		Format: format,
	}
}

// NewAnimatedImageEntity creates an ImageEntity from composited frames.
// Source is set to the first frame so static rendering paths keep working.
// frames must not be empty.
func NewAnimatedImageEntity(frames []Frame, path, format string) *ImageEntity {
	entity := NewImageEntity(frames[0].Image, path, format)
	entity.Frames = frames
	return entity
}

// IsAnimated reports whether the image has more than one frame.
func (e *ImageEntity) IsAnimated() bool {
	return len(e.Frames) > 1
}
