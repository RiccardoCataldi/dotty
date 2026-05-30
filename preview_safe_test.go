package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsBinaryDataNUL(t *testing.T) {
	if !isBinaryData([]byte("hello\x00world")) {
		t.Fatal("expected binary")
	}
	if isBinaryData([]byte("hello world")) {
		t.Fatal("expected text")
	}
}

func TestIsBinaryExtension(t *testing.T) {
	if !isBinaryExtension("/path/libusb0.sys") {
		t.Fatal(".sys should be binary")
	}
	if isBinaryExtension("/path/readme.txt") {
		t.Fatal(".txt should not be binary")
	}
}

func TestBuildFilePreviewBinary(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.dll")
	if err := os.WriteFile(path, []byte{0x4d, 0x5a, 0x90, 0x00, 0x01}, 0o644); err != nil {
		t.Fatal(err)
	}
	r := buildFilePreview(Entry{Name: "test.dll", Path: path, IsDir: false})
	if !strings.Contains(r.Content, "Binary file cannot be previewed") {
		t.Fatalf("expected binary message, got %q", r.Content)
	}
}

func TestSanitizePreviewLineStripsControls(t *testing.T) {
	in := "ok\x1b[31mred\x1b[0m\x07bell"
	out := sanitizePreviewLine(in)
	if strings.Contains(out, "\x1b") {
		t.Fatalf("ANSI should be stripped: %q", out)
	}
}
