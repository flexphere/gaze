package domain

import "time"

// MediaType represents the type of media file.
type MediaType int

const (
	// MediaTypeImage represents a still image file.
	MediaTypeImage MediaType = iota
	// MediaTypeVideo represents a video file.
	MediaTypeVideo
)

// VideoInfo holds metadata about a video file.
type VideoInfo struct {
	Width     int
	Height    int
	FrameRate float64
	Duration  time.Duration
}
