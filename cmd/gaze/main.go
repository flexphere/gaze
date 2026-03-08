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
		Use:          "gaze <image>",
		Short:        "Terminal image viewer with zoom and pan",
		Long:         "gaze is a terminal image viewer that supports zoom, pan, and mouse interaction using Kitty Graphics Protocol.",
		Args:         cobra.ExactArgs(1),
		RunE:         runViewer,
		SilenceUsage: true,
	}

	cmd.Version = version

	return cmd
}

func runViewer(_ *cobra.Command, args []string) error {
	imagePath := args[0]

	// Load configuration
	configLoader := config.NewTOMLLoader()
	cfg, err := configLoader.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Load image
	imageLoader := filesystem.NewImageLoader()
	loadImageUC := usecase.NewLoadImageUseCase(imageLoader)
	img, err := loadImageUC.Execute(imagePath)
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

	// Clean up Kitty graphics — always attempt both cleanups
	var errs []error
	if err := kittyRenderer.ClearMinimap(); err != nil {
		errs = append(errs, fmt.Errorf("clearing minimap: %w", err))
	}
	if err := kittyRenderer.Clear(); err != nil {
		errs = append(errs, fmt.Errorf("clearing renderer: %w", err))
	}

	return errors.Join(errs...)
}
