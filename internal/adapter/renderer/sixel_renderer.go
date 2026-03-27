package renderer

import (
	"fmt"
	"image"
	"image/color"
	stddraw "image/draw"
	"math"
	"sort"
	"strings"

	xdraw "golang.org/x/image/draw"

	"github.com/flexphere/gaze/internal/domain"
)

// SixelRenderer implements RendererPort using the Sixel graphics protocol.
// Unlike Kitty, Sixel re-encodes image data each frame (no persistent upload).
type SixelRenderer struct {
	img   image.Image
	imgW  int
	imgH  int
	cellW float64 // cell width in pixels
	cellH float64 // cell height in pixels

	minimapBase  *image.RGBA
	minimapFrame *image.RGBA
	minimapW     int
	minimapH     int

	// Minimap indicator cache to skip re-encode when unchanged.
	prevIndicator   [4]int
	prevBorderColor string
	prevCached      bool
	prevSixel       string
}

// NewSixelRenderer creates a new SixelRenderer.
// cellW and cellH are the terminal cell dimensions in pixels, used to calculate
// the output pixel size for Sixel encoding.
func NewSixelRenderer(cellW, cellH float64) *SixelRenderer {
	return &SixelRenderer{cellW: cellW, cellH: cellH}
}

// Upload stores the image in memory for later rendering.
// Sixel has no persistent terminal-side image storage, so this only caches locally.
func (r *SixelRenderer) Upload(img *domain.ImageEntity) error {
	r.img = img.Source
	r.imgW = img.Width
	r.imgH = img.Height
	return nil
}

// Display generates the Sixel escape sequence for the given viewport.
func (r *SixelRenderer) Display(vp *domain.Viewport) (string, error) {
	if r.img == nil {
		return "", nil
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

	// Calculate display columns/rows preserving aspect ratio (same logic as Kitty)
	cellAspect := vp.CellAspect()
	imgAspect := float64(srcW) / float64(srcH)
	termAspect := float64(cols) / (float64(rows) * cellAspect)

	var displayCols, displayRows int
	if imgAspect > termAspect {
		displayCols = cols
		displayRows = int(math.Round(float64(cols) / imgAspect / cellAspect))
	} else {
		displayRows = rows
		displayCols = int(math.Round(float64(rows) * cellAspect * imgAspect))
	}
	if displayCols <= 0 {
		displayCols = 1
	}
	if displayRows <= 0 {
		displayRows = 1
	}

	// Compute pixel dimensions for Sixel output
	pixW := int(math.Round(float64(displayCols) * r.cellW))
	pixH := int(math.Round(float64(displayRows) * r.cellH))
	if pixW <= 0 {
		pixW = 1
	}
	if pixH <= 0 {
		pixH = 1
	}

	scaled := scaleRegion(r.img, rect, pixW, pixH)

	return "\x1b[H" + encodeSixel(scaled), nil
}

// Clear is a no-op for Sixel. Images are part of the screen content
// and get overwritten by the next render.
func (r *SixelRenderer) Clear() error {
	return nil
}

// UploadMinimap creates a downscaled thumbnail base image for the minimap.
func (r *SixelRenderer) UploadMinimap(img *domain.ImageEntity, cols, rows int, cellAspect float64) error {
	const baseCellW = 8.0
	cellH := baseCellW * cellAspect
	pixW := int(math.Round(float64(cols) * baseCellW))
	pixH := int(math.Round(float64(rows) * cellH))

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
	r.minimapBase = image.NewRGBA(image.Rect(0, 0, pixW, pixH))
	xdraw.CatmullRom.Scale(r.minimapBase, r.minimapBase.Bounds(), img.Source, img.Source.Bounds(), xdraw.Over, nil)
	r.minimapFrame = image.NewRGBA(image.Rect(0, 0, pixW, pixH))

	r.prevIndicator = [4]int{}
	r.prevBorderColor = ""
	r.prevCached = false
	r.prevSixel = ""

	return nil
}

// DisplayMinimap composites the viewport indicator onto the minimap and
// returns Sixel output positioned at the bottom-right corner.
func (r *SixelRenderer) DisplayMinimap(vp *domain.Viewport, cols, rows int, borderColor string) (string, error) {
	if r.minimapBase == nil || cols <= 0 || rows <= 0 {
		return "", nil
	}

	imgW := float64(vp.ImgWidth)
	imgH := float64(vp.ImgHeight)
	if imgW <= 0 || imgH <= 0 {
		return "", nil
	}

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

	if pxRight <= pxLeft {
		pxRight = pxLeft + 1
	}
	if pxBottom <= pxTop {
		pxBottom = pxTop + 1
	}
	if pxRight > r.minimapW {
		pxRight = r.minimapW
	}
	if pxBottom > r.minimapH {
		pxBottom = r.minimapH
	}

	// Placement position (bottom-right corner)
	startCol := vp.TermWidth - cols + 1 // 1-based
	startRow := vp.TermHeight - rows + 1
	if startCol < 1 {
		startCol = 1
	}
	if startRow < 1 {
		startRow = 1
	}

	cursorPos := fmt.Sprintf("\x1b[%d;%dH", startRow, startCol)

	// Skip re-encode if indicator and border color haven't changed
	indicator := [4]int{pxLeft, pxTop, pxRight, pxBottom}
	if r.prevCached && indicator == r.prevIndicator && borderColor == r.prevBorderColor {
		return cursorPos + r.prevSixel, nil
	}

	// Composite: copy base then draw indicator
	copy(r.minimapFrame.Pix, r.minimapBase.Pix)
	border := parseHexColor(borderColor)
	drawRect(r.minimapFrame, pxLeft, pxTop, pxRight, pxBottom, border)

	sixelData := encodeSixel(r.minimapFrame)

	r.prevIndicator = indicator
	r.prevBorderColor = borderColor
	r.prevCached = true
	r.prevSixel = sixelData

	return cursorPos + sixelData, nil
}

// ClearMinimap invalidates the minimap cache.
func (r *SixelRenderer) ClearMinimap() error {
	r.prevCached = false
	r.prevSixel = ""
	return nil
}

// --- Image helpers ---

// scaleRegion crops a region from the source image and scales it to the target size.
func scaleRegion(src image.Image, rect image.Rectangle, w, h int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, rect, xdraw.Over, nil)
	return dst
}

// --- Sixel encoding ---

// sixelPalette is a pre-computed uniform 216-color palette (6R x 6G x 6B).
var sixelPalette = buildSixelPalette()

func buildSixelPalette() color.Palette {
	const levels = 6
	palette := make(color.Palette, 0, levels*levels*levels)
	for ri := range levels {
		for gi := range levels {
			for bi := range levels {
				palette = append(palette, color.RGBA{
					R: uint8(ri * 255 / (levels - 1)), //nolint:gosec // ri is in [0,5], result fits uint8
					G: uint8(gi * 255 / (levels - 1)), //nolint:gosec // gi is in [0,5], result fits uint8
					B: uint8(bi * 255 / (levels - 1)), //nolint:gosec // bi is in [0,5], result fits uint8
					A: 255,
				})
			}
		}
	}
	return palette
}

// encodeSixel converts an RGBA image to a Sixel escape sequence.
// Uses Floyd-Steinberg dithering with a 216-color uniform palette.
func encodeSixel(img *image.RGBA) string {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	if w <= 0 || h <= 0 {
		return ""
	}

	// Quantize to paletted image using Floyd-Steinberg dithering
	paletted := image.NewPaletted(image.Rect(0, 0, w, h), sixelPalette)
	stddraw.FloydSteinberg.Draw(paletted, paletted.Bounds(), img, bounds.Min)

	var out strings.Builder
	out.Grow(w * h / 2) // rough estimate

	// DCS q with raster attributes (Pan;Pad;Ph;Pv)
	fmt.Fprintf(&out, "\x1bPq\"1;1;%d;%d", w, h)

	// Define palette colors (only those actually used)
	usedGlobal := make([]bool, len(sixelPalette))
	for _, idx := range paletted.Pix[:w*h] {
		usedGlobal[idx] = true
	}
	for i, used := range usedGlobal {
		if !used {
			continue
		}
		c, ok := sixelPalette[i].(color.RGBA)
		if !ok {
			continue
		}
		r := int(math.Round(float64(c.R) / 255.0 * 100.0))
		g := int(math.Round(float64(c.G) / 255.0 * 100.0))
		b := int(math.Round(float64(c.B) / 255.0 * 100.0))
		fmt.Fprintf(&out, "#%d;2;%d;%d;%d", i, r, g, b)
	}

	// Encode pixel data in 6-row bands
	for bandY := 0; bandY < h; bandY += 6 {
		bandH := 6
		if bandY+bandH > h {
			bandH = h - bandY
		}

		// Collect colors used in this band
		usedInBand := make(map[uint8]bool)
		for dy := 0; dy < bandH; dy++ {
			rowOff := (bandY + dy) * paletted.Stride
			for x := 0; x < w; x++ {
				usedInBand[paletted.Pix[rowOff+x]] = true
			}
		}

		sortedColors := make([]uint8, 0, len(usedInBand))
		for ci := range usedInBand {
			sortedColors = append(sortedColors, ci)
		}
		sort.Slice(sortedColors, func(i, j int) bool { return sortedColors[i] < sortedColors[j] })

		// Output each color pass
		for colorIdx, ci := range sortedColors {
			fmt.Fprintf(&out, "#%d", ci)

			var prevChar byte
			runLen := 0

			for x := 0; x < w; x++ {
				var bits byte
				for dy := 0; dy < 6; dy++ {
					if bandY+dy < h {
						rowOff := (bandY + dy) * paletted.Stride
						if paletted.Pix[rowOff+x] == ci {
							bits |= 1 << uint(dy)
						}
					}
				}
				ch := bits + 63

				if runLen > 0 && ch == prevChar {
					runLen++
				} else {
					if runLen > 0 {
						writeSixelRLE(&out, prevChar, runLen)
					}
					prevChar = ch
					runLen = 1
				}
			}
			if runLen > 0 {
				writeSixelRLE(&out, prevChar, runLen)
			}

			if colorIdx < len(sortedColors)-1 {
				out.WriteByte('$') // carriage return within band
			}
		}

		if bandY+6 < h {
			out.WriteByte('-') // new sixel line
		}
	}

	// String terminator
	out.WriteString("\x1b\\")
	return out.String()
}

// writeSixelRLE writes a run-length encoded Sixel character.
// Uses explicit RLE notation (!N<char>) for runs of 4 or more.
func writeSixelRLE(out *strings.Builder, ch byte, count int) {
	if count <= 3 {
		for i := 0; i < count; i++ {
			out.WriteByte(ch)
		}
	} else {
		fmt.Fprintf(out, "!%d%c", count, ch)
	}
}
