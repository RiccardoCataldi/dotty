package main

import (
	"path/filepath"
	"strings"
	"unicode"
)

const binaryProbeSize = 8192

// Binary extensions Telescope/fzf typically refuse to cat into the terminal.
var binaryExtensions = map[string]bool{
	".7z": true, ".a": true, ".avi": true, ".bin": true, ".bmp": true,
	".bz2": true, ".class": true, ".dat": true, ".db": true, ".dll": true,
	".dylib": true, ".elf": true, ".eot": true, ".exe": true, ".gif": true,
	".gz": true, ".ico": true, ".jar": true, ".jpeg": true, ".jpg": true,
	".lz": true, ".lz4": true, ".mkv": true, ".mp3": true, ".mp4": true,
	".o": true, ".obj": true, ".otf": true, ".pdb": true, ".pdf": true,
	".png": true, ".so": true, ".sqlite": true, ".sys": true, ".tar": true,
	".ttf": true, ".wasm": true, ".webm": true, ".woff": true, ".woff2": true,
	".xz": true, ".zip": true, ".zst": true,
}

func isBinaryExtension(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return binaryExtensions[ext]
}

// isBinaryData mirrors fzf/telescope: NUL in the first 8KiB means binary.
func isBinaryData(data []byte) bool {
	n := len(data)
	if n > binaryProbeSize {
		n = binaryProbeSize
	}
	for i := 0; i < n; i++ {
		if data[i] == 0 {
			return true
		}
	}
	return false
}

func stripANSI(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] != '\x1b' {
			b.WriteByte(s[i])
			continue
		}
		i++
		if i < len(s) && s[i] == '[' {
			i++
			for i < len(s) && (s[i] < 0x40 || s[i] > 0x7e) {
				i++
			}
		}
	}
	return b.String()
}

func sanitizePreviewLine(line string) string {
	line = stripANSI(line)
	var b strings.Builder
	b.Grow(len(line))
	for _, r := range line {
		switch {
		case r == '\t':
			b.WriteRune(r)
		case r >= 0x20 && r != 0x7f && r != '\u2028' && r != '\u2029':
			if unicode.IsPrint(r) {
				b.WriteRune(r)
			} else {
				b.WriteRune('·')
			}
		default:
			// drop CR, ESC, other controls (Telescope never prints these raw)
		}
	}
	return b.String()
}

func sanitizePreviewContent(content string) string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = sanitizePreviewLine(line)
	}
	return strings.Join(lines, "\n")
}

