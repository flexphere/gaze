package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/flexphere/gaze/internal/adapter/config"
	"github.com/flexphere/gaze/internal/adapter/renderer"
	"github.com/flexphere/gaze/internal/adapter/tui"
	"github.com/flexphere/gaze/internal/infrastructure/filesystem"
	"github.com/flexphere/gaze/internal/usecase"
)

var (
	version      = "dev"
	rendererFlag string
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "gaze <image>",
		Short:        "Terminal image viewer with zoom and pan",
		Long:         "gaze is a terminal image viewer that supports zoom, pan, and mouse interaction using Kitty Graphics Protocol or Sixel.",
		Args:         cobra.ExactArgs(1),
		RunE:         runViewer,
		SilenceUsage: true,
	}

	cmd.Version = version
	cmd.Flags().StringVar(&rendererFlag, "renderer", "auto", "renderer backend: auto, kitty, sixel")

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

	// Select renderer
	r := selectRenderer(rendererFlag)
	vpCtrl := usecase.NewViewportControlUseCase()
	renderFrameUC := usecase.NewRenderFrameUseCase(r)

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

	// Clean up renderer resources
	if err := r.Clear(); err != nil {
		return fmt.Errorf("clearing renderer: %w", err)
	}

	return nil
}

// selectRenderer chooses the renderer backend based on the flag value.
func selectRenderer(flag string) usecase.RendererPort {
	switch flag {
	case "kitty":
		return renderer.NewKittyRenderer()
	case "sixel":
		return renderer.NewSixelRenderer()
	default:
		return detectRenderer()
	}
}

// detectRenderer auto-detects the best renderer for the current terminal.
func detectRenderer() usecase.RendererPort {
	if isTmux() {
		return renderer.NewSixelRenderer()
	}
	return renderer.NewKittyRenderer()
}

// isTmux returns true if running inside a tmux session.
func isTmux() bool {
	if os.Getenv("TMUX") != "" {
		return true
	}
	term := os.Getenv("TERM")
	return strings.HasPrefix(term, "screen") || strings.HasPrefix(term, "tmux")
}
