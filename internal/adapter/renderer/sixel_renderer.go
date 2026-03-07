package renderer

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"math"

	"github.com/mattn/go-sixel"
	xdraw "golang.org/x/image/draw"

	"github.com/flexphere/gaze/internal/domain"
)

const (
	// Default cell pixel dimensions when terminal pixel size is unknown.
	defaultCellWidth  = 8
	defaultCellHeight = 16
)

// SixelRenderer implements RendererPort using the Sixel graphics protocol.
type SixelRenderer struct {
	img image.Image
}

// NewSixelRenderer creates a new SixelRenderer.
func NewSixelRenderer() *SixelRenderer {
	return &SixelRenderer{}
}

// Upload stores the image in memory for later encoding.
func (r *SixelRenderer) Upload(img *domain.ImageEntity) error {
	r.img = img.Source
	return nil
}

// Display generates Sixel escape sequences for the given viewport.
func (r *SixelRenderer) Display(vp *domain.Viewport) (string, error) {
	if r.img == nil {
		return "", fmt.Errorf("no image uploaded")
	}

	rect := vp.VisibleRect()
	srcW := rect.Dx()
	srcH := rect.Dy()
	if srcW <= 0 || srcH <= 0 {
		return "", nil
	}

	cols := vp.TermWidth
	rows := vp.TermHeight
	if cols <= 0 || rows <= 0 {
		return "", nil
	}

	// Crop visible region from source image
	cropped := image.NewRGBA(image.Rect(0, 0, srcW, srcH))
	draw.Draw(cropped, cropped.Bounds(), r.img, rect.Min, draw.Src)

	// Calculate display pixel size from terminal cells
	pixW := cols * defaultCellWidth
	pixH := rows * defaultCellHeight

	// Preserve aspect ratio
	imgAspect := float64(srcW) / float64(srcH)
	termAspect := float64(pixW) / float64(pixH)

	var displayW, displayH int
	if imgAspect > termAspect {
		displayW = pixW
		displayH = int(math.Round(float64(pixW) / imgAspect))
	} else {
		displayH = pixH
		displayW = int(math.Round(float64(pixH) * imgAspect))
	}

	if displayW <= 0 {
		displayW = 1
	}
	if displayH <= 0 {
		displayH = 1
	}

	// Resize cropped image to display pixel dimensions
	resized := image.NewRGBA(image.Rect(0, 0, displayW, displayH))
	xdraw.BiLinear.Scale(resized, resized.Bounds(), cropped, cropped.Bounds(), xdraw.Src, nil)

	// Encode to Sixel
	var buf bytes.Buffer
	enc := sixel.NewEncoder(&buf)
	enc.Dither = true
	if err := enc.Encode(resized); err != nil {
		return "", fmt.Errorf("encoding sixel: %w", err)
	}

	// Move cursor to top-left and output sixel data
	output := "\x1b[H" + buf.String()
	return output, nil
}

// Clear is a no-op for Sixel (terminal manages cleanup via alt screen).
func (r *SixelRenderer) Clear() error {
	return nil
}
