package domain

import "image"

// ImageEntity holds the decoded image and its metadata.
type ImageEntity struct {
	Source image.Image
	Width  int
	Height int
	Path   string
	Format string
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
