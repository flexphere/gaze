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

func TestRenderFrameUseCase_Execute_Success(t *testing.T) {
	renderer := &mockRenderer{displayOut: "\x1b[image data]"}
	uc := NewRenderFrameUseCase(renderer)

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
	uc := NewRenderFrameUseCase(renderer)

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
	uc := NewRenderFrameUseCase(renderer)

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
	uc := NewRenderFrameUseCase(renderer)

	img := domain.NewImageEntity(image.NewRGBA(image.Rect(0, 0, 100, 100)), "test.png", "png")
	vp := domain.NewViewport(domain.ViewportConfig{
		ZoomStep: 0.1, PanStep: 0.05, MinZoom: 0.1, MaxZoom: 20.0,
	})

	_, err := uc.Execute(img, vp)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
