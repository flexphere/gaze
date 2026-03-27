package renderer

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/flexphere/gaze/internal/domain"
)

// TmuxRenderer wraps KittyRenderer, adding DCS passthrough so that
// Kitty graphics escape sequences reach the outer terminal through tmux.
// Cursor coordinates are offset by the pane position so that images
// render inside the correct pane.
type TmuxRenderer struct {
	inner    *KittyRenderer
	paneTop  int
	paneLeft int
}

// NewTmuxRenderer creates a TmuxRenderer wrapping a KittyRenderer.
func NewTmuxRenderer() *TmuxRenderer {
	top, left := queryTmuxPaneOffset()
	return &TmuxRenderer{inner: NewKittyRenderer(), paneTop: top, paneLeft: left}
}

// RefreshPaneOffset re-queries the tmux pane position.
// Call this on window resize to stay in sync with pane layout changes.
func (r *TmuxRenderer) RefreshPaneOffset() {
	r.paneTop, r.paneLeft = queryTmuxPaneOffset()
}

// queryTmuxPaneOffset returns the current pane's top-left corner offset
// within the outer terminal by querying tmux.
func queryTmuxPaneOffset() (top, left int) {
	out, err := exec.Command("tmux", "display-message", "-p", "#{pane_top} #{pane_left}").Output()
	if err != nil {
		return 0, 0
	}
	parts := strings.Fields(strings.TrimSpace(string(out)))
	if len(parts) >= 2 {
		top, _ = strconv.Atoi(parts[0])
		left, _ = strconv.Atoi(parts[1])
	}
	return top, left
}

// Upload encodes and transmits the image with tmux DCS passthrough wrapping.
func (r *TmuxRenderer) Upload(img *domain.ImageEntity) error {
	r.inner.imageID = atomic.AddUint32(&imageIDCounter, 1)
	r.inner.imgW = img.Width
	r.inner.imgH = img.Height

	seq, err := buildUploadSequence(r.inner.imageID, img.Source)
	if err != nil {
		return err
	}
	fmt.Print(r.wrapAllKittySequences(seq))
	return nil
}

// Display returns the Kitty placement sequence wrapped for tmux.
func (r *TmuxRenderer) Display(vp *domain.Viewport) (string, error) {
	seq, err := r.inner.Display(vp)
	if err != nil {
		return "", err
	}
	return r.wrapAllKittySequences(seq), nil
}

// Clear removes the image from the terminal via tmux passthrough.
func (r *TmuxRenderer) Clear() error {
	if r.inner.imageID > 0 {
		seq := fmt.Sprintf("\x1b_Ga=d,d=i,i=%d\x1b\\", r.inner.imageID)
		fmt.Print(r.wrapAllKittySequences(seq))
	}
	return nil
}

// UploadMinimap creates a downscaled thumbnail, deleting the old one via tmux passthrough.
func (r *TmuxRenderer) UploadMinimap(img *domain.ImageEntity, cols, rows int, cellAspect float64) error {
	if r.inner.minimapID > 0 {
		seq := fmt.Sprintf("\x1b_Ga=d,d=i,i=%d\x1b\\", r.inner.minimapID)
		fmt.Print(r.wrapAllKittySequences(seq))
	}
	r.inner.minimapID = atomic.AddUint32(&imageIDCounter, 1)
	r.inner.prepareMinimapBase(img, cols, rows, cellAspect)
	return nil
}

// DisplayMinimap returns the minimap sequence wrapped for tmux.
func (r *TmuxRenderer) DisplayMinimap(vp *domain.Viewport, cols, rows int, borderColor string) (string, error) {
	seq, err := r.inner.DisplayMinimap(vp, cols, rows, borderColor)
	if err != nil {
		return "", err
	}
	return r.wrapAllKittySequences(seq), nil
}

// ClearMinimap removes the minimap from the terminal via tmux passthrough.
func (r *TmuxRenderer) ClearMinimap() error {
	if r.inner.minimapID > 0 {
		seq := fmt.Sprintf("\x1b_Ga=d,d=i,i=%d\x1b\\", r.inner.minimapID)
		fmt.Print(r.wrapAllKittySequences(seq))
	}
	r.inner.prevCached = false
	return nil
}

// wrapAllKittySequences finds all Kitty APC sequences (\x1b_G...\x1b\\) in
// the input string and wraps each one in DCS passthrough. A CSI cursor
// positioning sequence (\x1b[...H) immediately preceding a Kitty APC is
// included inside the same passthrough — with row/col adjusted by the
// pane offset — so that the image renders inside the correct tmux pane.
func (r *TmuxRenderer) wrapAllKittySequences(s string) string {
	var out strings.Builder
	out.Grow(len(s) * 2)

	for i := 0; i < len(s); {
		// Look for Kitty APC start: \x1b_G
		if i+2 < len(s) && s[i] == 0x1b && s[i+1] == '_' && s[i+2] == 'G' {
			// Find the ST terminator: \x1b\\
			end := strings.Index(s[i+3:], "\x1b\\")
			if end >= 0 {
				seqEnd := i + 3 + end + 2 // include \x1b\\

				// Check if a CSI cursor position (\x1b[...H) was just written
				// to out and pull it into the passthrough with pane offset applied.
				prefix := r.extractTrailingCursorMove(&out)

				kittySeq := s[i:seqEnd]
				escaped := strings.ReplaceAll(prefix+kittySeq, "\x1b", "\x1b\x1b")
				out.WriteString("\x1bPtmux;")
				out.WriteString(escaped)
				out.WriteString("\x1b\\")
				i = seqEnd
				continue
			}
		}
		out.WriteByte(s[i])
		i++
	}

	return out.String()
}

// extractTrailingCursorMove checks if the builder ends with a CSI cursor
// position sequence (\x1b[<row>;<col>H or \x1b[H). If found, it removes
// that sequence from the builder and returns a new cursor move with
// row and col adjusted by the pane offset.
func (r *TmuxRenderer) extractTrailingCursorMove(b *strings.Builder) string {
	s := b.String()

	// Scan backwards for \x1b[...H pattern
	if len(s) < 2 {
		return ""
	}
	if s[len(s)-1] != 'H' {
		return ""
	}

	// Walk backwards from the 'H' to find ESC [
	j := len(s) - 2
	for j >= 0 && ((s[j] >= '0' && s[j] <= '9') || s[j] == ';') {
		j--
	}
	if j < 1 || s[j] != '[' || s[j-1] != 0x1b {
		return ""
	}

	csiStart := j - 1
	params := s[j+1 : len(s)-1] // between '[' and 'H'

	// Parse row;col (default 1;1 for bare \x1b[H)
	row, col := 1, 1
	if params != "" {
		parts := strings.SplitN(params, ";", 2)
		if v, err := strconv.Atoi(parts[0]); err == nil {
			row = v
		}
		if len(parts) == 2 {
			if v, err := strconv.Atoi(parts[1]); err == nil {
				col = v
			}
		}
	}

	// Rebuild the builder without the cursor sequence
	b.Reset()
	b.WriteString(s[:csiStart])

	// Return adjusted cursor move
	return fmt.Sprintf("\x1b[%d;%dH", row+r.paneTop, col+r.paneLeft)
}
