package usecase

import (
	"errors"
	"image"
	"testing"

	"github.com/flexphere/gaze/internal/domain"
)

type mockImageLoader struct {
	img *domain.ImageEntity
	err error
}

func (m *mockImageLoader) Load(_ string) (*domain.ImageEntity, error) {
	return m.img, m.err
}

func TestLoadImageUseCase_Execute_Success(t *testing.T) {
	want := domain.NewImageEntity(
		image.NewRGBA(image.Rect(0, 0, 100, 100)),
		"test.png",
		"png",
	)
	loader := &mockImageLoader{img: want}
	uc := NewLoadImageUseCase(loader)

	got, err := uc.Execute("test.png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Error("returned image does not match")
	}
}

func TestLoadImageUseCase_Execute_Error(t *testing.T) {
	loader := &mockImageLoader{err: errors.New("file not found")}
	uc := NewLoadImageUseCase(loader)

	_, err := uc.Execute("nonexistent.png")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
