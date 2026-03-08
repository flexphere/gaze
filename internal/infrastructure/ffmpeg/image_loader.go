package ffmpeg

import (
	"bytes"
	"fmt"
	"image"
	"os/exec"
	"strconv"
	"strings"

	"github.com/flexphere/gaze/internal/domain"
	"github.com/flexphere/gaze/internal/usecase"
)

// ImageLoader decodes images using ffmpeg CLI, falling back to a standard Go decoder.
type ImageLoader struct {
	fallback usecase.ImageLoaderPort
}

// NewImageLoader creates an ImageLoader that tries the Go standard decoder first,
// then falls back to ffmpeg for unsupported formats.
func NewImageLoader(fallback usecase.ImageLoaderPort) *ImageLoader {
	return &ImageLoader{fallback: fallback}
}

// Load decodes an image. It first attempts the Go standard decoder;
// if that fails, it shells out to ffmpeg.
func (l *ImageLoader) Load(path string) (*domain.ImageEntity, error) {
	img, err := l.fallback.Load(path)
	if err == nil {
		return img, nil
	}

	return l.loadWithFFmpeg(path)
}

// probe returns the width, height, and codec name of the first video stream.
func probe(path string) (width, height int, codec string, err error) {
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height,codec_name",
		"-of", "default=noprint_wrappers=1",
		path,
	) //nolint:gosec // path is user-provided CLI argument

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return 0, 0, "", fmt.Errorf("running ffprobe: %w: %s", err, stderr.String())
	}

	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return 0, 0, "", fmt.Errorf("ffprobe returned no output for %q", path)
	}

	fields := make(map[string]string)
	for line := range strings.SplitSeq(output, "\n") {
		k, v, ok := strings.Cut(line, "=")
		if ok {
			fields[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}

	wStr, ok := fields["width"]
	if !ok {
		return 0, 0, "", fmt.Errorf("ffprobe output missing width")
	}
	w, err := strconv.Atoi(wStr)
	if err != nil {
		return 0, 0, "", fmt.Errorf("parsing width: %w", err)
	}

	hStr, ok := fields["height"]
	if !ok {
		return 0, 0, "", fmt.Errorf("ffprobe output missing height")
	}
	h, err := strconv.Atoi(hStr)
	if err != nil {
		return 0, 0, "", fmt.Errorf("parsing height: %w", err)
	}

	codec = fields["codec_name"]

	return w, h, codec, nil
}

func (l *ImageLoader) loadWithFFmpeg(path string) (*domain.ImageEntity, error) {
	w, h, codec, err := probe(path)
	if err != nil {
		return nil, fmt.Errorf("probing image: %w", err)
	}

	cmd := exec.Command(
		"ffmpeg",
		"-i", path,
		"-frames:v", "1",
		"-f", "rawvideo",
		"-pix_fmt", "rgba",
		"pipe:1",
	) //nolint:gosec // path is user-provided CLI argument

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("decoding with ffmpeg: %w: %s", err, stderr.String())
	}

	expected := w * h * 4
	if stdout.Len() != expected {
		return nil, fmt.Errorf("ffmpeg output size mismatch: got %d bytes, want %d", stdout.Len(), expected)
	}

	img := &image.RGBA{
		Pix:    stdout.Bytes(),
		Stride: w * 4,
		Rect:   image.Rect(0, 0, w, h),
	}

	return domain.NewImageEntity(img, path, codec), nil
}
