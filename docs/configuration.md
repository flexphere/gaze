# Configuration

gaze is configured via a TOML file located at:

```
~/.config/gaze/config.toml
```

If the file does not exist, all default values are used. You only need to specify the settings you want to change.

## Keybindings

Each action can be bound to one or more keys.

```toml
[keybindings]
pan_up       = ["k", "up"]
pan_down     = ["j", "down"]
pan_left     = ["h", "left"]
pan_right    = ["l", "right"]
zoom_in      = ["+", "="]
zoom_out     = ["-", "_"]
reset_view   = ["0", "r"]
fit_to_window = ["f"]
quit         = ["q", "ctrl+c", "esc"]
```

### Key Format

- Single characters: `"k"`, `"f"`, `"+"`, `"-"`
- Arrow keys: `"up"`, `"down"`, `"left"`, `"right"`
- Special keys: `"esc"`, `"enter"`, `"tab"`, `"backspace"`
- Modifier combinations: `"ctrl+c"`, `"ctrl+q"`

### Defaults

| Action | Default Keys |
|--------|-------------|
| Pan up | `k`, `Up` |
| Pan down | `j`, `Down` |
| Pan left | `h`, `Left` |
| Pan right | `l`, `Right` |
| Zoom in | `+`, `=` |
| Zoom out | `-`, `_` |
| Reset view | `0`, `r` |
| Fit to window | `f` |
| Quit | `q`, `Ctrl+C`, `Esc` |

## Mouse

```toml
[mouse]
drag_to_pan         = true     # Enable drag-to-pan
scroll_to_zoom      = true     # Enable scroll-to-zoom
scroll_sensitivity  = 0.15     # Zoom amount per scroll tick (0.0 - 1.0)
```

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `drag_to_pan` | bool | `true` | Left-click drag to pan the image |
| `scroll_to_zoom` | bool | `true` | Scroll wheel to zoom in/out at cursor position |
| `scroll_sensitivity` | float | `0.15` | How much to zoom per scroll tick. Higher = faster zoom |

## Viewport

```toml
[viewport]
zoom_step = 0.1    # Zoom increment per keypress (10%)
pan_step  = 0.05   # Pan distance as fraction of visible area (5%)
min_zoom  = 0.1    # Minimum zoom level (10%)
max_zoom  = 20.0   # Maximum zoom level (2000%)
```

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `zoom_step` | float | `0.1` | Zoom change per keyboard press. `0.1` = 10% per press |
| `pan_step` | float | `0.05` | Pan distance as fraction of visible area. `0.05` = 5% per press |
| `min_zoom` | float | `0.1` | Minimum zoom level. `0.1` = 10% (zoomed out) |
| `max_zoom` | float | `20.0` | Maximum zoom level. `20.0` = 2000% (zoomed in) |

## Examples

### WASD Navigation

```toml
[keybindings]
pan_up    = ["w", "up"]
pan_down  = ["s", "down"]
pan_left  = ["a", "left"]
pan_right = ["d", "right"]
```

### Disable Mouse

```toml
[mouse]
drag_to_pan    = false
scroll_to_zoom = false
```

### Fine-grained Zoom

```toml
[viewport]
zoom_step = 0.05          # 5% per keypress (default: 10%)

[mouse]
scroll_sensitivity = 0.05  # Slower scroll zoom
```

### Fast Navigation

```toml
[viewport]
zoom_step = 0.25   # 25% per keypress
pan_step  = 0.15   # 15% of visible area per pan
```
