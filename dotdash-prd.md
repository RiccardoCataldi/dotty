# PRD — dotdash
**Product Requirements Document**
Version 1.1 — TUI dashboard for dotfile navigation

---

## 1. Overview

**dotdash** is a terminal-based dashboard for navigating and managing dotfiles on a Unix/Linux system. It targets developers who accumulate many configuration files and folders in their home directory (`.zshrc`, `.gitconfig`, `.cursor/`, `.aws/`, `.config/nvim/`, etc.) and want a fast, keyboard-driven way to find, view, and manage them — without leaving the terminal, regardless of which directory they are currently in.

The tool is invoked with a single command (e.g. `dotdash`) from **any working directory** in the terminal. It always operates on `~` — the current directory is irrelevant.

The UX model is **Telescope/fzf-style**: a fuzzy-searchable list on the left with a live file preview on the right, all inside a single TUI screen. Browse, preview, and copy file paths via keyboard shortcuts.

---

## 2. Goals

- Provide a single-command entry point to browse and manage all dotfiles in `~`, callable from any directory
- Make it faster to locate and view config files than doing it manually in the terminal
- Browse and preview dotfiles without leaving the TUI
- Zero configuration required out of the box
- Feel native to a terminal workflow (works inside tmux, over SSH, no browser needed)

---

## 3. Non-Goals

- No dotfile syncing or version control (not a chezmoi replacement)
- No file creation or in-app editing from scratch
- No multi-machine management
- No GUI / web interface

---

## 4. Tech Stack

| Layer | Choice | Reason |
|---|---|---|
| Language | Go | Fast compilation, single static binary, strong TUI ecosystem |
| TUI framework | Bubble Tea (`github.com/charmbracelet/bubbletea`) | Industry standard for Go TUIs, powers lazygit/lazydocker |
| TUI components | Bubbles (`github.com/charmbracelet/bubbles`) | Provides viewport, textinput, list primitives |
| Styling | Lipgloss (`github.com/charmbracelet/lipgloss`) | CSS-like styling for terminal layouts |
| Min Go version | 1.22 | |

The final artifact must be a **single compiled binary** with no runtime dependencies.

---

## 5. Layout

```
┌─────────────────────┐ ┌──────────────────────────────────────────┐
│  dotdash            │ │  .config/nvim/init.lua              42%  │
│  / nvi█             │ │────────────────────────────────────────  │
│─────────────────────│ │   1  -- neovim config                    │
│ ▸ .config/nvim/     │ │   2                                      │
│   .config/nvim/i... │ │   3  vim.opt.number = true               │
│   .config/nvim/l... │ │   4  vim.opt.relativenumber = true       │
│   .vimrc            │ │   5                                      │
│                     │ │   6  -- [plugins]                        │
│                     │ │   7  require("lazy").setup({})           │
│                     │ │                                          │
└─────────────────────┘ └──────────────────────────────────────────┘
  3/47 files    j/k navigate  / search  Tab panel  q quit
```

The screen is divided into three zones:

**Left panel (1/3 width)** — file list with fuzzy search input at the top.
**Right panel (2/3 width)** — file preview with line numbers and scroll percentage.
**Status bar (1 line)** — match count and keybinding hints.

Both panels have rounded borders. The active panel (receiving keyboard input) renders its border in the accent color (purple `#7D56F4`); the inactive one uses a dim color.

---

## 6. File Scanning

### 6.1 What to scan

Scan the user's home directory (`os.UserHomeDir()`) for **immediate children** whose name starts with `.`:

- **Dot files** — e.g. `.zshrc`, `.gitconfig`, `.bashrc`, `.tmux.conf`
- **Dot directories** — e.g. `.config/`, `.cursor/`, `.aws/`, `.ssh/`

For each dot **directory**, also scan one level deep and include its direct children (both files and subdirectories) as individual list entries.

Entry display format:
- File directly in `~`: `.zshrc`
- Directory: `.config/`
- File inside a dot dir: `.config/nvim/init.lua`
- Subdirectory inside a dot dir: `.config/nvim/`

### 6.2 Directories to skip (default blocklist)

The following directories generate too much noise and should be excluded by default:

```
.cache
.local
.mozilla
.dbus
.pki
.gnupg
.var
.snap
```

### 6.3 File size limit for preview

Do not attempt to preview files larger than **512 KB**. Show a message instead: `"File too large to preview (> 512 KB)"`.

---

## 7. Fuzzy Search

Pressing `/` activates search mode. A cursor (`█`) appears after the query string in the list panel header.

**Matching algorithm:** fuzzy — a file matches if all characters of the query appear in the filename in order (not necessarily contiguous). This is the same algorithm used by fzf and Telescope.

Example: query `nvi` matches `.config/nvim/init.lua` because n→v→i appear in order.

**Matched characters** are highlighted in green (`#73F59F`) within the list entry.

The list updates in real time as the user types. The cursor resets to position 0 on every query change.

Pressing `Esc` clears the query and restores the full list. Pressing `Enter` confirms the search and exits search mode (focus stays in list, query remains active as a filter).

---

## 8. Preview Panel

### 8.1 File preview

Displays the raw content of the selected file with:

- **Line numbers** — right-aligned, dim color, 4-digit width. Format: `   1  <content>`
- **Scroll percentage** — shown in the panel title bar, e.g. `42%`. Hidden if the entire file fits on screen.
- **Maximum lines rendered** — 500. If the file exceeds this, append a dim line: `... (N more lines not shown)`

### 8.2 Lightweight syntax highlighting

Apply the following rules in order (first match wins, no external library):

| Pattern | Style |
|---|---|
| Line starts with `#`, `//`, or `--` (after trimming) | Green `#6A9955` — comment |
| Line starts with `[` or `[[` (after trimming) | Blue `#569CD6`, bold — section header (TOML/INI) |
| All other lines | Default terminal foreground |

### 8.3 Directory preview

If the selected entry is a directory, show a plain listing of its contents:
- Subdirectories in blue `#569CD6` with trailing `/`
- Files in default foreground
- No line numbers

### 8.4 Scroll

When focus is switched to the preview panel (`Tab`), `j`/`k` scroll the content line by line. `d`/`u` scroll half a page. The Bubbles `viewport` component handles this.

---

## 9. Keyboard Controls

### 9.1 Navigation

| Key | Context | Action |
|---|---|---|
| `j` / `↓` | List focused | Move selection down |
| `k` / `↑` | List focused | Move selection up |
| `g` | List focused | Jump to first entry |
| `G` | List focused | Jump to last entry |
| `/` | Any | Enter search mode |
| `Esc` | Search mode | Exit search, clear query |
| `Esc` | Normal mode | Clear active query filter |
| `Enter` | Search mode | Confirm search, exit search mode |
| `Tab` | Any | Toggle focus between list and preview panel |
| `j` / `k` | Preview focused | Scroll preview up/down |
| `d` / `u` | Preview focused | Scroll preview half-page down/up |
| `q` / `Ctrl+C` | Any | Quit |

### 9.2 Actions

| Key | Context | Action |
|---|---|---|
| `y` | List focused, file selected | **Copy path** — copy the absolute path of the selected file to the system clipboard (uses `xclip`, `xsel`, or `pbcopy` depending on OS; show a brief status message if clipboard tool is not available) |

---

## 10. Status Bar

Single line at the bottom of the screen. Two sections:

**Left:** `{matched}/{total} files` — e.g. `3/47 files`

**Right:** condensed keybinding hints — `j/k navigate  / search  y copy path  Tab panel  q quit`

Both sections use a dim/italic style. The gap between them is filled with spaces to align right section to the right edge.

---

## 11. Styling Tokens

| Token | Value | Usage |
|---|---|---|
| Accent | `#7D56F4` | Active panel border, title, active item arrow, search prompt |
| Match highlight | `#73F59F` | Fuzzy-matched characters in list |
| Comment | `#6A9955` | Comment lines in preview |
| Keyword/section | `#569CD6` | Section headers in preview, directories |
| Dim | `#777777` (dark) / `#A49FA5` (light) | Line numbers, status bar, inactive borders |
| Error | `#FF6B6B` | Unreadable file message |

Borders: `lipgloss.RoundedBorder()` on both panels.

Active item in list: `▸` prefix, bold, accent color.
Directory entries in list: blue `#569CD6`, no arrow.
Normal file entries: default foreground.

---

## 12. Startup Behavior

On launch, dotdash must:

1. Scan `~` as described in section 6
2. If no entries are found, print `"No dotfiles found in ~"` and exit with code 1
3. Otherwise, open the TUI in alternate screen mode (`tea.WithAltScreen()`)
4. Pre-select the first entry and populate the preview panel immediately
5. Enable mouse support (`tea.WithMouseCellMotion()`) for scrolling the preview

No arguments or flags are required for basic usage. Optionally accept a `--home <path>` flag to override the scan root (useful for testing).

---

## 13. Global Installation & Invocation

A core requirement is that the tool is callable **from any directory** in the terminal, just like `k9s`, `htop`, or `lazygit`. The binary must be placed in a directory on `$PATH`.

### 13.1 Install methods

**Method A — go install (recommended for Go users):**
```bash
go install github.com/<user>/dotdash@latest
# binary lands in $(go env GOPATH)/bin/dotdash
# ensure $(go env GOPATH)/bin is in $PATH
```

**Method B — manual build:**
```bash
git clone https://github.com/<user>/dotdash
cd dotdash
go build -o dotdash .
sudo mv dotdash /usr/local/bin/
```

**Method C — pre-built release binary:**
Download from GitHub Releases, `chmod +x`, move to `/usr/local/bin/` or `~/bin/`.

### 13.2 Invocation

Once installed, the tool is invoked with a single command from any path:

```bash
dotdash
```

The current working directory is **completely ignored**. The tool always scans `~` regardless of where it is called from. This is the same behavior as `k9s` (which always connects to the current kubeconfig context, not a local directory).

### 13.3 PATH note for the README

The README must include a clear note:

> **Make sure the install directory is in your `$PATH`.** For Go installs, add `export PATH=$PATH:$(go env GOPATH)/bin` to your `.zshrc` or `.bashrc` if not already present.

---

## 14. Project Structure

```
dotdash/
├── main.go          # Entry point, tea.Program setup, CLI flags
├── model.go         # Model struct, Init/Update/View
├── scan.go          # scanDotfiles() function
├── preview.go       # File reading, syntax highlighting, directory listing
├── crud.go          # Copy-path to clipboard
├── styles.go        # All lipgloss style definitions
├── keys.go          # Keybinding constants
├── go.mod
├── go.sum
└── README.md
```

Split into multiple files for maintainability. Keep each file under ~200 lines.

---

## 15. README Requirements

The README must include:

- One-paragraph description
- ASCII art screenshot of the layout
- Install instructions (all three methods from section 13.1)
- `$PATH` setup note
- Complete keybindings table (navigation + actions)
- Description of what directories are scanned and what is skipped
- Section on how to customize the skip list

---

## 16. Out of Scope for v1 (Future Backlog)

- `--config` flag to define custom skip lists via YAML/TOML
- Recursive scan beyond 1 level deep inside dot directories
- File content search (grep across all dotfiles)
- Git status indicators per file
- Mouse click to select entries
- File creation from within the TUI
- Split `.config` subdirectories as top-level groups
