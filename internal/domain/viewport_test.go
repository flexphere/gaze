package domain

import (
	"image"
	"math"
	"testing"
)

func defaultViewportConfig() ViewportConfig {
	return ViewportConfig{
		ZoomStep: 0.1,
		PanStep:  0.05,
		MinZoom:  0.1,
		MaxZoom:  20.0,
	}
}

func setupViewport(imgW, imgH, termW, termH int) *Viewport {
	vp := NewViewport(defaultViewportConfig())
	vp.TermWidth = termW
	vp.TermHeight = termH
	vp.ImgWidth = imgW
	vp.ImgHeight = imgH
	return vp
}

func TestNewViewport(t *testing.T) {
	cfg := defaultViewportConfig()
	vp := NewViewport(cfg)

	if vp.ZoomLevel != 1.0 {
		t.Errorf("ZoomLevel = %f, want 1.0", vp.ZoomLevel)
	}
	if vp.OffsetX != 0 {
		t.Errorf("OffsetX = %f, want 0", vp.OffsetX)
	}
	if vp.OffsetY != 0 {
		t.Errorf("OffsetY = %f, want 0", vp.OffsetY)
	}
}

func TestViewport_VisibleWidth(t *testing.T) {
	tests := []struct {
		name      string
		imgWidth  int
		zoomLevel float64
		want      float64
	}{
		{"zoom 1.0", 1000, 1.0, 1000},
		{"zoom 2.0", 1000, 2.0, 500},
		{"zoom 0.5", 1000, 0.5, 2000},
		{"zoom 10.0", 1000, 10.0, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vp := setupViewport(tt.imgWidth, 500, 80, 24)
			vp.ZoomLevel = tt.zoomLevel

			got := vp.VisibleWidth()
			if math.Abs(got-tt.want) > 0.001 {
				t.Errorf("VisibleWidth() = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestViewport_VisibleHeight(t *testing.T) {
	vp := setupViewport(1000, 800, 80, 24)
	vp.ZoomLevel = 2.0

	got := vp.VisibleHeight()
	want := 400.0
	if math.Abs(got-want) > 0.001 {
		t.Errorf("VisibleHeight() = %f, want %f", got, want)
	}
}

func TestViewport_VisibleRect(t *testing.T) {
	tests := []struct {
		name    string
		offsetX float64
		offsetY float64
		zoom    float64
		imgW    int
		imgH    int
		want    image.Rectangle
	}{
		{
			name: "full image at zoom 1.0",
			zoom: 1.0, imgW: 800, imgH: 600,
			want: image.Rect(0, 0, 800, 600),
		},
		{
			name: "zoomed in 2x at origin",
			zoom: 2.0, imgW: 800, imgH: 600,
			want: image.Rect(0, 0, 400, 300),
		},
		{
			name:    "zoomed in 2x with offset",
			offsetX: 100, offsetY: 50,
			zoom: 2.0, imgW: 800, imgH: 600,
			want: image.Rect(100, 50, 500, 350),
		},
		{
			name:    "clamped to image bounds",
			offsetX: 700, offsetY: 500,
			zoom: 2.0, imgW: 800, imgH: 600,
			want: image.Rect(700, 500, 800, 600),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vp := setupViewport(tt.imgW, tt.imgH, 80, 24)
			vp.ZoomLevel = tt.zoom
			vp.OffsetX = tt.offsetX
			vp.OffsetY = tt.offsetY

			got := vp.VisibleRect()
			if got != tt.want {
				t.Errorf("VisibleRect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestViewport_ZoomIn(t *testing.T) {
	vp := setupViewport(1000, 800, 80, 24)
	initialZoom := vp.ZoomLevel

	vp.ZoomIn()

	if vp.ZoomLevel <= initialZoom {
		t.Errorf("ZoomIn should increase ZoomLevel, got %f (was %f)", vp.ZoomLevel, initialZoom)
	}
}

func TestViewport_ZoomOut(t *testing.T) {
	vp := setupViewport(1000, 800, 80, 24)
	vp.ZoomLevel = 5.0

	vp.ZoomOut()

	if vp.ZoomLevel >= 5.0 {
		t.Errorf("ZoomOut should decrease ZoomLevel, got %f", vp.ZoomLevel)
	}
}

func TestViewport_Zoom_ClampsToMinMax(t *testing.T) {
	vp := setupViewport(1000, 800, 80, 24)

	// Zoom way in
	for range 100 {
		vp.ZoomIn()
	}
	if vp.ZoomLevel > vp.maxZoom {
		t.Errorf("ZoomLevel %f exceeds maxZoom %f", vp.ZoomLevel, vp.maxZoom)
	}

	// Zoom way out
	for range 200 {
		vp.ZoomOut()
	}
	if vp.ZoomLevel < vp.minZoom {
		t.Errorf("ZoomLevel %f below minZoom %f", vp.ZoomLevel, vp.minZoom)
	}
}

func TestViewport_Zoom_PreservesCenter(t *testing.T) {
	vp := setupViewport(1000, 800, 80, 24)
	vp.ZoomLevel = 2.0
	vp.OffsetX = 200
	vp.OffsetY = 150
	vp.Clamp()

	centerXBefore := vp.OffsetX + vp.VisibleWidth()/2
	centerYBefore := vp.OffsetY + vp.VisibleHeight()/2

	vp.Zoom(0.1)

	centerXAfter := vp.OffsetX + vp.VisibleWidth()/2
	centerYAfter := vp.OffsetY + vp.VisibleHeight()/2

	if math.Abs(centerXAfter-centerXBefore) > 1.0 {
		t.Errorf("Center X shifted: before=%f, after=%f", centerXBefore, centerXAfter)
	}
	if math.Abs(centerYAfter-centerYBefore) > 1.0 {
		t.Errorf("Center Y shifted: before=%f, after=%f", centerYBefore, centerYAfter)
	}
}

func TestViewport_ZoomAt_PreservesCursorPoint(t *testing.T) {
	vp := setupViewport(1000, 800, 80, 24)
	vp.ZoomLevel = 2.0
	vp.OffsetX = 100
	vp.OffsetY = 80
	vp.Clamp()

	termX := 40
	termY := 12
	srcXBefore := vp.OffsetX + float64(termX)/float64(vp.TermWidth)*vp.VisibleWidth()
	srcYBefore := vp.OffsetY + float64(termY)/float64(vp.TermHeight)*vp.VisibleHeight()

	vp.ZoomAt(0.1, termX, termY)

	srcXAfter := vp.OffsetX + float64(termX)/float64(vp.TermWidth)*vp.VisibleWidth()
	srcYAfter := vp.OffsetY + float64(termY)/float64(vp.TermHeight)*vp.VisibleHeight()

	if math.Abs(srcXAfter-srcXBefore) > 1.0 {
		t.Errorf("Cursor point X shifted: before=%f, after=%f", srcXBefore, srcXAfter)
	}
	if math.Abs(srcYAfter-srcYBefore) > 1.0 {
		t.Errorf("Cursor point Y shifted: before=%f, after=%f", srcYBefore, srcYAfter)
	}
}

func TestViewport_ZoomAt_ZeroTerminalSize(t *testing.T) {
	vp := setupViewport(1000, 800, 0, 0)
	vp.ZoomLevel = 1.0

	// Should not panic
	vp.ZoomAt(0.1, 0, 0)

	if vp.ZoomLevel != 1.0 {
		t.Errorf("ZoomAt with zero terminal size should be no-op, got ZoomLevel=%f", vp.ZoomLevel)
	}
}

func TestViewport_Pan(t *testing.T) {
	vp := setupViewport(1000, 800, 80, 24)
	vp.ZoomLevel = 2.0

	vp.Pan(50, 30)

	if vp.OffsetX != 50 {
		t.Errorf("OffsetX = %f, want 50", vp.OffsetX)
	}
	if vp.OffsetY != 30 {
		t.Errorf("OffsetY = %f, want 30", vp.OffsetY)
	}
}

func TestViewport_Pan_ClampsAtEdges(t *testing.T) {
	vp := setupViewport(1000, 800, 80, 24)
	vp.ZoomLevel = 2.0

	// Pan far beyond bounds
	vp.Pan(9999, 9999)

	maxX := float64(vp.ImgWidth) - vp.VisibleWidth()
	maxY := float64(vp.ImgHeight) - vp.VisibleHeight()
	if vp.OffsetX > maxX+0.001 {
		t.Errorf("OffsetX %f exceeds max %f", vp.OffsetX, maxX)
	}
	if vp.OffsetY > maxY+0.001 {
		t.Errorf("OffsetY %f exceeds max %f", vp.OffsetY, maxY)
	}

	// Pan far negative
	vp.Pan(-9999, -9999)

	if vp.OffsetX < -0.001 {
		t.Errorf("OffsetX %f is negative", vp.OffsetX)
	}
	if vp.OffsetY < -0.001 {
		t.Errorf("OffsetY %f is negative", vp.OffsetY)
	}
}

func TestViewport_PanByStep(t *testing.T) {
	vp := setupViewport(1000, 800, 80, 24)
	vp.ZoomLevel = 2.0 // visible: 500x400

	vp.PanByStep(1, 0) // right by 5% of 500 = 25

	wantX := 500.0 * 0.05
	if math.Abs(vp.OffsetX-wantX) > 0.001 {
		t.Errorf("OffsetX = %f, want %f", vp.OffsetX, wantX)
	}
}

func TestViewport_FitToWindow(t *testing.T) {
	vp := setupViewport(1000, 800, 80, 24)
	vp.ZoomLevel = 5.0
	vp.OffsetX = 300
	vp.OffsetY = 200

	vp.FitToWindow()

	if vp.ZoomLevel != 1.0 {
		t.Errorf("ZoomLevel = %f, want 1.0", vp.ZoomLevel)
	}
	if vp.OffsetX != 0 {
		t.Errorf("OffsetX = %f, want 0", vp.OffsetX)
	}
	if vp.OffsetY != 0 {
		t.Errorf("OffsetY = %f, want 0", vp.OffsetY)
	}
}

func TestViewport_SetImageSize(t *testing.T) {
	vp := setupViewport(100, 100, 80, 24)
	vp.ZoomLevel = 3.0
	vp.OffsetX = 50

	vp.SetImageSize(2000, 1500)

	if vp.ImgWidth != 2000 {
		t.Errorf("ImgWidth = %d, want 2000", vp.ImgWidth)
	}
	if vp.ImgHeight != 1500 {
		t.Errorf("ImgHeight = %d, want 1500", vp.ImgHeight)
	}
	if vp.ZoomLevel != 1.0 {
		t.Errorf("ZoomLevel = %f, want 1.0 (should reset)", vp.ZoomLevel)
	}
}

func TestViewport_SetTerminalSize(t *testing.T) {
	vp := setupViewport(1000, 800, 80, 24)
	vp.ZoomLevel = 2.0
	vp.OffsetX = 600 // beyond valid after shrink

	vp.SetTerminalSize(40, 12)

	if vp.TermWidth != 40 {
		t.Errorf("TermWidth = %d, want 40", vp.TermWidth)
	}
	if vp.TermHeight != 12 {
		t.Errorf("TermHeight = %d, want 12", vp.TermHeight)
	}
}

func TestViewport_Clamp_CentersWhenImageSmaller(t *testing.T) {
	vp := setupViewport(100, 100, 80, 24)
	vp.ZoomLevel = 0.5 // visible area is 200x200, larger than image 100x100

	vp.Clamp()

	// Should center: offset = (100 - 200) / 2 = -50
	wantX := (float64(100) - vp.VisibleWidth()) / 2
	wantY := (float64(100) - vp.VisibleHeight()) / 2
	if math.Abs(vp.OffsetX-wantX) > 0.001 {
		t.Errorf("OffsetX = %f, want %f (centered)", vp.OffsetX, wantX)
	}
	if math.Abs(vp.OffsetY-wantY) > 0.001 {
		t.Errorf("OffsetY = %f, want %f (centered)", vp.OffsetY, wantY)
	}
}

func TestViewport_ZoomPercentage(t *testing.T) {
	tests := []struct {
		zoom float64
		want int
	}{
		{1.0, 100},
		{2.0, 200},
		{0.5, 50},
		{1.55, 155},
	}

	for _, tt := range tests {
		vp := setupViewport(100, 100, 80, 24)
		vp.ZoomLevel = tt.zoom

		got := vp.ZoomPercentage()
		if got != tt.want {
			t.Errorf("ZoomPercentage() at zoom %f = %d, want %d", tt.zoom, got, tt.want)
		}
	}
}

func TestViewport_VisibleWidth_ZeroZoom(t *testing.T) {
	vp := setupViewport(1000, 800, 80, 24)
	vp.ZoomLevel = 0

	got := vp.VisibleWidth()
	if got != 1000 {
		t.Errorf("VisibleWidth() with zero zoom = %f, want 1000", got)
	}
}
