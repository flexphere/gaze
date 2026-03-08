package ffmpeg

import (
	"bytes"
	"os/exec"
	"strconv"
	"strings"
)

// Available reports whether both ffmpeg and ffprobe are found in PATH.
func Available() bool {
	_, errF := exec.LookPath("ffmpeg")
	_, errP := exec.LookPath("ffprobe")
	return errF == nil && errP == nil
}

// IsVideo reports whether the file at path is a video (not a still image).
// Returns false if ffprobe cannot inspect the file.
func IsVideo(path string) bool {
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=nb_frames,duration",
		"-of", "default=noprint_wrappers=1",
		path,
	) //nolint:gosec // path is user-provided CLI argument

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return false
	}

	fields := make(map[string]string)
	for line := range strings.SplitSeq(strings.TrimSpace(stdout.String()), "\n") {
		k, v, ok := strings.Cut(line, "=")
		if ok {
			fields[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}

	if nbStr, ok := fields["nb_frames"]; ok && nbStr != "N/A" {
		nb, err := strconv.Atoi(nbStr)
		if err == nil && nb > 1 {
			return true
		}
	}

	if durStr, ok := fields["duration"]; ok && durStr != "N/A" {
		dur, err := strconv.ParseFloat(durStr, 64)
		if err == nil && dur > 0 {
			return true
		}
	}

	return false
}
