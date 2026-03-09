package tui

import (
	"github.com/flexphere/gaze/internal/domain"
	"github.com/flexphere/gaze/internal/usecase"
)

// Model is the Bubbletea model for the image viewer.
type Model struct {
	image       *domain.ImageEntity
	viewport    *domain.Viewport
	config      *domain.Config
	keymap      KeyMap
	vpCtrl      usecase.ViewportControlUseCase
	renderFrame usecase.RenderFrameUseCase

	renderedFrame string
	err           error
	ready         bool

	// Mouse drag state
	dragging       bool
	dragStartTermX int
	dragStartTermY int
	dragStartOffX  float64
	dragStartOffY  float64
}

// NewModel creates a new TUI model.
func NewModel(
	img *domain.ImageEntity,
	cfg *domain.Config,
	vpCtrl usecase.ViewportControlUseCase,
	renderFrame usecase.RenderFrameUseCase,
) Model {
	vp := domain.NewViewport(cfg.Viewport)
	cellW, cellH := queryCellSize()
	vp.SetCellAspectRatio(cellH / cellW)
	vp.SetImageSize(img.Width, img.Height)

	return Model{
		image:       img,
		viewport:    vp,
		config:      cfg,
		keymap:      NewKeyMap(cfg.KeyBindings),
		vpCtrl:      vpCtrl,
		renderFrame: renderFrame,
	}
}
