package usecase

import (
	"fmt"
	"math"

	"github.com/flexphere/gaze/internal/domain"
)

// RenderFrameUseCase produces a renderable string for the current viewport.
type RenderFrameUseCase interface {
	Execute(img *domain.ImageEntity, vp *domain.Viewport) (string, error)
	SetMinimapEnabled(enabled bool)
	MinimapEnabled() bool
	// SetFrame selects which animation frame the next Execute renders.
	SetFrame(index int)
}

const (
	minMinimapCols = 5
	minMinimapRows = 3
)

type renderFrameUseCase struct {
	renderer        RendererPort
	minimapCfg      domain.MinimapConfig
	uploaded        bool
	minimapUploaded bool
	minimapShown    bool

	// frame is the animation frame to render; uploadedFrame is the one the
	// terminal currently holds. Upload sends Frames[0] (Source) initially.
	frame         int
	uploadedFrame int
}

// NewRenderFrameUseCase creates a new RenderFrameUseCase.
func NewRenderFrameUseCase(renderer RendererPort, minimapCfg domain.MinimapConfig) RenderFrameUseCase {
	return &renderFrameUseCase{renderer: renderer, minimapCfg: minimapCfg}
}

func (uc *renderFrameUseCase) Execute(img *domain.ImageEntity, vp *domain.Viewport) (string, error) {
	if !uc.uploaded {
		if err := uc.renderer.Upload(img); err != nil {
			return "", fmt.Errorf("uploading image: %w", err)
		}
		uc.uploaded = true
		uc.uploadedFrame = 0
	}

	if img.IsAnimated() && uc.frame != uc.uploadedFrame && uc.frame < len(img.Frames) {
		if err := uc.renderer.UploadFrame(img.Frames[uc.frame].Image); err != nil {
			return "", fmt.Errorf("uploading animation frame: %w", err)
		}
		uc.uploadedFrame = uc.frame
	}

	output, err := uc.renderer.Display(vp)
	if err != nil {
		return "", fmt.Errorf("displaying frame: %w", err)
	}

	// Append minimap when zoomed in and enabled
	shouldShowMinimap := false
	if uc.minimapCfg.Enabled && vp.IsZoomed() {
		minimapCols, minimapRows := uc.minimapSize(vp)
		if minimapCols >= minMinimapCols && minimapRows >= minMinimapRows {
			shouldShowMinimap = true
			if !uc.minimapUploaded {
				if err := uc.renderer.UploadMinimap(img, minimapCols, minimapRows, vp.CellAspect()); err != nil {
					return "", fmt.Errorf("uploading minimap: %w", err)
				}
				uc.minimapUploaded = true
			}

			mmOutput, err := uc.renderer.DisplayMinimap(vp, minimapCols, minimapRows, uc.minimapCfg.BorderColor)
			if err != nil {
				return "", fmt.Errorf("displaying minimap: %w", err)
			}
			output += mmOutput
		}
	}

	// Clear minimap when transitioning from shown to hidden
	if uc.minimapShown && !shouldShowMinimap {
		if err := uc.renderer.ClearMinimap(); err != nil {
			return "", fmt.Errorf("clearing minimap: %w", err)
		}
	}
	uc.minimapShown = shouldShowMinimap

	return output, nil
}

func (uc *renderFrameUseCase) SetMinimapEnabled(enabled bool) {
	uc.minimapCfg.Enabled = enabled
}

func (uc *renderFrameUseCase) MinimapEnabled() bool {
	return uc.minimapCfg.Enabled
}

func (uc *renderFrameUseCase) SetFrame(index int) {
	uc.frame = index
}

// minimapSize calculates the minimap display size in terminal cells.
func (uc *renderFrameUseCase) minimapSize(vp *domain.Viewport) (cols, rows int) {
	cols = int(math.Round(float64(vp.TermWidth) * uc.minimapCfg.Size))
	if cols < 1 {
		cols = 1
	}

	// Preserve image aspect ratio using actual cell aspect ratio
	imgAspect := float64(vp.ImgWidth) / float64(vp.ImgHeight)
	rows = int(math.Round(float64(cols) / imgAspect / vp.CellAspect()))
	if rows < 1 {
		rows = 1
	}

	return cols, rows
}
