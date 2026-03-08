package renderer

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"strings"
	"sync/atomic"

	"golang.org/x/image/draw"

	"github.com/flexphere/gaze/internal/domain"
)

var imageIDCounter uint32

// KittyRenderer implements RendererPort using the Kitty Graphics Protocol.
type KittyRenderer struct {
	imageID      uint32
	imgW         int
	imgH         int
	minimapID    uint32
	minimapBase  *image.RGBA // downscaled thumbnail (reused each frame)
	minimapFrame *image.RGBA // reusable work buffer for compositing
	minimapW     int         // minimap image width in pixels
	minimapH     int         // minimap image height in pixels
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

	return r.uploadImage(r.imageID, img.Source)
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

// UploadMinimap creates a downscaled thumbnail base image for the minimap.
// The base is kept in memory; actual upload happens in DisplayMinimap each frame
// (with the viewport indicator rectangle drawn on top).
func (r *KittyRenderer) UploadMinimap(img *domain.ImageEntity, cols, rows int) error {
	// Delete existing minimap image from terminal before assigning a new ID
	if r.minimapID > 0 {
		fmt.Printf("\x1b_Ga=d,d=i,i=%d\x1b\\", r.minimapID)
	}
	r.minimapID = atomic.AddUint32(&imageIDCounter, 1)

	// Calculate pixel dimensions for the minimap.
	// Use cols*8 and rows*16 as approximate pixel sizes per cell.
	pixW := cols * 8
	pixH := rows * 16

	// Preserve aspect ratio within the target area
	imgAspect := float64(img.Width) / float64(img.Height)
	targetAspect := float64(pixW) / float64(pixH)

	if imgAspect > targetAspect {
		pixH = int(math.Round(float64(pixW) / imgAspect))
	} else {
		pixW = int(math.Round(float64(pixH) * imgAspect))
	}

	if pixW < 1 {
		pixW = 1
	}
	if pixH < 1 {
		pixH = 1
	}

	r.minimapW = pixW
	r.minimapH = pixH

	// Downscale image using high-quality CatmullRom interpolation
	r.minimapBase = image.NewRGBA(image.Rect(0, 0, pixW, pixH))
	draw.CatmullRom.Scale(r.minimapBase, r.minimapBase.Bounds(), img.Source, img.Source.Bounds(), draw.Over, nil)

	// Allocate reusable work buffer for per-frame compositing
	r.minimapFrame = image.NewRGBA(image.Rect(0, 0, pixW, pixH))

	return nil
}

// DisplayMinimap composites the viewport indicator onto the minimap base,
// and returns all escape sequences (delete + upload + placement) as a single string
// so they are written atomically through Bubbletea's View() output.
func (r *KittyRenderer) DisplayMinimap(vp *domain.Viewport, cols, rows int, borderColor string) (string, error) {
	if r.minimapID == 0 || r.minimapBase == nil || cols <= 0 || rows <= 0 {
		return "", nil
	}

	imgW := float64(vp.ImgWidth)
	imgH := float64(vp.ImgHeight)
	if imgW <= 0 || imgH <= 0 {
		return "", nil
	}

	// Reuse work buffer instead of allocating each frame
	copy(r.minimapFrame.Pix, r.minimapBase.Pix)

	// Calculate indicator rectangle in pixel coordinates
	visRect := vp.VisibleRect()
	pxLeft := int(math.Round(float64(visRect.Min.X) / imgW * float64(r.minimapW)))
	pxTop := int(math.Round(float64(visRect.Min.Y) / imgH * float64(r.minimapH)))
	pxRight := int(math.Round(float64(visRect.Max.X) / imgW * float64(r.minimapW)))
	pxBottom := int(math.Round(float64(visRect.Max.Y) / imgH * float64(r.minimapH)))

	pxLeft = clampInt(pxLeft, 0, r.minimapW-1)
	pxTop = clampInt(pxTop, 0, r.minimapH-1)
	pxRight = clampInt(pxRight, 1, r.minimapW)
	pxBottom = clampInt(pxBottom, 1, r.minimapH)

	// Ensure the rectangle has at least 1px width/height after clamping
	if pxRight <= pxLeft {
		pxRight = pxLeft + 1
	}
	if pxBottom <= pxTop {
		pxBottom = pxTop + 1
	}
	// Re-clamp to bounds
	if pxRight > r.minimapW {
		pxRight = r.minimapW
	}
	if pxBottom > r.minimapH {
		pxBottom = r.minimapH
	}

	// Draw indicator rectangle border with configured color and dark outline
	border := parseHexColor(borderColor)
	drawRect(r.minimapFrame, pxLeft, pxTop, pxRight, pxBottom, border)

	// Build all escape sequences into one string so they are output atomically.
	// This prevents timing issues between delete/upload and placement.
	var out strings.Builder

	// 1. Delete previous minimap image
	fmt.Fprintf(&out, "\x1b_Ga=d,d=i,i=%d\x1b\\", r.minimapID)

	// 2. Upload new minimap frame
	uploadSeq, err := buildUploadSequence(r.minimapID, r.minimapFrame)
	if err != nil {
		return "", fmt.Errorf("encoding minimap frame: %w", err)
	}
	out.WriteString(uploadSeq)

	// 3. Position and place minimap at bottom-right (above status bar)
	startCol := vp.TermWidth - cols + 1  // 1-based
	startRow := vp.TermHeight - rows + 1 // 1-based

	if startCol < 1 {
		startCol = 1
	}
	if startRow < 1 {
		startRow = 1
	}

	fmt.Fprintf(&out, "\x1b[%d;%dH", startRow, startCol)
	// z=1 ensures minimap is rendered above the main image (z=0 default)
	fmt.Fprintf(&out, "\x1b_Ga=p,i=%d,c=%d,r=%d,z=1,q=2\x1b\\", r.minimapID, cols, rows)

	return out.String(), nil
}

// ClearMinimap removes the minimap image from the terminal.
func (r *KittyRenderer) ClearMinimap() error {
	if r.minimapID > 0 {
		fmt.Printf("\x1b_Ga=d,d=i,i=%d\x1b\\", r.minimapID)
	}
	return nil
}

// drawRect draws a visible indicator rectangle on the image.
// Uses the given border color with a dark outline for contrast.
func drawRect(img *image.RGBA, left, top, right, bottom int, border color.RGBA) {
	outline := color.RGBA{R: 0, G: 0, B: 0, A: 200}

	// Draw outline (1px outside the border)
	drawRectBorder(img, left-1, top-1, right+1, bottom+1, outline)
	// Draw main border
	drawRectBorder(img, left, top, right, bottom, border)
}

// parseHexColor parses a hex color string like "#FF0000" into a color.RGBA.
// Falls back to white on invalid input.
func parseHexColor(hex string) color.RGBA {
	c := color.RGBA{R: 255, G: 255, B: 255, A: 230}
	if len(hex) == 0 {
		return c
	}
	if hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return c
	}
	r, ok1 := parseHexByte(hex[0:2])
	g, ok2 := parseHexByte(hex[2:4])
	b, ok3 := parseHexByte(hex[4:6])
	if !ok1 || !ok2 || !ok3 {
		return c
	}
	return color.RGBA{R: r, G: g, B: b, A: 230}
}

func parseHexByte(s string) (byte, bool) {
	var val byte
	for _, c := range s {
		val <<= 4
		switch {
		case c >= '0' && c <= '9':
			val |= byte(c - '0')
		case c >= 'a' && c <= 'f':
			val |= byte(c - 'a' + 10)
		case c >= 'A' && c <= 'F':
			val |= byte(c - 'A' + 10)
		default:
			return 0, false
		}
	}
	return val, true
}

// drawRectBorder draws a 1px rectangle border on the image.
func drawRectBorder(img *image.RGBA, left, top, right, bottom int, c color.RGBA) {
	bounds := img.Bounds()

	for x := left; x < right; x++ {
		if x >= bounds.Min.X && x < bounds.Max.X {
			if top >= bounds.Min.Y && top < bounds.Max.Y {
				img.SetRGBA(x, top, c)
			}
			if bottom-1 >= bounds.Min.Y && bottom-1 < bounds.Max.Y {
				img.SetRGBA(x, bottom-1, c)
			}
		}
	}
	for y := top; y < bottom; y++ {
		if y >= bounds.Min.Y && y < bounds.Max.Y {
			if left >= bounds.Min.X && left < bounds.Max.X {
				img.SetRGBA(left, y, c)
			}
			if right-1 >= bounds.Min.X && right-1 < bounds.Max.X {
				img.SetRGBA(right-1, y, c)
			}
		}
	}
}

// buildUploadSequence creates the Kitty upload escape sequences as a string
// instead of writing directly to stdout.
func buildUploadSequence(id uint32, img image.Image) (string, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("encoding image to PNG: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())

	var out strings.Builder
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
			fmt.Fprintf(&out, "\x1b_Gi=%d,f=100,a=t,t=d,m=%d;%s\x1b\\", id, more, chunk)
		} else {
			fmt.Fprintf(&out, "\x1b_Gi=%d,m=%d;%s\x1b\\", id, more, chunk)
		}
	}

	return out.String(), nil
}

// uploadImage encodes and transmits an image to the terminal.
func (r *KittyRenderer) uploadImage(id uint32, img image.Image) error {
	seq, err := buildUploadSequence(id, img)
	if err != nil {
		return err
	}
	fmt.Print(seq)
	return nil
}

func clampInt(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
