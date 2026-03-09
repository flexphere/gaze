package usecase

import (
	"errors"
	"image"
	"testing"

	"github.com/flexphere/gaze/internal/domain"
)

type mockRenderer struct {
	uploadErr  error
	displayOut string
	displayErr error
	uploadCnt  int
	displayCnt int

	minimapUploadCnt  int
	minimapUploadErr  error
	minimapDisplayOut string
	minimapDisplayErr error
	minimapDisplayCnt int
	minimapClearCnt   int
	minimapClearErr   error
}

func (m *mockRenderer) Upload(_ *domain.ImageEntity) error {
	m.uploadCnt++
	return m.uploadErr
}

func (m *mockRenderer) Display(_ *domain.Viewport) (string, error) {
	m.displayCnt++
	return m.displayOut, m.displayErr
}

func (m *mockRenderer) Clear() error {
	return nil
}

func (m *mockRenderer) UploadMinimap(_ *domain.ImageEntity, _, _ int) error {
	m.minimapUploadCnt++
	return m.minimapUploadErr
}

func (m *mockRenderer) DisplayMinimap(_ *domain.Viewport, _, _ int, _ string) (string, error) {
	m.minimapDisplayCnt++
	return m.minimapDisplayOut, m.minimapDisplayErr
}

func (m *mockRenderer) ClearMinimap() error {
	m.minimapClearCnt++
	return m.minimapClearErr
}

func (m *mockRenderer) SetVideoMode(_ bool) {}

func TestRenderFrameUseCase_SetMinimapEnabled(t *testing.T) {
	renderer := &mockRenderer{displayOut: "main"}
	cfg := domain.MinimapConfig{Enabled: true, Size: 0.2}
	uc := NewRenderFrameUseCase(renderer, cfg)

	if !uc.MinimapEnabled() {
		t.Error("minimap should be enabled initially")
	}

	uc.SetMinimapEnabled(false)
	if uc.MinimapEnabled() {
		t.Error("minimap should be disabled after SetMinimapEnabled(false)")
	}

	uc.SetMinimapEnabled(true)
	if !uc.MinimapEnabled() {
		t.Error("minimap should be re-enabled after SetMinimapEnabled(true)")
	}
}

func TestRenderFrameUseCase_Execute_MinimapToggleOff(t *testing.T) {
	renderer := &mockRenderer{
		displayOut:        "main",
		minimapDisplayOut: "mm",
	}
	cfg := domain.MinimapConfig{Enabled: true, Size: 0.2}
	uc := NewRenderFrameUseCase(renderer, cfg)

	img := domain.NewImageEntity(image.NewRGBA(image.Rect(0, 0, 800, 600)), "test.png", "png")
	vp := domain.NewViewport(domain.ViewportConfig{
		ZoomStep: 0.1, PanStep: 0.05, MinZoom: 0.1, MaxZoom: 20.0,
	})
	vp.ImgWidth = 800
	vp.ImgHeight = 600
	vp.TermWidth = 80
	vp.TermHeight = 24
	vp.ZoomLevel = 2.0

	// First frame — minimap shown
	got, err := uc.Execute(img, vp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "mainmm" {
		t.Errorf("output = %q, want %q", got, "mainmm")
	}

	// Toggle off
	uc.SetMinimapEnabled(false)

	// Next frame — minimap hidden, clear called
	got, err = uc.Execute(img, vp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "main" {
		t.Errorf("output = %q, want %q (minimap toggled off)", got, "main")
	}
	if renderer.minimapClearCnt != 1 {
		t.Errorf("minimap clear should be called once, got %d", renderer.minimapClearCnt)
	}
}

func TestRenderFrameUseCase_Execute_Success(t *testing.T) {
	renderer := &mockRenderer{displayOut: "\x1b[image data]"}
	uc := NewRenderFrameUseCase(renderer, domain.MinimapConfig{})

	img := domain.NewImageEntity(image.NewRGBA(image.Rect(0, 0, 100, 100)), "test.png", "png")
	vp := domain.NewViewport(domain.ViewportConfig{
		ZoomStep: 0.1, PanStep: 0.05, MinZoom: 0.1, MaxZoom: 20.0,
	})
	vp.ImgWidth = 100
	vp.ImgHeight = 100
	vp.TermWidth = 80
	vp.TermHeight = 24

	got, err := uc.Execute(img, vp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "\x1b[image data]" {
		t.Errorf("output = %q, want %q", got, "\x1b[image data]")
	}
	if renderer.uploadCnt != 1 {
		t.Errorf("upload called %d times, want 1", renderer.uploadCnt)
	}
}

func TestRenderFrameUseCase_Execute_UploadsOnlyOnce(t *testing.T) {
	renderer := &mockRenderer{displayOut: "frame"}
	uc := NewRenderFrameUseCase(renderer, domain.MinimapConfig{})

	img := domain.NewImageEntity(image.NewRGBA(image.Rect(0, 0, 100, 100)), "test.png", "png")
	vp := domain.NewViewport(domain.ViewportConfig{
		ZoomStep: 0.1, PanStep: 0.05, MinZoom: 0.1, MaxZoom: 20.0,
	})
	vp.ImgWidth = 100
	vp.ImgHeight = 100
	vp.TermWidth = 80
	vp.TermHeight = 24

	for range 5 {
		_, err := uc.Execute(img, vp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if renderer.uploadCnt != 1 {
		t.Errorf("upload called %d times, want 1", renderer.uploadCnt)
	}
	if renderer.displayCnt != 5 {
		t.Errorf("display called %d times, want 5", renderer.displayCnt)
	}
}

func TestRenderFrameUseCase_Execute_UploadError(t *testing.T) {
	renderer := &mockRenderer{uploadErr: errors.New("upload failed")}
	uc := NewRenderFrameUseCase(renderer, domain.MinimapConfig{})

	img := domain.NewImageEntity(image.NewRGBA(image.Rect(0, 0, 100, 100)), "test.png", "png")
	vp := domain.NewViewport(domain.ViewportConfig{
		ZoomStep: 0.1, PanStep: 0.05, MinZoom: 0.1, MaxZoom: 20.0,
	})

	_, err := uc.Execute(img, vp)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRenderFrameUseCase_Execute_DisplayError(t *testing.T) {
	renderer := &mockRenderer{displayErr: errors.New("display failed")}
	uc := NewRenderFrameUseCase(renderer, domain.MinimapConfig{})

	img := domain.NewImageEntity(image.NewRGBA(image.Rect(0, 0, 100, 100)), "test.png", "png")
	vp := domain.NewViewport(domain.ViewportConfig{
		ZoomStep: 0.1, PanStep: 0.05, MinZoom: 0.1, MaxZoom: 20.0,
	})

	_, err := uc.Execute(img, vp)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRenderFrameUseCase_Execute_MinimapEnabled(t *testing.T) {
	renderer := &mockRenderer{
		displayOut:        "main",
		minimapDisplayOut: "minimap",
	}
	cfg := domain.MinimapConfig{Enabled: true, Size: 0.2}
	uc := NewRenderFrameUseCase(renderer, cfg)

	img := domain.NewImageEntity(image.NewRGBA(image.Rect(0, 0, 800, 600)), "test.png", "png")
	vp := domain.NewViewport(domain.ViewportConfig{
		ZoomStep: 0.1, PanStep: 0.05, MinZoom: 0.1, MaxZoom: 20.0,
	})
	vp.ImgWidth = 800
	vp.ImgHeight = 600
	vp.TermWidth = 80
	vp.TermHeight = 24
	vp.ZoomLevel = 2.0 // IsZoomed() = true

	got, err := uc.Execute(img, vp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != "mainminimap" {
		t.Errorf("output = %q, want %q", got, "mainminimap")
	}
	if renderer.minimapUploadCnt != 1 {
		t.Errorf("minimap upload called %d times, want 1", renderer.minimapUploadCnt)
	}
	if renderer.minimapDisplayCnt != 1 {
		t.Errorf("minimap display called %d times, want 1", renderer.minimapDisplayCnt)
	}
}

func TestRenderFrameUseCase_Execute_MinimapDisabledWhenNotZoomed(t *testing.T) {
	renderer := &mockRenderer{displayOut: "main"}
	cfg := domain.MinimapConfig{Enabled: true, Size: 0.2}
	uc := NewRenderFrameUseCase(renderer, cfg)

	img := domain.NewImageEntity(image.NewRGBA(image.Rect(0, 0, 800, 600)), "test.png", "png")
	vp := domain.NewViewport(domain.ViewportConfig{
		ZoomStep: 0.1, PanStep: 0.05, MinZoom: 0.1, MaxZoom: 20.0,
	})
	vp.ImgWidth = 800
	vp.ImgHeight = 600
	vp.TermWidth = 80
	vp.TermHeight = 24
	vp.ZoomLevel = 1.0 // IsZoomed() = false

	got, err := uc.Execute(img, vp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != "main" {
		t.Errorf("output = %q, want %q (no minimap when not zoomed)", got, "main")
	}
	if renderer.minimapUploadCnt != 0 {
		t.Errorf("minimap upload should not be called when not zoomed, got %d", renderer.minimapUploadCnt)
	}
}

func TestRenderFrameUseCase_Execute_MinimapConfigDisabled(t *testing.T) {
	renderer := &mockRenderer{displayOut: "main"}
	cfg := domain.MinimapConfig{Enabled: false, Size: 0.2}
	uc := NewRenderFrameUseCase(renderer, cfg)

	img := domain.NewImageEntity(image.NewRGBA(image.Rect(0, 0, 800, 600)), "test.png", "png")
	vp := domain.NewViewport(domain.ViewportConfig{
		ZoomStep: 0.1, PanStep: 0.05, MinZoom: 0.1, MaxZoom: 20.0,
	})
	vp.ImgWidth = 800
	vp.ImgHeight = 600
	vp.TermWidth = 80
	vp.TermHeight = 24
	vp.ZoomLevel = 2.0

	got, err := uc.Execute(img, vp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != "main" {
		t.Errorf("output = %q, want %q (minimap disabled)", got, "main")
	}
}

func TestRenderFrameUseCase_Execute_MinimapUploadsOnlyOnce(t *testing.T) {
	renderer := &mockRenderer{
		displayOut:        "main",
		minimapDisplayOut: "mm",
	}
	cfg := domain.MinimapConfig{Enabled: true, Size: 0.2}
	uc := NewRenderFrameUseCase(renderer, cfg)

	img := domain.NewImageEntity(image.NewRGBA(image.Rect(0, 0, 800, 600)), "test.png", "png")
	vp := domain.NewViewport(domain.ViewportConfig{
		ZoomStep: 0.1, PanStep: 0.05, MinZoom: 0.1, MaxZoom: 20.0,
	})
	vp.ImgWidth = 800
	vp.ImgHeight = 600
	vp.TermWidth = 80
	vp.TermHeight = 24
	vp.ZoomLevel = 2.0

	for range 5 {
		_, err := uc.Execute(img, vp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if renderer.minimapUploadCnt != 1 {
		t.Errorf("minimap upload called %d times, want 1", renderer.minimapUploadCnt)
	}
	if renderer.minimapDisplayCnt != 5 {
		t.Errorf("minimap display called %d times, want 5", renderer.minimapDisplayCnt)
	}
}

func TestRenderFrameUseCase_Execute_MinimapSkippedForSmallTerminal(t *testing.T) {
	renderer := &mockRenderer{displayOut: "main"}
	cfg := domain.MinimapConfig{Enabled: true, Size: 0.2}
	uc := NewRenderFrameUseCase(renderer, cfg)

	img := domain.NewImageEntity(image.NewRGBA(image.Rect(0, 0, 800, 600)), "test.png", "png")
	vp := domain.NewViewport(domain.ViewportConfig{
		ZoomStep: 0.1, PanStep: 0.05, MinZoom: 0.1, MaxZoom: 20.0,
	})
	vp.ImgWidth = 800
	vp.ImgHeight = 600
	vp.TermWidth = 20 // 20 * 0.2 = 4 cols, below minimum 5
	vp.TermHeight = 10
	vp.ZoomLevel = 2.0

	got, err := uc.Execute(img, vp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != "main" {
		t.Errorf("output = %q, want %q (minimap too small)", got, "main")
	}
	if renderer.minimapUploadCnt != 0 {
		t.Errorf("minimap upload should not be called for small terminal, got %d", renderer.minimapUploadCnt)
	}
}

func TestRenderFrameUseCase_Execute_MinimapUploadError(t *testing.T) {
	renderer := &mockRenderer{
		displayOut:       "main",
		minimapUploadErr: errors.New("minimap upload failed"),
	}
	cfg := domain.MinimapConfig{Enabled: true, Size: 0.2}
	uc := NewRenderFrameUseCase(renderer, cfg)

	img := domain.NewImageEntity(image.NewRGBA(image.Rect(0, 0, 800, 600)), "test.png", "png")
	vp := domain.NewViewport(domain.ViewportConfig{
		ZoomStep: 0.1, PanStep: 0.05, MinZoom: 0.1, MaxZoom: 20.0,
	})
	vp.ImgWidth = 800
	vp.ImgHeight = 600
	vp.TermWidth = 80
	vp.TermHeight = 24
	vp.ZoomLevel = 2.0

	_, err := uc.Execute(img, vp)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRenderFrameUseCase_Execute_MinimapDisplayError(t *testing.T) {
	renderer := &mockRenderer{
		displayOut:        "main",
		minimapDisplayErr: errors.New("minimap display failed"),
	}
	cfg := domain.MinimapConfig{Enabled: true, Size: 0.2}
	uc := NewRenderFrameUseCase(renderer, cfg)

	img := domain.NewImageEntity(image.NewRGBA(image.Rect(0, 0, 800, 600)), "test.png", "png")
	vp := domain.NewViewport(domain.ViewportConfig{
		ZoomStep: 0.1, PanStep: 0.05, MinZoom: 0.1, MaxZoom: 20.0,
	})
	vp.ImgWidth = 800
	vp.ImgHeight = 600
	vp.TermWidth = 80
	vp.TermHeight = 24
	vp.ZoomLevel = 2.0

	_, err := uc.Execute(img, vp)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRenderFrameUseCase_Execute_MinimapClearError(t *testing.T) {
	renderer := &mockRenderer{
		displayOut:        "main",
		minimapDisplayOut: "mm",
		minimapClearErr:   errors.New("clear failed"),
	}
	cfg := domain.MinimapConfig{Enabled: true, Size: 0.2}
	uc := NewRenderFrameUseCase(renderer, cfg)

	img := domain.NewImageEntity(image.NewRGBA(image.Rect(0, 0, 800, 600)), "test.png", "png")
	vp := domain.NewViewport(domain.ViewportConfig{
		ZoomStep: 0.1, PanStep: 0.05, MinZoom: 0.1, MaxZoom: 20.0,
	})
	vp.ImgWidth = 800
	vp.ImgHeight = 600
	vp.TermWidth = 80
	vp.TermHeight = 24
	vp.ZoomLevel = 2.0

	// First frame — minimap shown
	_, err := uc.Execute(img, vp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Zoom out — triggers ClearMinimap which returns error
	vp.ZoomLevel = 1.0
	_, err = uc.Execute(img, vp)
	if err == nil {
		t.Fatal("expected error from ClearMinimap, got nil")
	}
}

func TestRenderFrameUseCase_Execute_MinimapClearedOnZoomOut(t *testing.T) {
	renderer := &mockRenderer{
		displayOut:        "main",
		minimapDisplayOut: "mm",
	}
	cfg := domain.MinimapConfig{Enabled: true, Size: 0.2}
	uc := NewRenderFrameUseCase(renderer, cfg)

	img := domain.NewImageEntity(image.NewRGBA(image.Rect(0, 0, 800, 600)), "test.png", "png")
	vp := domain.NewViewport(domain.ViewportConfig{
		ZoomStep: 0.1, PanStep: 0.05, MinZoom: 0.1, MaxZoom: 20.0,
	})
	vp.ImgWidth = 800
	vp.ImgHeight = 600
	vp.TermWidth = 80
	vp.TermHeight = 24
	vp.ZoomLevel = 2.0 // zoomed in — minimap shown

	_, err := uc.Execute(img, vp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if renderer.minimapDisplayCnt != 1 {
		t.Fatalf("minimap should be displayed when zoomed, got %d", renderer.minimapDisplayCnt)
	}

	// Zoom out — minimap should be cleared
	vp.ZoomLevel = 1.0
	_, err = uc.Execute(img, vp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if renderer.minimapClearCnt != 1 {
		t.Errorf("minimap clear should be called once on zoom-out, got %d", renderer.minimapClearCnt)
	}

	// Another frame at zoom=1 — should NOT call clear again
	_, err = uc.Execute(img, vp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if renderer.minimapClearCnt != 1 {
		t.Errorf("minimap clear should not be called again, got %d", renderer.minimapClearCnt)
	}
}
