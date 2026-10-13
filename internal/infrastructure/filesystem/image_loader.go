package filesystem

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"time"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"

	"github.com/flexphere/gaze/internal/domain"
)

// GIF delays are stored in 1/100s. Browsers treat delays of 0 or 1 as 100ms
// because many GIFs in the wild rely on that; we do the same so such GIFs do
// not spin the tick loop.
const (
	gifDelayUnit         = 10 * time.Millisecond
	minGIFDelayUnits     = 2
	defaultGIFFrameDelay = 100 * time.Millisecond
)

// ImageLoader loads images from the filesystem.
type ImageLoader struct{}

// NewImageLoader creates a new ImageLoader.
func NewImageLoader() *ImageLoader {
	return &ImageLoader{}
}

// Load reads and decodes an image file.
// Supports PNG, JPEG, GIF, BMP, TIFF, WebP via image.Decode.
// Animated GIFs are additionally decoded into composited frames.
func (l *ImageLoader) Load(path string) (*domain.ImageEntity, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path is user-provided CLI argument
	if err != nil {
		return nil, fmt.Errorf("opening image file: %w", err)
	}

	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decoding image: %w", err)
	}

	if format == "gif" {
		// A GIF whose later frames are broken can still be shown as a static
		// image, so a DecodeAll failure falls back to the first frame.
		g, err := gif.DecodeAll(bytes.NewReader(data))
		if err == nil && len(g.Image) > 1 {
			return domain.NewAnimatedImageEntity(compositeGIFFrames(g), path, format), nil
		}
	}

	return domain.NewImageEntity(img, path, format), nil
}

// compositeGIFFrames renders each GIF frame onto a full logical-screen canvas,
// applying the disposal method of the previous frame, so every returned frame
// has the same size and can be displayed on its own.
func compositeGIFFrames(g *gif.GIF) []domain.Frame {
	canvas := image.NewRGBA(gifScreenBounds(g))
	frames := make([]domain.Frame, 0, len(g.Image))

	for i, src := range g.Image {
		disposal := byte(gif.DisposalNone)
		if i < len(g.Disposal) {
			disposal = g.Disposal[i]
		}

		var restore *image.RGBA
		if disposal == gif.DisposalPrevious {
			restore = cloneRGBA(canvas)
		}

		draw.Draw(canvas, src.Bounds(), src, src.Bounds().Min, draw.Over)

		delay := 0
		if i < len(g.Delay) {
			delay = g.Delay[i]
		}
		frames = append(frames, domain.Frame{
			Image: cloneRGBA(canvas),
			Delay: gifFrameDelay(delay),
		})

		switch disposal {
		case gif.DisposalBackground:
			draw.Draw(canvas, src.Bounds(), image.Transparent, image.Point{}, draw.Src)
		case gif.DisposalPrevious:
			canvas = restore
		}
	}

	return frames
}

// gifScreenBounds returns the logical screen bounds. Some encoders leave the
// logical screen size as zero, in which case the union of frame bounds is used.
func gifScreenBounds(g *gif.GIF) image.Rectangle {
	if g.Config.Width > 0 && g.Config.Height > 0 {
		return image.Rect(0, 0, g.Config.Width, g.Config.Height)
	}
	var r image.Rectangle
	for _, src := range g.Image {
		r = r.Union(src.Bounds())
	}
	return image.Rect(0, 0, r.Max.X, r.Max.Y)
}

func gifFrameDelay(units int) time.Duration {
	if units < minGIFDelayUnits {
		return defaultGIFFrameDelay
	}
	return time.Duration(units) * gifDelayUnit
}

func cloneRGBA(src *image.RGBA) *image.RGBA {
	dst := image.NewRGBA(src.Bounds())
	copy(dst.Pix, src.Pix)
	return dst
}
