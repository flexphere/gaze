package tui

import (
	"image"
	"time"

	"golang.org/x/image/draw"

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
	scaledBuf    *image.RGBA // reusable buffer for video frame scaling
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

const (
	// Approximate pixel sizes per terminal cell for Kitty Graphics Protocol.
	pxPerCol = 8
	pxPerRow = 16
)

// scaleVideoFrame scales a video frame to fit terminal display resolution.
// This avoids expensive PNG encoding of full-resolution frames.
func (m *Model) scaleVideoFrame(src image.Image) *domain.ImageEntity {
	if m.viewport.TermWidth <= 0 || m.viewport.TermHeight <= 0 {
		return domain.NewImageEntity(src, m.image.Path, "video")
	}

	maxW := m.viewport.TermWidth * pxPerCol
	maxH := m.viewport.TermHeight * pxPerRow

	srcBounds := src.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	// Skip scaling if already small enough
	if srcW <= maxW && srcH <= maxH {
		return domain.NewImageEntity(src, m.image.Path, "video")
	}

	// Scale preserving aspect ratio
	scaleX := float64(maxW) / float64(srcW)
	scaleY := float64(maxH) / float64(srcH)
	scale := min(scaleX, scaleY)
	dstW := max(int(float64(srcW)*scale), 1)
	dstH := max(int(float64(srcH)*scale), 1)

	// Reuse buffer if same dimensions
	if m.scaledBuf == nil || m.scaledBuf.Bounds().Dx() != dstW || m.scaledBuf.Bounds().Dy() != dstH {
		m.scaledBuf = image.NewRGBA(image.Rect(0, 0, dstW, dstH))
	}

	draw.ApproxBiLinear.Scale(m.scaledBuf, m.scaledBuf.Bounds(), src, srcBounds, draw.Src, nil)

	// Update viewport to match scaled dimensions
	m.viewport.SetImageSize(dstW, dstH)

	return domain.NewImageEntity(m.scaledBuf, m.image.Path, "video")
}
