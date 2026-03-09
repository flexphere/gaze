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
	vp.CellAspectRatio = 2.0 // default: cells are twice as tall as wide
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

	// VisibleWidth = 1000/2 = 500
	// VisibleHeight = 500 * 24 * 2.0 / 80 = 300
	got := vp.VisibleHeight()
	want := 300.0
	if math.Abs(got-want) > 0.001 {
		t.Errorf("VisibleHeight() = %f, want %f", got, want)
	}
}

func TestViewport_VisibleHeight_WithCellAspectRatio(t *testing.T) {
	tests := []struct {
		name            string
		imgW, imgH      int
		termW, termH    int
		zoom            float64
		cellAspectRatio float64
		want            float64
	}{
		{
			name: "default 2:1 cells",
			imgW: 1000, imgH: 800, termW: 80, termH: 24,
			zoom: 1.0, cellAspectRatio: 2.0,
			// VW=1000, VH=1000*24*2/80=600
			want: 600,
		},
		{
			name: "square cells",
			imgW: 1000, imgH: 800, termW: 80, termH: 24,
			zoom: 1.0, cellAspectRatio: 1.0,
			// VW=1000, VH=1000*24*1/80=300
			want: 300,
		},
		{
			name: "zoomed with non-default aspect",
			imgW: 800, imgH: 600, termW: 100, termH: 50,
			zoom: 2.0, cellAspectRatio: 1.8,
			// VW=400, VH=400*50*1.8/100=360
			want: 360,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vp := setupViewport(tt.imgW, tt.imgH, tt.termW, tt.termH)
			vp.ZoomLevel = tt.zoom
			vp.CellAspectRatio = tt.cellAspectRatio

			got := vp.VisibleHeight()
			if math.Abs(got-tt.want) > 0.001 {
				t.Errorf("VisibleHeight() = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestViewport_VisibleRect(t *testing.T) {
	// With termW=80, termH=24, cellAspect=2.0:
	// At zoom 1.0: VW=imgW, VH=imgW*24*2/80 = imgW*0.6
	// At zoom 2.0: VW=imgW/2, VH=(imgW/2)*0.6

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
			// VW=800, VH=800*24*2/80=480 < 600, so height clamped to image
			want: image.Rect(0, 0, 800, 480),
		},
		{
			name: "zoomed in 2x at origin",
			zoom: 2.0, imgW: 800, imgH: 600,
			// VW=400, VH=400*24*2/80=240
			want: image.Rect(0, 0, 400, 240),
		},
		{
			name:    "zoomed in 2x with offset",
			offsetX: 100, offsetY: 50,
			zoom: 2.0, imgW: 800, imgH: 600,
			// VW=400, VH=240
			want: image.Rect(100, 50, 500, 290),
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
	vp.fitZoom = 0.5
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
	vp.fitZoom = 0.5
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
	vp.fitZoom = 0.5 // set fitZoom so Pan works at zoom=2.0

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
	vp.ZoomLevel = 2.0 // visible: 500x300 (with cell aspect 2.0)
	vp.fitZoom = 0.5   // allow zoom=2.0 to be zoomed in

	vp.PanByStep(1, 0) // right by 5% of 500 = 25

	wantX := 500.0 * 0.05
	if math.Abs(vp.OffsetX-wantX) > 0.001 {
		t.Errorf("OffsetX = %f, want %f", vp.OffsetX, wantX)
	}
}

func TestViewport_FitToWindow(t *testing.T) {
	tests := []struct {
		name            string
		imgW, imgH      int
		termW, termH    int
		cellAspectRatio float64
		wantZoom        float64
	}{
		{
			name: "wide image fits by width",
			imgW: 1000, imgH: 200, termW: 80, termH: 24,
			cellAspectRatio: 2.0,
			// zoomH = 1000*24*2/(80*200) = 3.0 > 1.0 → zoom = 1.0
			wantZoom: 1.0,
		},
		{
			name: "tall image fits by height",
			imgW: 800, imgH: 600, termW: 80, termH: 24,
			cellAspectRatio: 2.0,
			// zoomH = 800*24*2/(80*600) = 0.8 < 1.0 → zoom = 0.8
			wantZoom: 0.8,
		},
		{
			name: "square image with 2:1 cells",
			imgW: 1000, imgH: 1000, termW: 80, termH: 24,
			cellAspectRatio: 2.0,
			// zoomH = 1000*24*2/(80*1000) = 0.6 → zoom = 0.6
			wantZoom: 0.6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vp := setupViewport(tt.imgW, tt.imgH, tt.termW, tt.termH)
			vp.CellAspectRatio = tt.cellAspectRatio
			vp.ZoomLevel = 5.0
			vp.OffsetX = 300
			vp.OffsetY = 200

			vp.FitToWindow()

			if math.Abs(vp.ZoomLevel-tt.wantZoom) > 0.001 {
				t.Errorf("ZoomLevel = %f, want %f", vp.ZoomLevel, tt.wantZoom)
			}
			if vp.OffsetX > 0.001 || vp.OffsetX < -float64(tt.imgW) {
				t.Errorf("OffsetX = %f, should be near 0 or centered", vp.OffsetX)
			}
		})
	}
}

func TestViewport_FitToWindow_ZeroTerminalSize(t *testing.T) {
	vp := setupViewport(1000, 800, 0, 0)
	vp.FitToWindow()

	if vp.ZoomLevel != 1.0 {
		t.Errorf("ZoomLevel = %f, want 1.0 for zero terminal", vp.ZoomLevel)
	}
}

func TestViewport_FitToWindow_ZeroTerminalSize_UpdatesFitZoom(t *testing.T) {
	vp := setupViewport(1000, 800, 80, 24)
	vp.FitToWindow()
	// fitZoom is now < 1.0
	prevFitZoom := vp.fitZoom

	// Set terminal to zero (early return path)
	vp.TermWidth = 0
	vp.TermHeight = 0
	vp.FitToWindow()

	if vp.ZoomLevel != 1.0 {
		t.Errorf("ZoomLevel = %f, want 1.0", vp.ZoomLevel)
	}
	if vp.fitZoom != vp.ZoomLevel {
		t.Errorf("fitZoom = %f, want %f (should match ZoomLevel); was %f before early return",
			vp.fitZoom, vp.ZoomLevel, prevFitZoom)
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
	// FitToWindow: zoomH = 2000*24*2/(80*1500) = 0.8 → zoom = 0.8
	wantZoom := 0.8
	if math.Abs(vp.ZoomLevel-wantZoom) > 0.001 {
		t.Errorf("ZoomLevel = %f, want %f (should fit)", vp.ZoomLevel, wantZoom)
	}
}

func TestViewport_SetTerminalSize(t *testing.T) {
	vp := setupViewport(1000, 800, 80, 24)
	vp.ZoomLevel = 2.0
	vp.fitZoom = 0.5 // pretend we're zoomed in
	vp.OffsetX = 400

	vp.SetTerminalSize(40, 12)

	if vp.TermWidth != 40 {
		t.Errorf("TermWidth = %d, want 40", vp.TermWidth)
	}
	if vp.TermHeight != 12 {
		t.Errorf("TermHeight = %d, want 12", vp.TermHeight)
	}
	// ZoomLevel should remain 2.0 since we were zoomed in (not at fit)
	if math.Abs(vp.ZoomLevel-2.0) > 0.001 {
		t.Errorf("ZoomLevel = %f, want 2.0 (zoomed in, should keep)", vp.ZoomLevel)
	}
}

func TestViewport_SetTerminalSize_NoRefitWhenZoomedOut(t *testing.T) {
	vp := setupViewport(1000, 800, 80, 24)
	vp.FitToWindow() // fitZoom = 0.75
	// Zoom out below fit level
	vp.ZoomLevel = 0.3

	vp.SetTerminalSize(100, 50)

	// Should NOT re-fit because we were zoomed out (not at fit level)
	if math.Abs(vp.ZoomLevel-0.3) > 0.001 {
		t.Errorf("ZoomLevel = %f, want 0.3 (zoomed out, should not re-fit)", vp.ZoomLevel)
	}
}

func TestViewport_SetTerminalSize_RefitsWhenAtFit(t *testing.T) {
	vp := setupViewport(1000, 800, 80, 24)
	vp.FitToWindow() // zoomH = 1000*24*2/(80*800) = 0.75
	fitBefore := vp.ZoomLevel

	// Change to different aspect ratio terminal
	vp.SetTerminalSize(100, 50)
	// zoomH = 1000*50*2/(100*800) = 1.25 → zoom = min(1.0, 1.25) = 1.0

	if math.Abs(vp.ZoomLevel-fitBefore) < 0.001 {
		t.Errorf("ZoomLevel should be recalculated: before=%f, after=%f", fitBefore, vp.ZoomLevel)
	}
	if vp.TermWidth != 100 || vp.TermHeight != 50 {
		t.Errorf("Terminal size not updated")
	}
}

func TestViewport_Clamp_CentersWhenImageSmaller(t *testing.T) {
	vp := setupViewport(100, 100, 80, 24)
	vp.ZoomLevel = 0.5 // VW=200, VH=200*24*2/80=120

	vp.Clamp()

	// Should center: offset = (imgDim - visibleDim) / 2
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
		name    string
		zoom    float64
		fitZoom float64
		want    int
	}{
		{"at fit level", 0.6, 0.6, 100},
		{"2x fit level", 1.2, 0.6, 200},
		{"half of fit", 0.3, 0.6, 50},
		{"unset fitZoom defaults to 1.0", 1.55, 0, 155},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vp := setupViewport(100, 100, 80, 24)
			vp.ZoomLevel = tt.zoom
			vp.fitZoom = tt.fitZoom

			got := vp.ZoomPercentage()
			if got != tt.want {
				t.Errorf("ZoomPercentage() at zoom %f (fit %f) = %d, want %d", tt.zoom, tt.fitZoom, got, tt.want)
			}
		})
	}
}

func TestViewport_IsZoomed(t *testing.T) {
	tests := []struct {
		name      string
		zoomLevel float64
		fitZoom   float64
		want      bool
	}{
		{"at fit level", 0.6, 0.6, false},
		{"just above fit within tolerance", 0.6003, 0.6, false},
		{"zoomed in", 0.9, 0.6, true},
		{"zoomed in 2x from fit", 1.2, 0.6, true},
		{"below fit", 0.3, 0.6, false},
		{"slightly above tolerance", 0.6012, 0.6, true},
		{"unset fitZoom defaults to 1.0", 1.5, 0, true},
		{"unset fitZoom at 1.0", 1.0, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vp := setupViewport(1000, 800, 80, 24)
			vp.ZoomLevel = tt.zoomLevel
			vp.fitZoom = tt.fitZoom

			got := vp.IsZoomed()
			if got != tt.want {
				t.Errorf("IsZoomed() at zoom %f (fit %f) = %v, want %v", tt.zoomLevel, tt.fitZoom, got, tt.want)
			}
		})
	}
}

func TestViewport_SetCellAspectRatio(t *testing.T) {
	tests := []struct {
		name  string
		input float64
		want  float64
	}{
		{"normal value", 2.0, 2.0},
		{"clamped to min", -1.0, 0.1},
		{"clamped to max", 100.0, 10.0},
		{"zero clamped to min", 0.0, 0.1},
		{"edge min", 0.1, 0.1},
		{"edge max", 10.0, 10.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vp := setupViewport(100, 100, 80, 24)
			vp.SetCellAspectRatio(tt.input)
			if math.Abs(vp.CellAspectRatio-tt.want) > 0.001 {
				t.Errorf("CellAspectRatio = %f, want %f", vp.CellAspectRatio, tt.want)
			}
		})
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
