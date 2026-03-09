package main

import (
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/flexphere/gaze/internal/adapter/config"
	"github.com/flexphere/gaze/internal/adapter/renderer"
	"github.com/flexphere/gaze/internal/adapter/tui"
	"github.com/flexphere/gaze/internal/domain"
	"github.com/flexphere/gaze/internal/infrastructure/ffmpeg"
	"github.com/flexphere/gaze/internal/infrastructure/filesystem"
	"github.com/flexphere/gaze/internal/usecase"
)

var version = "dev"

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "gaze <file>",
		Short:        "Terminal image/video viewer with zoom and pan",
		Long:         "gaze is a terminal media viewer that supports zoom, pan, and mouse interaction using Kitty Graphics Protocol. Video playback requires ffmpeg.",
		Args:         cobra.ExactArgs(1),
		RunE:         runViewer,
		SilenceUsage: true,
	}

	cmd.Version = version

	return cmd
}

func runViewer(_ *cobra.Command, args []string) error {
	filePath := args[0]

	// Load configuration
	configLoader := config.NewTOMLLoader()
	cfg, err := configLoader.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Detect media type and run appropriate viewer
	if ffmpeg.Available() && ffmpeg.IsVideo(filePath) {
		return runVideoViewer(filePath, cfg)
	}
	return runImageViewer(filePath, cfg)
}

func runImageViewer(filePath string, cfg *domain.Config) error {
	// Load image — use ffmpeg for extended format support when available
	var imageLoader usecase.ImageLoaderPort
	stdLoader := filesystem.NewImageLoader()
	if ffmpeg.Available() {
		imageLoader = ffmpeg.NewImageLoader(stdLoader)
	} else {
		imageLoader = stdLoader
	}
	loadImageUC := usecase.NewLoadImageUseCase(imageLoader)
	img, err := loadImageUC.Execute(filePath)
	if err != nil {
		return fmt.Errorf("loading image: %w", err)
	}

	// Create renderer and use cases
	kittyRenderer := renderer.NewKittyRenderer()
	vpCtrl := usecase.NewViewportControlUseCase()
	renderFrameUC := usecase.NewRenderFrameUseCase(kittyRenderer, cfg.Minimap)

	// Create TUI model
	model := tui.NewModel(img, cfg, vpCtrl, renderFrameUC)

	// Run Bubbletea program
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("running viewer: %w", err)
	}

	// Clean up Kitty graphics
	var errs []error
	if err := kittyRenderer.ClearMinimap(); err != nil {
		errs = append(errs, fmt.Errorf("clearing minimap: %w", err))
	}
	if err := kittyRenderer.Clear(); err != nil {
		errs = append(errs, fmt.Errorf("clearing renderer: %w", err))
	}

	return errors.Join(errs...)
}

func runVideoViewer(filePath string, cfg *domain.Config) error {
	// Open video decoder
	decoder := ffmpeg.NewVideoDecoder()
	videoInfo, err := decoder.Open(filePath)
	if err != nil {
		return fmt.Errorf("opening video: %w", err)
	}

	// Get first frame to display immediately
	firstFrame, err := decoder.NextFrame()
	if err != nil {
		_ = decoder.Close() //nolint:errcheck // best-effort cleanup
		return fmt.Errorf("reading first frame: %w", err)
	}
	firstImg := domain.NewImageEntity(firstFrame, filePath, "video")

	// Create renderer and use cases
	kittyRenderer := renderer.NewKittyRenderer()
	kittyRenderer.SetVideoMode(true)
	renderFrameUC := usecase.NewRenderFrameUseCase(kittyRenderer, cfg.Minimap)

	// Create video TUI model
	model := tui.NewVideoModel(firstImg, decoder, videoInfo, cfg, renderFrameUC)

	// Run Bubbletea program
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("running viewer: %w", err)
	}

	// Clean up
	var errs []error
	if err := decoder.Close(); err != nil {
		errs = append(errs, fmt.Errorf("closing decoder: %w", err))
	}
	if err := kittyRenderer.Clear(); err != nil {
		errs = append(errs, fmt.Errorf("clearing renderer: %w", err))
	}

	return errors.Join(errs...)
}
