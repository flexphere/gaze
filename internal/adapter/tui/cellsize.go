//go:build !windows

package tui

import (
	"os"

	"golang.org/x/sys/unix"
)

const (
	defaultCellWidth  = 8.0
	defaultCellHeight = 16.0
)

// queryCellSize returns the terminal cell pixel dimensions using TIOCGWINSZ.
// Falls back to 8x16 if pixel dimensions are unavailable.
func queryCellSize() (cellWidth, cellHeight float64) {
	ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return defaultCellWidth, defaultCellHeight
	}

	if ws.Xpixel == 0 || ws.Ypixel == 0 || ws.Col == 0 || ws.Row == 0 {
		return defaultCellWidth, defaultCellHeight
	}

	cellWidth = float64(ws.Xpixel) / float64(ws.Col)
	cellHeight = float64(ws.Ypixel) / float64(ws.Row)

	return cellWidth, cellHeight
}
