package tui

const (
	defaultCellWidth  = 8.0
	defaultCellHeight = 16.0
)

// queryCellSize returns default cell pixel dimensions on Windows.
// TIOCGWINSZ is not available on Windows.
func queryCellSize() (cellWidth, cellHeight float64) {
	return defaultCellWidth, defaultCellHeight
}
