package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strings"

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
	var (
		staticMode   bool
		rendererType string
	)

	cmd := &cobra.Command{
		Use:          "gaze <image>",
		Short:        "Terminal image viewer with zoom and pan",
		Long:         "gaze is a terminal image viewer that supports zoom, pan, and mouse interaction using Kitty or Sixel graphics protocol.",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if staticMode {
			return runStatic(args, rendererType)
		}
		return runViewer(cmd, args, rendererType)
	}

	cmd.Flags().BoolVar(&staticMode, "static", false, "Display image and exit without interactive mode")
	cmd.Flags().StringVarP(&rendererType, "renderer", "r", "kitty", "Renderer protocol (kitty, sixel)")
	cmd.Version = version

	return cmd
}

func createRenderer(rendererType string, cellW, cellH float64) (usecase.RendererPort, error) {
	switch rendererType {
	case "kitty":
		return renderer.NewKittyRenderer(), nil
	case "sixel":
		return renderer.NewSixelRenderer(cellW, cellH), nil
	default:
		return nil, fmt.Errorf("unknown renderer type %q: supported values are kitty, sixel", rendererType)
	}
}

func runStatic(args []string, rendererType string) error {
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
		return fmt.Errorf("determining terminal size: --static requires a TTY")
	}

	// Calculate native cell dimensions (1:1 pixel mapping), capped at terminal size
	cellW, cellH := tui.QueryCellSize()
	nativeCols := int(math.Ceil(float64(img.Width) / cellW))
	nativeRows := int(math.Ceil(float64(img.Height) / cellH))
	displayCols := nativeCols
	if displayCols > cols {
		displayCols = cols
	}
	displayRows := nativeRows
	if displayRows > rows {
		displayRows = rows
	}

	// Build viewport via constructor to inherit default zoom/pan limits
	cfg := domain.DefaultConfig()
	vp := domain.NewViewport(cfg.Viewport)
	vp.SetCellAspectRatio(cellH / cellW)
	vp.SetTerminalSize(displayCols, displayRows)
	vp.SetImageSize(img.Width, img.Height)

	// Upload and display
	imgRenderer, err := createRenderer(rendererType, cellW, cellH)
	if err != nil {
		return err
	}
	if err := imgRenderer.Upload(img); err != nil {
		return fmt.Errorf("uploading image: %w", err)
	}

	output, err := imgRenderer.Display(vp)
	if err != nil {
		return fmt.Errorf("displaying image: %w", err)
	}

	// Strip cursor-to-home (\x1b[H) used by interactive mode; static displays inline
	output = strings.TrimPrefix(output, "\x1b[H")

	// Calculate actual display rows to position cursor below the image
	cellAspect := vp.CellAspect()
	imgAspect := float64(img.Width) / float64(img.Height)
	termAspect := float64(displayCols) / (float64(displayRows) * cellAspect)
	dispRows := displayRows
	if imgAspect > termAspect {
		dispRows = int(math.Round(float64(displayCols) / imgAspect / cellAspect))
	}
	if dispRows <= 0 {
		dispRows = 1
	}

	// Display image and move cursor below it
	fmt.Print(output)
	fmt.Printf("\x1b[%dB\n", dispRows)
	return nil
}

func runViewer(_ *cobra.Command, args []string, rendererType string) error {
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
	cellW, cellH := tui.QueryCellSize()
	imgRenderer, err := createRenderer(rendererType, cellW, cellH)
	if err != nil {
		return err
	}
	vpCtrl := usecase.NewViewportControlUseCase()
	renderFrameUC := usecase.NewRenderFrameUseCase(imgRenderer, cfg.Minimap)

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

	// Clean up renderer — always attempt both cleanups
	var errs []error
	if err := imgRenderer.ClearMinimap(); err != nil {
		errs = append(errs, fmt.Errorf("clearing minimap: %w", err))
	}
	if err := imgRenderer.Clear(); err != nil {
		errs = append(errs, fmt.Errorf("clearing renderer: %w", err))
	}

	return errors.Join(errs...)
}
