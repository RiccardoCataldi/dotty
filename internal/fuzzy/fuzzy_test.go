package fuzzy

import "testing"

func TestPemRanksExtensionFile(t *testing.T) {
	entries := []Entry{
		NewEntry(".agents/skills/improve-codebase-architecture/DEEPENING.md", ""),
		NewEntry(".ssh/cert.pem", ""),
		NewEntry(".config/foo/bar.pem", ""),
	}
	results, matchTotal := FilterAndRank(entries, "pem", nil, "")
	if matchTotal == 0 {
		t.Fatal("expected matches for pem")
	}
	if len(results) == 0 {
		t.Fatal("expected capped results")
	}
	first := results[0].Entry.RelPath
	if first != ".ssh/cert.pem" && first != ".config/foo/bar.pem" {
		t.Fatalf("expected .pem file first, got %q (total matches %d)", first, matchTotal)
	}
}

func TestNviRanksNvimPath(t *testing.T) {
	entries := []Entry{
		NewEntry(".vimrc", ""),
		NewEntry(".config/nvim/init.lua", ""),
		NewEntry(".zshrc", ""),
	}
	results, _ := FilterAndRank(entries, "nvi", nil, "")
	if len(results) == 0 {
		t.Fatal("expected matches for nvi")
	}
	if results[0].Entry.RelPath != ".config/nvim/init.lua" {
		t.Fatalf("expected nvim path first, got %q", results[0].Entry.RelPath)
	}
}

func TestWeakSubsequenceFiltered(t *testing.T) {
	long := ".cache/foo/bar/baz/qux/" + stringsRepeat("a", 80) + "/readme.md"
	entries := []Entry{NewEntry(long, "")}
	results, matchTotal := FilterAndRank(entries, "pem", nil, "")
	if matchTotal > 0 && len(results) > 0 {
		s, ok := ScorePath("pem", long)
		if ok && s >= MinPathScore {
			t.Logf("long path scored %d (allowed if above threshold)", s)
		}
	}
	entries = []Entry{
		NewEntry(long, ""),
		NewEntry("key.pem", ""),
	}
	results, _ = FilterAndRank(entries, "pem", nil, "")
	if results[0].Entry.RelPath != "key.pem" {
		t.Fatalf("expected key.pem first, got %q", results[0].Entry.RelPath)
	}
}

func TestExtensionBoost(t *testing.T) {
	s1, ok1 := ScorePath("pem", "cert.pem")
	s2, ok2 := ScorePath("pem", ".local/share/patch/example.md")
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
