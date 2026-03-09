package domain

import (
	"image"
	"math"
)

// Viewport represents the visible region of the image.
type Viewport struct {
	OffsetX   float64 // horizontal offset in source image pixels
	OffsetY   float64 // vertical offset in source image pixels
	ZoomLevel float64 // current zoom level; fit-to-window level is stored in fitZoom and may be < 1.0

	TermWidth       int     // terminal width in cells
	TermHeight      int     // terminal height in cells (minus status bar)
	ImgWidth        int     // source image width in pixels
	ImgHeight       int     // source image height in pixels
	CellAspectRatio float64 // cell height / cell width (e.g. 2.0 means cell is twice as tall as wide)

	fitZoom  float64 // zoom level calculated by FitToWindow
	minZoom  float64
	maxZoom  float64
	zoomStep float64
	panStep  float64
}

const (
	defaultCellAspectRatio = 2.0
	minCellAspectRatio     = 0.1
	maxCellAspectRatio     = 10.0
)

// NewViewport creates a new Viewport with the given configuration.
func NewViewport(cfg ViewportConfig) *Viewport {
	return &Viewport{
		ZoomLevel:       1.0,
		CellAspectRatio: defaultCellAspectRatio,
		minZoom:         cfg.MinZoom,
		maxZoom:         cfg.MaxZoom,
		zoomStep:        cfg.ZoomStep,
		panStep:         cfg.PanStep,
	}
}

// SetImageSize updates source image dimensions and resets the view.
func (v *Viewport) SetImageSize(w, h int) {
	v.ImgWidth = w
	v.ImgHeight = h
	v.FitToWindow()
}

// SetTerminalSize updates terminal dimensions and recalculates fit zoom.
func (v *Viewport) SetTerminalSize(w, h int) {
	wasAtFit := v.isAtFitLevel()
	v.TermWidth = w
	v.TermHeight = h
	if wasAtFit {
		v.FitToWindow()
	} else {
		v.Clamp()
	}
}

// isAtFitLevel returns true when the viewport zoom is at the fit-to-window level.
// Unlike !IsZoomed(), this does not match zoomed-out states below fitZoom.
func (v *Viewport) isAtFitLevel() bool {
	fit := v.fitZoom
	if fit <= 0 {
		fit = 1.0
	}
	return math.Abs(v.ZoomLevel-fit) <= fit*0.001
}

// SetCellAspectRatio sets the cell height-to-width ratio.
// Values are clamped to [0.1, 10.0].
func (v *Viewport) SetCellAspectRatio(ratio float64) {
	v.CellAspectRatio = clampFloat(ratio, minCellAspectRatio, maxCellAspectRatio)
}

// CellAspect returns the effective cell aspect ratio, defaulting to 2.0 if unset.
func (v *Viewport) CellAspect() float64 {
	if v.CellAspectRatio <= 0 {
		return defaultCellAspectRatio
	}
	return v.CellAspectRatio
}

// VisibleWidth returns the width of the visible region in source pixels.
func (v *Viewport) VisibleWidth() float64 {
	if v.ZoomLevel <= 0 {
		return float64(v.ImgWidth)
	}
	return float64(v.ImgWidth) / v.ZoomLevel
}

// VisibleHeight returns the height of the visible region in source pixels.
// It accounts for the terminal cell aspect ratio so that the source rectangle
// matches the physical aspect ratio of the terminal display area.
func (v *Viewport) VisibleHeight() float64 {
	if v.ZoomLevel <= 0 || v.TermWidth <= 0 || v.TermHeight <= 0 {
		return float64(v.ImgHeight)
	}
	vw := v.VisibleWidth()
	// termPixelHeight / termPixelWidth = (TermHeight * cellH) / (TermWidth * cellW)
	//                                  = (TermHeight * cellAspect) / TermWidth
	return vw * float64(v.TermHeight) * v.CellAspect() / float64(v.TermWidth)
}

// VisibleRect returns the source image rectangle visible in the viewport.
func (v *Viewport) VisibleRect() image.Rectangle {
	vw := v.VisibleWidth()
	vh := v.VisibleHeight()

	x0 := int(math.Round(v.OffsetX))
	y0 := int(math.Round(v.OffsetY))
	x1 := int(math.Round(v.OffsetX + vw))
	y1 := int(math.Round(v.OffsetY + vh))

	// Clamp to image bounds
	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	if x1 > v.ImgWidth {
		x1 = v.ImgWidth
	}
	if y1 > v.ImgHeight {
		y1 = v.ImgHeight
	}

	return image.Rect(x0, y0, x1, y1)
}

// Zoom adjusts zoom level by delta, centering on the current viewport center.
func (v *Viewport) Zoom(delta float64) {
	centerX := v.OffsetX + v.VisibleWidth()/2
	centerY := v.OffsetY + v.VisibleHeight()/2

	v.ZoomLevel *= (1 + delta)
	v.ZoomLevel = clampFloat(v.ZoomLevel, v.minZoom, v.maxZoom)

	v.OffsetX = centerX - v.VisibleWidth()/2
	v.OffsetY = centerY - v.VisibleHeight()/2
	v.Clamp()
}

// ZoomIn increases zoom level by the configured step.
func (v *Viewport) ZoomIn() {
	v.Zoom(v.zoomStep)
}

// ZoomOut decreases zoom level by the configured step.
func (v *Viewport) ZoomOut() {
	v.Zoom(-v.zoomStep)
}

// ZoomAt adjusts zoom level centered on a specific terminal cell position.
func (v *Viewport) ZoomAt(delta float64, termX, termY int) {
	if v.TermWidth <= 0 || v.TermHeight <= 0 {
		return
	}

	// Convert terminal position to source image coordinates
	srcX := v.OffsetX + float64(termX)/float64(v.TermWidth)*v.VisibleWidth()
	srcY := v.OffsetY + float64(termY)/float64(v.TermHeight)*v.VisibleHeight()

	v.ZoomLevel *= (1 + delta)
	v.ZoomLevel = clampFloat(v.ZoomLevel, v.minZoom, v.maxZoom)

	// Recalculate offset so the point under cursor stays fixed
	v.OffsetX = srcX - float64(termX)/float64(v.TermWidth)*v.VisibleWidth()
	v.OffsetY = srcY - float64(termY)/float64(v.TermHeight)*v.VisibleHeight()
	v.Clamp()
}

// Pan moves the viewport by dx, dy in source pixels.
func (v *Viewport) Pan(dx, dy float64) {
	v.OffsetX += dx
	v.OffsetY += dy
	v.Clamp()
}

// PanByStep moves viewport by panStep percentage of the visible area.
func (v *Viewport) PanByStep(dirX, dirY int) {
	dx := float64(dirX) * v.VisibleWidth() * v.panStep
	dy := float64(dirY) * v.VisibleHeight() * v.panStep
	v.Pan(dx, dy)
}

// FitToWindow resets zoom and offset to show the entire image.
// It calculates the zoom level so the entire image fits within the terminal,
// accounting for the cell aspect ratio.
func (v *Viewport) FitToWindow() {
	v.OffsetX = 0
	v.OffsetY = 0

	if v.TermWidth <= 0 || v.TermHeight <= 0 || v.ImgWidth <= 0 || v.ImgHeight <= 0 {
		v.ZoomLevel = 1.0
		v.fitZoom = v.ZoomLevel
		return
	}

	// At zoom z, visible width = ImgWidth/z, visible height = (ImgWidth/z) * (TermHeight*cellAspect/TermWidth)
	// For the image to fit:
	//   VisibleWidth >= ImgWidth  => z <= 1.0
	//   VisibleHeight >= ImgHeight => (ImgWidth/z) * (TermHeight*cellAspect/TermWidth) >= ImgHeight
	//                               => z <= ImgWidth * TermHeight * cellAspect / (TermWidth * ImgHeight)
	zoomW := 1.0
	zoomH := float64(v.ImgWidth) * float64(v.TermHeight) * v.CellAspect() / (float64(v.TermWidth) * float64(v.ImgHeight))

	v.ZoomLevel = math.Min(zoomW, zoomH)
	if v.ZoomLevel < v.minZoom {
		v.ZoomLevel = v.minZoom
	}
	v.fitZoom = v.ZoomLevel
	v.Clamp()
}

// Clamp ensures offset stays within valid bounds.
func (v *Viewport) Clamp() {
	vw := v.VisibleWidth()
	vh := v.VisibleHeight()

	maxOffsetX := float64(v.ImgWidth) - vw
	maxOffsetY := float64(v.ImgHeight) - vh

	if maxOffsetX < 0 {
		// Image fits entirely; center it
		v.OffsetX = maxOffsetX / 2
	} else {
		v.OffsetX = clampFloat(v.OffsetX, 0, maxOffsetX)
	}

	if maxOffsetY < 0 {
		v.OffsetY = maxOffsetY / 2
	} else {
		v.OffsetY = clampFloat(v.OffsetY, 0, maxOffsetY)
	}
}

// ZoomPercentage returns the current zoom as a display percentage relative to fit level.
func (v *Viewport) ZoomPercentage() int {
	fit := v.fitZoom
	if fit <= 0 {
		fit = 1.0
	}
	return int(math.Round(v.ZoomLevel / fit * 100))
}

// IsZoomed returns true when the viewport is zoomed in beyond fit-to-window.
func (v *Viewport) IsZoomed() bool {
	fit := v.fitZoom
	if fit <= 0 {
		fit = 1.0
	}
	return v.ZoomLevel > fit*1.001
}

func clampFloat(val, min, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
