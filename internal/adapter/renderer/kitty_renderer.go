package renderer

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"math"
	"sync/atomic"

	"github.com/flexphere/gaze/internal/domain"
)

var imageIDCounter uint32

// KittyRenderer implements RendererPort using the Kitty Graphics Protocol.
type KittyRenderer struct {
	imageID uint32
	imgW    int
	imgH    int
}

// NewKittyRenderer creates a new KittyRenderer.
func NewKittyRenderer() *KittyRenderer {
	return &KittyRenderer{}
}

// Upload encodes and transmits the image to the terminal via Kitty graphics protocol.
func (r *KittyRenderer) Upload(img *domain.ImageEntity) error {
	r.imageID = atomic.AddUint32(&imageIDCounter, 1)
	r.imgW = img.Width
	r.imgH = img.Height

	// Encode image to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img.Source); err != nil {
		return fmt.Errorf("encoding image to PNG: %w", err)
	}

	// Base64 encode the PNG data
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())

	// Send in chunks (Kitty protocol limits payload per escape sequence)
	const chunkSize = 4096
	for i := 0; i < len(encoded); i += chunkSize {
		end := i + chunkSize
		if end > len(encoded) {
			end = len(encoded)
		}
		chunk := encoded[i:end]

		more := 1
		if end >= len(encoded) {
			more = 0
		}

		if i == 0 {
			// First chunk: include transmission parameters
			fmt.Printf("\x1b_Gi=%d,f=100,a=t,t=d,m=%d;%s\x1b\\", r.imageID, more, chunk)
		} else {
			// Continuation chunk
			fmt.Printf("\x1b_Gi=%d,m=%d;%s\x1b\\", r.imageID, more, chunk)
		}
	}

	return nil
}

// Display generates the Kitty graphics escape sequence for the given viewport.
func (r *KittyRenderer) Display(vp *domain.Viewport) (string, error) {
	rect := vp.VisibleRect()

	srcX := rect.Min.X
	srcY := rect.Min.Y
	srcW := rect.Dx()
	srcH := rect.Dy()

	// Clamp source dimensions to valid range
	if srcW <= 0 || srcH <= 0 {
		return "", nil
	}

	// Terminal display size
	cols := vp.TermWidth
	rows := vp.TermHeight
	if cols <= 0 || rows <= 0 {
		return "", nil
	}

	// Calculate appropriate display columns/rows preserving aspect ratio
	imgAspect := float64(srcW) / float64(srcH)
	termAspect := float64(cols) / (float64(rows) * 2) // cells are ~2:1 height:width

	var displayCols, displayRows int
	if imgAspect > termAspect {
		// Image is wider than terminal area
		displayCols = cols
		displayRows = int(math.Round(float64(cols) / imgAspect / 2))
	} else {
		// Image is taller than terminal area
		displayRows = rows
		displayCols = int(math.Round(float64(rows) * 2 * imgAspect))
	}

	if displayCols <= 0 {
		displayCols = 1
	}
	if displayRows <= 0 {
		displayRows = 1
	}

	// Clear previous display and show new frame
	// Move cursor to top-left, clear screen area, then display
	output := "\x1b[H" // move cursor to top-left
	output += fmt.Sprintf("\x1b_Ga=p,i=%d,x=%d,y=%d,w=%d,h=%d,c=%d,r=%d,q=2\x1b\\",
		r.imageID, srcX, srcY, srcW, srcH, displayCols, displayRows)

	return output, nil
}

// Clear removes the image from the terminal.
func (r *KittyRenderer) Clear() error {
	if r.imageID > 0 {
		fmt.Printf("\x1b_Ga=d,d=i,i=%d\x1b\\", r.imageID)
	}
	return nil
}
