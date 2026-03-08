package ffmpeg

import (
	"os/exec"
	"path/filepath"
	"testing"
)

func TestAvailable(t *testing.T) {
	_, errF := exec.LookPath("ffmpeg")
	_, errP := exec.LookPath("ffprobe")
	bothInstalled := errF == nil && errP == nil

	got := Available()
	if got != bothInstalled {
		t.Errorf("Available() = %v, want %v", got, bothInstalled)
	}
}

func TestIsVideo_Image(t *testing.T) {
	skipIfFFmpegUnavailable(t)

	path := filepath.Join("..", "..", "..", "testdata", "test_100x100.png")
	if IsVideo(path) {
		t.Error("IsVideo returned true for a PNG image")
	}
}

func TestIsVideo_Video(t *testing.T) {
	skipIfFFmpegUnavailable(t)

	dir := t.TempDir()
	path := createTestVideo(t, dir, "test.mp4", 80, 60)
	if !IsVideo(path) {
		t.Error("IsVideo returned false for a video file")
	}
}

func TestIsVideo_NonexistentFile(t *testing.T) {
	skipIfFFmpegUnavailable(t)

	if IsVideo("/nonexistent/file.mp4") {
		t.Error("IsVideo returned true for nonexistent file")
	}
}
