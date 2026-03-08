package ffmpeg

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/flexphere/gaze/internal/domain"
)

// VideoDecoder decodes video frames using ffmpeg CLI.
type VideoDecoder struct {
	info   *domain.VideoInfo
	path   string
	cmd    *exec.Cmd
	stdout io.ReadCloser
}

// NewVideoDecoder creates a new VideoDecoder.
func NewVideoDecoder() *VideoDecoder {
	return &VideoDecoder{}
}

// Open probes the video file and starts the ffmpeg decode process.
func (d *VideoDecoder) Open(path string) (*domain.VideoInfo, error) {
	info, err := probeVideo(path)
	if err != nil {
		return nil, fmt.Errorf("probing video: %w", err)
	}

	d.info = info
	d.path = path

	if err := d.startFFmpeg(0); err != nil {
		return nil, fmt.Errorf("starting decoder: %w", err)
	}

	return info, nil
}

// NextFrame reads the next video frame as an RGBA image.
// Returns io.EOF when the video ends.
func (d *VideoDecoder) NextFrame() (image.Image, error) {
	frameSize := d.info.Width * d.info.Height * 4
	pix := make([]byte, frameSize)

	if _, err := io.ReadFull(d.stdout, pix); err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, io.EOF
		}
		return nil, fmt.Errorf("reading frame: %w", err)
	}

	return &image.RGBA{
		Pix:    pix,
		Stride: d.info.Width * 4,
		Rect:   image.Rect(0, 0, d.info.Width, d.info.Height),
	}, nil
}

// Seek restarts the decoder at the given position.
func (d *VideoDecoder) Seek(pos time.Duration) error {
	d.stopFFmpeg()
	if err := d.startFFmpeg(pos); err != nil {
		return fmt.Errorf("seeking to %v: %w", pos, err)
	}
	return nil
}

// Close stops the ffmpeg process and releases resources.
func (d *VideoDecoder) Close() error {
	d.stopFFmpeg()
	return nil
}

func (d *VideoDecoder) startFFmpeg(seekPos time.Duration) error {
	args := []string{"-nostdin", "-loglevel", "error"}

	if seekPos > 0 {
		args = append(args, "-ss", formatDuration(seekPos))
	}

	args = append(args,
		"-i", d.path,
		"-f", "rawvideo",
		"-pix_fmt", "rgba",
		"-an",
		"pipe:1",
	)

	cmd := exec.Command("ffmpeg", args...) //nolint:gosec // args constructed internally
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("creating stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting ffmpeg: %w", err)
	}

	d.cmd = cmd
	d.stdout = stdout
	return nil
}

func (d *VideoDecoder) stopFFmpeg() {
	if d.cmd == nil || d.cmd.Process == nil {
		return
	}
	// Kill and Wait errors are expected during cleanup
	_ = d.cmd.Process.Kill() //nolint:errcheck // process may have already exited
	_ = d.cmd.Wait()         //nolint:errcheck // error expected after kill
	d.cmd = nil
	d.stdout = nil
}

func probeVideo(path string) (*domain.VideoInfo, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height,r_frame_rate,duration",
		"-of", "default=noprint_wrappers=1",
		path,
	) //nolint:gosec // path is user-provided CLI argument

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("running ffprobe: %w: %s", err, stderr.String())
	}

	output := strings.TrimSpace(stdout.String())
	fields := make(map[string]string)
	for line := range strings.SplitSeq(output, "\n") {
		k, v, ok := strings.Cut(line, "=")
		if ok {
			fields[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}

	wStr, ok := fields["width"]
	if !ok {
		return nil, fmt.Errorf("ffprobe output missing width")
	}
	w, err := strconv.Atoi(wStr)
	if err != nil {
		return nil, fmt.Errorf("parsing width: %w", err)
	}

	hStr, ok := fields["height"]
	if !ok {
		return nil, fmt.Errorf("ffprobe output missing height")
	}
	h, err := strconv.Atoi(hStr)
	if err != nil {
		return nil, fmt.Errorf("parsing height: %w", err)
	}

	fps := parseFrameRate(fields["r_frame_rate"])
	dur := parseDuration(fields["duration"])

	return &domain.VideoInfo{
		Width:     w,
		Height:    h,
		FrameRate: fps,
		Duration:  dur,
	}, nil
}

func parseFrameRate(s string) float64 {
	num, den, ok := strings.Cut(s, "/")
	if !ok {
		return 0
	}
	n, err := strconv.ParseFloat(num, 64)
	if err != nil {
		return 0
	}
	d, err := strconv.ParseFloat(den, 64)
	if err != nil || d == 0 {
		return 0
	}
	return n / d
}

func parseDuration(s string) time.Duration {
	if s == "" || s == "N/A" {
		return 0
	}
	secs, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return time.Duration(secs * float64(time.Second))
}

func formatDuration(d time.Duration) string {
	return fmt.Sprintf("%.3f", d.Seconds())
}
