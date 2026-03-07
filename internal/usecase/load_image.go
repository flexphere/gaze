package usecase

import "github.com/flexphere/gaze/internal/domain"

// LoadImageUseCase loads an image file.
type LoadImageUseCase interface {
	Execute(path string) (*domain.ImageEntity, error)
}

type loadImageUseCase struct {
	loader ImageLoaderPort
}

// NewLoadImageUseCase creates a new LoadImageUseCase.
func NewLoadImageUseCase(loader ImageLoaderPort) LoadImageUseCase {
	return &loadImageUseCase{loader: loader}
}

func (uc *loadImageUseCase) Execute(path string) (*domain.ImageEntity, error) {
	return uc.loader.Load(path)
}
