package usecase

import (
	"fmt"

	"github.com/flexphere/gaze/internal/domain"
)

// RenderFrameUseCase produces a renderable string for the current viewport.
type RenderFrameUseCase interface {
	Execute(img *domain.ImageEntity, vp *domain.Viewport) (string, error)
}

type renderFrameUseCase struct {
	renderer RendererPort
	uploaded bool
}

// NewRenderFrameUseCase creates a new RenderFrameUseCase.
func NewRenderFrameUseCase(renderer RendererPort) RenderFrameUseCase {
	return &renderFrameUseCase{renderer: renderer}
}

func (uc *renderFrameUseCase) Execute(img *domain.ImageEntity, vp *domain.Viewport) (string, error) {
	if !uc.uploaded {
		if err := uc.renderer.Upload(img); err != nil {
			return "", fmt.Errorf("uploading image: %w", err)
		}
		uc.uploaded = true
	}

	output, err := uc.renderer.Display(vp)
	if err != nil {
		return "", fmt.Errorf("displaying frame: %w", err)
	}

	return output, nil
}
