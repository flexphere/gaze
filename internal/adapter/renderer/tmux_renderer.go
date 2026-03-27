package renderer

import (
	"fmt"
	"os"
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
// Returns an error if not running inside a tmux session.
func NewTmuxRenderer() (*TmuxRenderer, error) {
	if os.Getenv("TMUX") == "" {
		return nil, fmt.Errorf("tmux renderer requires a tmux session (TMUX environment variable not set)")
	}
	top, left, err := queryTmuxPaneOffset()
	if err != nil {
		return nil, fmt.Errorf("querying tmux pane offset: %w", err)
	}
	return &TmuxRenderer{inner: NewKittyRenderer(), paneTop: top, paneLeft: left}, nil
}

// RefreshPaneOffset re-queries the tmux pane position.
// On failure, the previous offset is retained to avoid disrupting rendering.
func (r *TmuxRenderer) RefreshPaneOffset() {
	top, left, err := queryTmuxPaneOffset()
	if err != nil {
		return // keep previous values
	}
	r.paneTop = top
	r.paneLeft = left
}

// queryTmuxPaneOffset returns the current pane's top-left corner offset
// within the outer terminal by querying tmux.
func queryTmuxPaneOffset() (top, left int, err error) {
	out, err := exec.Command("tmux", "display-message", "-p", "#{pane_top} #{pane_left}").Output()
	if err != nil {
		return 0, 0, fmt.Errorf("running tmux display-message: %w", err)
	}
	parts := strings.Fields(strings.TrimSpace(string(out)))
	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("unexpected tmux output: %q", string(out))
	}
	top, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("parsing pane_top %q: %w", parts[0], err)
	}
	left, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("parsing pane_left %q: %w", parts[1], err)
	}
	return top, left, nil
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
//
// Uses bulk copies from the input string to avoid re-copying the output
// buffer when extracting cursor moves.
func (r *TmuxRenderer) wrapAllKittySequences(s string) string {
	var out strings.Builder
	out.Grow(len(s))

	written := 0 // index in s up to which we've flushed to out

	for i := 0; i < len(s); {
		// Look for Kitty APC start: \x1b_G
		if i+2 < len(s) && s[i] == 0x1b && s[i+1] == '_' && s[i+2] == 'G' {
			// Find the ST terminator: \x1b\\
			end := strings.Index(s[i+3:], "\x1b\\")
			if end >= 0 {
				seqEnd := i + 3 + end + 2 // include \x1b\\

				// Scan backwards in the pending input for a cursor move
				prefix, flushEnd := r.findCursorMoveInPending(s[written:i])

				// Flush pending bytes up to (but not including) the cursor move
				out.WriteString(s[written : written+flushEnd])

				// Wrap cursor move + Kitty sequence with cursor save/restore
				// so that the outer terminal cursor is restored after drawing.
				kittySeq := s[i:seqEnd]
				content := "\x1b7" + prefix + kittySeq + "\x1b8" // DECSC ... DECRC
				escaped := strings.ReplaceAll(content, "\x1b", "\x1b\x1b")
				out.WriteString("\x1bPtmux;")
				out.WriteString(escaped)
				out.WriteString("\x1b\\")

				written = seqEnd
				i = seqEnd
				continue
			}
		}
		i++
	}

	// Flush remaining
	out.WriteString(s[written:])

	return out.String()
}

// findCursorMoveInPending scans the pending (not yet flushed) portion of the
// input for a trailing CSI cursor position sequence (\x1b[<row>;<col>H).
// If found, returns the adjusted cursor move string and the flush boundary
// (number of bytes to flush before the cursor move). If not found, returns
// ("", len(pending)) so the caller flushes everything.
func (r *TmuxRenderer) findCursorMoveInPending(pending string) (cursorMove string, flushEnd int) {
	if len(pending) < 2 || pending[len(pending)-1] != 'H' {
		return "", len(pending)
	}

	// Walk backwards from the 'H' to find ESC [
	j := len(pending) - 2
	for j >= 0 && ((pending[j] >= '0' && pending[j] <= '9') || pending[j] == ';') {
		j--
	}
	if j < 1 || pending[j] != '[' || pending[j-1] != 0x1b {
		return "", len(pending)
	}

	csiStart := j - 1
	params := pending[j+1 : len(pending)-1] // between '[' and 'H'

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

	return fmt.Sprintf("\x1b[%d;%dH", row+r.paneTop, col+r.paneLeft), csiStart
}
