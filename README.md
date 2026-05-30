# dotty

Terminal UI for browsing dotfiles under your home directory. Run it from anywhere — it always scans `~`, shows an expandable tree on the left, file preview on the right, and a fuzzy finder to jump to any path quickly.

```
┌─────────────────────┐ ┌──────────────────────────────────────────┐
│  dotty              │ │  init.lua                           42%  │
│  /                  │ │────────────────────────────────────────  │
│─────────────────────│ │   1  -- neovim config                    │
│ ▸ .config/          │ │   2                                      │
│   ▾ .config/nvim/   │ │   3  vim.opt.number = true               │
│     init.lua        │ │   4  vim.opt.relativenumber = true       │
│   .zshrc            │ │                                          │
└─────────────────────┘ └──────────────────────────────────────────┘
  12/48 entries    j/k move  Enter open  / find  y copy  Tab preview  q quit
```

## Install

Requires Go 1.22+.

```bash
git clone https://github.com/RiccardoCataldi/dotty.git
cd dotty
go build -o dotty .
mv dotty ~/bin/   # or another directory on your PATH
```

### Clipboard (optional)

Copy path (`y`) uses the system clipboard:

| OS | Tool |
|----|------|
| macOS | `pbcopy` (built in) |
| Linux | `xclip` or `xsel` |

## Usage

```bash
dotty
```

The current working directory is ignored; only `$HOME` (or the override below) is scanned.

```bash
dotty --home /path/to/home   # custom root (testing or alternate home)
```

## What gets scanned

Top-level entries in the home directory whose names start with `.` (dotfiles and dot-directories). Directories load children lazily when you expand them in the tree.

Skipped dot-directories (noisy or sensitive):

`.cache` `.local` `.mozilla` `.dbus` `.pki` `.gnupg` `.var` `.snap`

To change the skip list, edit `defaultBlocklist` in `scan.go`.

## Keybindings

### Tree and preview

| Key | Action |
|-----|--------|
| `j` / `k`, `↓` / `↑` | Move selection in the tree |
| `g` / `G` | Jump to first / last visible row |
| `Enter`, `l`, `→` | Expand directory (loads children) or move into selection |
| `h`, `←` | Collapse parent or current directory |
| `Tab` | Toggle focus between tree and preview |
| `j` / `k`, `d` / `u` | Scroll preview (when preview focused) |
| `y` | Copy absolute path of selected **file** to clipboard |
| `q`, `Ctrl+C` | Quit |

### Fuzzy finder (`/`)

Opens an overlay to search all files under scanned dot trees (subdirectories included once expanded or picked).

| Key | Action |
|-----|--------|
| Type | Filter paths (subsequence match, ranked) |
| `j` / `k` | Move through results |
| `Enter` | Jump to file in tree (expands parents) |
| `Esc`, `q` | Close finder |

## Preview

- **Files:** syntax-free text preview (size-capped; large/binary files show a short message).
- **Directories:** placeholder in the preview pane.

## License

MIT
