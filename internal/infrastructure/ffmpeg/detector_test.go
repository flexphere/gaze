package ffmpeg

import (
	"os/exec"
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
