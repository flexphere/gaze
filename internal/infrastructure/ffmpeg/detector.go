package ffmpeg

import (
	"os/exec"
)

// Available reports whether both ffmpeg and ffprobe are found in PATH.
func Available() bool {
	_, errF := exec.LookPath("ffmpeg")
	_, errP := exec.LookPath("ffprobe")
	return errF == nil && errP == nil
}
