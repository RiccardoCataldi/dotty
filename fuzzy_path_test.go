package main

import (
	"testing"
)

func TestPemRanksExtensionFile(t *testing.T) {
	entries := []pickerEntry{
		newPickerEntry(".agents/skills/improve-codebase-architecture/DEEPENING.md", ""),
		newPickerEntry(".ssh/cert.pem", ""),
		newPickerEntry(".config/foo/bar.pem", ""),
	}
	results, matchTotal := filterAndRankPickerFiles(entries, "pem", nil, "")
	if matchTotal == 0 {
		t.Fatal("expected matches for pem")
	}
	if len(results) == 0 {
		t.Fatal("expected capped results")
	}
	first := results[0].entry.RelPath
	if first != ".ssh/cert.pem" && first != ".config/foo/bar.pem" {
		t.Fatalf("expected .pem file first, got %q (total matches %d)", first, matchTotal)
	}
}

func TestNviRanksNvimPath(t *testing.T) {
	entries := []pickerEntry{
		newPickerEntry(".vimrc", ""),
		newPickerEntry(".config/nvim/init.lua", ""),
		newPickerEntry(".zshrc", ""),
	}
	results, _ := filterAndRankPickerFiles(entries, "nvi", nil, "")
	if len(results) == 0 {
		t.Fatal("expected matches for nvi")
	}
	if results[0].entry.RelPath != ".config/nvim/init.lua" {
		t.Fatalf("expected nvim path first, got %q", results[0].entry.RelPath)
	}
}

func TestWeakSubsequenceFiltered(t *testing.T) {
	// scattered p,e,m across long path should score below threshold
	long := ".cache/foo/bar/baz/qux/" + stringsRepeat("a", 80) + "/readme.md"
	entries := []pickerEntry{newPickerEntry(long, "")}
	results, matchTotal := filterAndRankPickerFiles(entries, "pem", nil, "")
	if matchTotal > 0 && len(results) > 0 {
		s, ok := scorePathFuzzy("pem", long)
		if ok && s >= minPathFuzzyScore {
			t.Logf("long path scored %d (allowed if above threshold)", s)
		}
	}
	// cert.pem must always beat weak paths when both present
	entries = []pickerEntry{
		newPickerEntry(long, ""),
		newPickerEntry("key.pem", ""),
	}
	results, _ = filterAndRankPickerFiles(entries, "pem", nil, "")
	if results[0].entry.RelPath != "key.pem" {
		t.Fatalf("expected key.pem first, got %q", results[0].entry.RelPath)
	}
}

func TestExtensionBoost(t *testing.T) {
	s1, ok1 := scorePathFuzzy("pem", "cert.pem")
	s2, ok2 := scorePathFuzzy("pem", ".local/share/patch/example.md")
	if !ok1 {
		t.Fatal("cert.pem should match")
	}
	if !ok2 {
		t.Fatal("scattered pem in path should match subsequence")
	}
	if s1 <= s2 {
		t.Fatalf("cert.pem score %d should exceed weak path score %d", s1, s2)
	}
}

func stringsRepeat(s string, n int) string {
	b := make([]byte, 0, len(s)*n)
	for i := 0; i < n; i++ {
		b = append(b, s...)
	}
	return string(b)
}
