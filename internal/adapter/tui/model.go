package tui

import (
	"time"

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

	// Video playback state
	videoDecoder usecase.VideoDecoderPort
	videoInfo    *domain.VideoInfo
	playing      bool
	position     time.Duration
}

// NewModel creates a new TUI model for image viewing.
func NewModel(
	img *domain.ImageEntity,
	cfg *domain.Config,
	vpCtrl usecase.ViewportControlUseCase,
	renderFrame usecase.RenderFrameUseCase,
) Model {
	vp := domain.NewViewport(cfg.Viewport)
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

// NewVideoModel creates a new TUI model for video playback.
func NewVideoModel(
	firstFrame *domain.ImageEntity,
	decoder usecase.VideoDecoderPort,
	videoInfo *domain.VideoInfo,
	cfg *domain.Config,
	renderFrame usecase.RenderFrameUseCase,
) Model {
	vp := domain.NewViewport(cfg.Viewport)
	vp.SetImageSize(videoInfo.Width, videoInfo.Height)

	renderFrame.SetMinimapEnabled(false)

	return Model{
		image:        firstFrame,
		viewport:     vp,
		config:       cfg,
		keymap:       NewKeyMap(cfg.KeyBindings),
		renderFrame:  renderFrame,
		videoDecoder: decoder,
		videoInfo:    videoInfo,
		playing:      true,
	}
}

func (m Model) isVideoMode() bool {
	return m.videoDecoder != nil
}
