# gaze

A terminal image viewer with zoom and pan support, powered by the [Kitty Graphics Protocol](https://sw.kovidgoyal.net/kitty/graphics-protocol/).

## Features

- **High-performance rendering** - Image is uploaded once; zoom/pan updates only change the display region (~2.5ms/frame)
- **Keyboard & mouse controls** - Vim-style keys, arrow keys, mouse drag-to-pan, scroll-to-zoom
- **Configurable keybindings** - Customize all key mappings via TOML config file
- **Multiple image formats** - PNG, JPEG, GIF, BMP, TIFF, WebP
- **Clean Architecture** - Modular design for easy extension and testing

## Supported Terminals

| Terminal | Status |
|----------|--------|
| [Kitty](https://sw.kovidgoyal.net/kitty/) | Supported |
| [Ghostty](https://ghostty.org/) | Supported |
| [WezTerm](https://wezfurlong.org/wezterm/) | Supported |

> Other terminals supporting the Kitty Graphics Protocol should also work.

## Installation

```bash
go install github.com/flexphere/gaze/cmd/gaze@latest
```

Or build from source:

```bash
git clone https://github.com/flexphere/gaze.git
cd gaze
make build
# Binary is at ./bin/gaze
```

## Usage

```bash
gaze <image-file>
```

```bash
# Examples
gaze photo.png
gaze screenshot.jpg
gaze animation.gif
gaze image.webp
```

## Controls

### Keyboard

| Key | Action |
|-----|--------|
| `h` / `Left` | Pan left |
| `j` / `Down` | Pan down |
| `k` / `Up` | Pan up |
| `l` / `Right` | Pan right |
| `+` / `=` | Zoom in |
| `-` / `_` | Zoom out |
| `f` | Fit to window |
| `0` / `r` | Reset view |
| `q` / `Esc` / `Ctrl+C` | Quit |

### Mouse

| Action | Effect |
|--------|--------|
| Drag | Pan the image |
| Scroll up | Zoom in (at cursor position) |
| Scroll down | Zoom out (at cursor position) |

## Configuration

gaze reads configuration from `~/.config/gaze/config.toml`. If the file does not exist, default values are used.

```toml
[keybindings]
pan_up    = ["k", "up"]
pan_down  = ["j", "down"]
pan_left  = ["h", "left"]
pan_right = ["l", "right"]
zoom_in   = ["+", "="]
zoom_out  = ["-", "_"]
reset_view = ["0", "r"]
fit_to_window = ["f"]
quit      = ["q", "ctrl+c", "esc"]

[mouse]
drag_to_pan = true
scroll_to_zoom = true
scroll_sensitivity = 0.15

[viewport]
zoom_step = 0.1
pan_step  = 0.05
```

See [docs/configuration.md](docs/configuration.md) for full details.

## Architecture

gaze follows Clean Architecture with clear separation of concerns:

```
cmd/gaze/          Entry point + dependency injection
internal/
  domain/          Entities (ImageEntity, Viewport, Config) - no external deps
  usecase/         Port interfaces + use cases (LoadImage, ViewportControl, RenderFrame)
  adapter/
    tui/           Bubbletea TUI (Model/Update/View, KeyMap, StatusBar)
    config/        TOML configuration loader
    renderer/      Kitty Graphics Protocol implementation
  infrastructure/
    filesystem/    Image file loading (PNG, JPEG, GIF, BMP, TIFF, WebP)
```

## Development

```bash
make build    # Build binary
make test     # Run tests with race detector
make lint     # Run golangci-lint
make ci       # Run lint + test + build
make clean    # Remove build artifacts
```

## License

[MIT](LICENSE)
