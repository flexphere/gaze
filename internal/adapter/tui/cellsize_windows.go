package tui

const (
	defaultCellWidth  = 8.0
	defaultCellHeight = 16.0
)

// QueryCellSize returns default cell pixel dimensions on Windows.
// TIOCGWINSZ is not available on Windows.
func QueryCellSize() (cellWidth, cellHeight float64) {
	return defaultCellWidth, defaultCellHeight
}

// QueryTerminalSize returns (0, 0) on Windows as TIOCGWINSZ is not available.
func QueryTerminalSize() (cols, rows int) {
	return 0, 0
}
