package main

import (
	"errors"
	"fmt"
	"math"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/flexphere/gaze/internal/adapter/config"
	"github.com/flexphere/gaze/internal/adapter/renderer"
	"github.com/flexphere/gaze/internal/adapter/tui"
	"github.com/flexphere/gaze/internal/domain"
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
	var staticMode bool

	cmd := &cobra.Command{
		Use:          "gaze <image>",
		Short:        "Terminal image viewer with zoom and pan",
		Long:         "gaze is a terminal image viewer that supports zoom, pan, and mouse interaction using Kitty Graphics Protocol.",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if staticMode {
			return runStatic(args)
		}
		return runViewer(cmd, args)
	}

	cmd.Flags().BoolVar(&staticMode, "static", false, "Display image and exit without interactive mode")
	cmd.Version = version

	return cmd
}

func runStatic(args []string) error {
	imagePath := args[0]

	// Load image
	imageLoader := filesystem.NewImageLoader()
	loadImageUC := usecase.NewLoadImageUseCase(imageLoader)
	img, err := loadImageUC.Execute(imagePath)
	if err != nil {
		return fmt.Errorf("loading image: %w", err)
	}

	// Query terminal dimensions
	cols, rows := tui.QueryTerminalSize()
	if cols <= 0 || rows <= 0 {
		return fmt.Errorf("unable to determine terminal size")
	}

	// Calculate native cell dimensions (1:1 pixel mapping), capped at terminal width
	cellW, cellH := tui.QueryCellSize()
	nativeCols := int(math.Ceil(float64(img.Width) / cellW))
	nativeRows := int(math.Ceil(float64(img.Height) / cellH))
	displayCols := nativeCols
	if displayCols > cols {
		displayCols = cols
	}

	// Build viewport sized to native display area
	vp := &domain.Viewport{
		TermWidth:       displayCols,
		TermHeight:      nativeRows,
		ImgWidth:        img.Width,
		ImgHeight:       img.Height,
		CellAspectRatio: cellH / cellW,
		ZoomLevel:       1.0,
	}
	vp.FitToWindow()

	// Upload and display
	kittyRenderer := renderer.NewKittyRenderer()
	if err := kittyRenderer.Upload(img); err != nil {
		return fmt.Errorf("uploading image: %w", err)
	}

	output, err := kittyRenderer.Display(vp)
	if err != nil {
		return fmt.Errorf("displaying image: %w", err)
	}

	fmt.Print(output)
	return nil
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
