package usecase

import "github.com/flexphere/gaze/internal/domain"

// ViewportControlUseCase controls viewport state.
type ViewportControlUseCase interface {
	ZoomIn(vp *domain.Viewport)
	ZoomOut(vp *domain.Viewport)
	ZoomAtPoint(vp *domain.Viewport, delta float64, termX, termY int)
	PanUp(vp *domain.Viewport)
	PanDown(vp *domain.Viewport)
	PanLeft(vp *domain.Viewport)
	PanRight(vp *domain.Viewport)
	PanByPixels(vp *domain.Viewport, dx, dy float64)
	ResetView(vp *domain.Viewport)
	FitToWindow(vp *domain.Viewport)
}

type viewportControlUseCase struct{}

// NewViewportControlUseCase creates a new ViewportControlUseCase.
func NewViewportControlUseCase() ViewportControlUseCase {
	return &viewportControlUseCase{}
}

func (uc *viewportControlUseCase) ZoomIn(vp *domain.Viewport) {
	vp.ZoomIn()
}

func (uc *viewportControlUseCase) ZoomOut(vp *domain.Viewport) {
	vp.ZoomOut()
}

func (uc *viewportControlUseCase) ZoomAtPoint(vp *domain.Viewport, delta float64, termX, termY int) {
	vp.ZoomAt(delta, termX, termY)
}

func (uc *viewportControlUseCase) PanUp(vp *domain.Viewport) {
	vp.PanByStep(0, -1)
}

func (uc *viewportControlUseCase) PanDown(vp *domain.Viewport) {
	vp.PanByStep(0, 1)
}

func (uc *viewportControlUseCase) PanLeft(vp *domain.Viewport) {
	vp.PanByStep(-1, 0)
}

func (uc *viewportControlUseCase) PanRight(vp *domain.Viewport) {
	vp.PanByStep(1, 0)
}

func (uc *viewportControlUseCase) PanByPixels(vp *domain.Viewport, dx, dy float64) {
	vp.Pan(dx, dy)
}

func (uc *viewportControlUseCase) ResetView(vp *domain.Viewport) {
	vp.FitToWindow()
}

func (uc *viewportControlUseCase) FitToWindow(vp *domain.Viewport) {
	vp.FitToWindow()
}
