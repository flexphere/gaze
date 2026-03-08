package ffmpeg

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func createTestVideo(t *testing.T, dir, name string, w, h int) string {
	t.Helper()
	path := filepath.Join(dir, name)
	cmd := exec.Command("ffmpeg",
		"-f", "lavfi",
		"-i", fmt.Sprintf("color=c=blue:s=%dx%d:r=10:d=1", w, h),
		"-pix_fmt", "yuv420p",
		"-y",
		path,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("creating test video: %v: %s", err, out)
	}
	return path
}

func TestVideoDecoder_Open(t *testing.T) {
	skipIfFFmpegUnavailable(t)

	dir := t.TempDir()
	path := createTestVideo(t, dir, "test.mp4", 80, 60)

	decoder := NewVideoDecoder()
	info, err := decoder.Open(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer decoder.Close()

	if info.Width != 80 {
		t.Errorf("Width = %d, want 80", info.Width)
	}
	if info.Height != 60 {
		t.Errorf("Height = %d, want 60", info.Height)
	}
	if math.Abs(info.FrameRate-10.0) > 1.0 {
		t.Errorf("FrameRate = %f, want ~10", info.FrameRate)
	}
	if info.Duration < 900*time.Millisecond || info.Duration > 1100*time.Millisecond {
		t.Errorf("Duration = %v, want ~1s", info.Duration)
	}
}

func TestVideoDecoder_NextFrame(t *testing.T) {
	skipIfFFmpegUnavailable(t)

	dir := t.TempDir()
	path := createTestVideo(t, dir, "test.mp4", 80, 60)

	decoder := NewVideoDecoder()
	info, err := decoder.Open(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer decoder.Close()

	img, err := decoder.NextFrame()
	if err != nil {
		t.Fatalf("unexpected error reading first frame: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != info.Width || bounds.Dy() != info.Height {
		t.Errorf("frame size = %dx%d, want %dx%d", bounds.Dx(), bounds.Dy(), info.Width, info.Height)
	}
}

func TestVideoDecoder_NextFrame_EOF(t *testing.T) {
	skipIfFFmpegUnavailable(t)

	dir := t.TempDir()
	path := createTestVideo(t, dir, "test.mp4", 80, 60)

	decoder := NewVideoDecoder()
	if _, err := decoder.Open(path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer decoder.Close()

	frameCount := 0
	for {
		_, err := decoder.NextFrame()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("unexpected error at frame %d: %v", frameCount, err)
		}
		frameCount++
		if frameCount > 100 {
			t.Fatal("too many frames, expected ~10")
		}
	}

	if frameCount < 8 || frameCount > 12 {
		t.Errorf("frame count = %d, want ~10", frameCount)
	}
}

func TestVideoDecoder_Seek(t *testing.T) {
	skipIfFFmpegUnavailable(t)

	dir := t.TempDir()
	path := createTestVideo(t, dir, "test.mp4", 80, 60)

	decoder := NewVideoDecoder()
	if _, err := decoder.Open(path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer decoder.Close()

	if err := decoder.Seek(500 * time.Millisecond); err != nil {
		t.Fatalf("unexpected error seeking: %v", err)
	}

	img, err := decoder.NextFrame()
	if err != nil {
		t.Fatalf("unexpected error reading frame after seek: %v", err)
	}
	bounds := img.Bounds()
	if bounds.Dx() != 80 || bounds.Dy() != 60 {
		t.Errorf("frame size after seek = %dx%d, want 80x60", bounds.Dx(), bounds.Dy())
	}
}

func TestVideoDecoder_Open_FileNotFound(t *testing.T) {
	skipIfFFmpegUnavailable(t)

	decoder := NewVideoDecoder()
	_, err := decoder.Open("/nonexistent/video.mp4")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestVideoDecoder_NextFrame_BeforeOpen(t *testing.T) {
	decoder := NewVideoDecoder()
	_, err := decoder.NextFrame()
	if err == nil {
		t.Fatal("expected error calling NextFrame before Open")
	}
}

func TestVideoDecoder_Seek_BeforeOpen(t *testing.T) {
	decoder := NewVideoDecoder()
	err := decoder.Seek(0)
	if err == nil {
		t.Fatal("expected error calling Seek before Open")
	}
}

func TestProbeVideo(t *testing.T) {
	skipIfFFmpegUnavailable(t)

	dir := t.TempDir()
	path := createTestVideo(t, dir, "test.mp4", 80, 60)

	info, err := probeVideo(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Width != 80 {
		t.Errorf("Width = %d, want 80", info.Width)
	}
	if info.Height != 60 {
		t.Errorf("Height = %d, want 60", info.Height)
	}
}

func TestParseFrameRate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  float64
	}{
		{"30fps", "30/1", 30.0},
		{"24000/1001", "24000/1001", 23.976},
		{"zero denominator", "0/0", 0},
		{"invalid", "abc", 0},
		{"empty", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseFrameRate(tt.input)
			if tt.want == 0 {
				if got != 0 {
					t.Errorf("parseFrameRate(%q) = %f, want 0", tt.input, got)
				}
			} else if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("parseFrameRate(%q) = %f, want %f", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  time.Duration
	}{
		{"10 seconds", "10.000000", 10 * time.Second},
		{"N/A", "N/A", 0},
		{"empty", "", 0},
		{"1.5 seconds", "1.500000", 1500 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDuration(tt.input)
			if got != tt.want {
				t.Errorf("parseDuration(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
