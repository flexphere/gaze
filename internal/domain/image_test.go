package domain

import (
	"image"
	"testing"
	"time"
)

func TestNewImageEntity(t *testing.T) {
	tests := []struct {
		name       string
		width      int
		height     int
		path       string
		format     string
		wantWidth  int
		wantHeight int
	}{
		{
			name:       "standard image",
			width:      800,
			height:     600,
			path:       "/path/to/image.png",
			format:     "png",
			wantWidth:  800,
			wantHeight: 600,
		},
		{
			name:       "square image",
			width:      100,
			height:     100,
			path:       "test.jpg",
			format:     "jpeg",
			wantWidth:  100,
			wantHeight: 100,
		},
		{
			name:       "tall image",
			width:      100,
			height:     1000,
			path:       "tall.gif",
			format:     "gif",
			wantWidth:  100,
			wantHeight: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := image.NewRGBA(image.Rect(0, 0, tt.width, tt.height))
			entity := NewImageEntity(src, tt.path, tt.format)

			if entity.Width != tt.wantWidth {
				t.Errorf("Width = %d, want %d", entity.Width, tt.wantWidth)
			}
			if entity.Height != tt.wantHeight {
				t.Errorf("Height = %d, want %d", entity.Height, tt.wantHeight)
			}
			if entity.Path != tt.path {
				t.Errorf("Path = %q, want %q", entity.Path, tt.path)
			}
			if entity.Format != tt.format {
				t.Errorf("Format = %q, want %q", entity.Format, tt.format)
			}
			if entity.Source == nil {
				t.Error("Source is nil")
			}
			if entity.IsAnimated() {
				t.Error("IsAnimated() = true, want false for static image")
			}
		})
	}
}

func TestNewAnimatedImageEntity(t *testing.T) {
	newFrame := func(delay time.Duration) Frame {
		return Frame{Image: image.NewRGBA(image.Rect(0, 0, 40, 30)), Delay: delay}
	}

	tests := []struct {
		name         string
		frames       []Frame
		wantAnimated bool
	}{
		{
			name:         "single frame",
			frames:       []Frame{newFrame(100 * time.Millisecond)},
			wantAnimated: false,
		},
		{
			name:         "multiple frames",
			frames:       []Frame{newFrame(100 * time.Millisecond), newFrame(200 * time.Millisecond)},
			wantAnimated: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := NewAnimatedImageEntity(tt.frames, "anim.gif", "gif")

			if entity.Source != tt.frames[0].Image {
				t.Error("Source should be the first frame")
			}
			if entity.Width != 40 || entity.Height != 30 {
				t.Errorf("size = %dx%d, want 40x30", entity.Width, entity.Height)
			}
			if len(entity.Frames) != len(tt.frames) {
				t.Errorf("len(Frames) = %d, want %d", len(entity.Frames), len(tt.frames))
			}
			if got := entity.IsAnimated(); got != tt.wantAnimated {
				t.Errorf("IsAnimated() = %v, want %v", got, tt.wantAnimated)
			}
		})
	}
}
