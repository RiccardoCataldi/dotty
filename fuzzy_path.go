package main

import (
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

const (
	scoreMatch            = 16
	scoreGapPenalty       = 3
	scoreBoundaryBonus    = 8
	scoreConsecutiveBonus = 6
	scoreBasenameBonus    = 24
	scoreExtensionBoost   = 2000
	minPathFuzzyScore     = 40
	pickerResultCap       = 1000
)

type scoredPickerEntry struct {
	entry   pickerEntry
	score   int
	indices []int
}

type pathFuzzyResult struct {
	score    int
	text     []rune
	prevJ    []int
	queryLen int
	ok       bool
}

func filterAndRankPickerFiles(all []pickerEntry, query string, prev []pickerMatch, prevQuery string) ([]pickerMatch, int) {
	if query == "" {
		out := make([]pickerMatch, 0, min(len(all), pickerResultCap))
		for i, e := range all {
			if i >= pickerResultCap {
				break
			}
			out = append(out, pickerMatch{entry: e})
		}
		return out, len(all)
	}

	source := all
	if prevQuery != "" && strings.HasPrefix(query, prevQuery) && len(query) > len(prevQuery) && len(prev) > 0 {
		source = make([]pickerEntry, len(prev))
		for i, m := range prev {
			source[i] = m.entry
		}
	}

	qLower := []rune(strings.ToLower(query))
	scored := make([]scoredPickerEntry, 0, 256)
	for _, e := range source {
		if !quickSubsequence(qLower, e.relLower) {
			continue
		}
		s, indices, ok := scorePathFuzzyWithIndices(qLower, e)
		if !ok || s < minPathFuzzyScore {
			continue
		}
		scored = append(scored, scoredPickerEntry{entry: e, score: s, indices: indices})
	}

	matchCount := len(scored)

	sort.Slice(scored, func(i, j int) bool {
		if scored[i].score != scored[j].score {
			return scored[i].score > scored[j].score
		}
		li, lj := len(scored[i].entry.RelPath), len(scored[j].entry.RelPath)
		if li != lj {
			return li < lj
		}
		return scored[i].entry.RelPath < scored[j].entry.RelPath
	})

	if len(scored) > pickerResultCap {
		scored = scored[:pickerResultCap]
	}

	out := make([]pickerMatch, len(scored))
	for i, s := range scored {
		out[i] = pickerMatch{entry: s.entry, indices: s.indices}
	}
	return out, matchCount
}

func quickSubsequence(q, text []rune) bool {
	if len(q) == 0 {
		return true
	}
	qi := 0
	for _, r := range text {
		if r == q[qi] {
			qi++
			if qi == len(q) {
				return true
			}
		}
	}
	return false
}

type fuzzyDPBuf struct {
	dp    []int
	prevJ []int
}

var fuzzyDPBufPool = sync.Pool{
	New: func() any {
		return &fuzzyDPBuf{
			dp:    make([]int, 0, 32),
			prevJ: make([]int, 0, 32),
		}
	},
}

func acquireFuzzyDP(m int) (*fuzzyDPBuf, []int, []int) {
	buf := fuzzyDPBufPool.Get().(*fuzzyDPBuf)
	need := m + 1
	if cap(buf.dp) < need {
		buf.dp = make([]int, need)
		buf.prevJ = make([]int, need)
	} else {
		buf.dp = buf.dp[:need]
		buf.prevJ = buf.prevJ[:need]
	}
	return buf, buf.dp, buf.prevJ
}

func computePathFuzzy(qLower []rune, path string, tLower []rune) pathFuzzyResult {
	if len(qLower) == 0 {
		return pathFuzzyResult{ok: true}
	}

	n := len(tLower)
	m := len(qLower)

	baseLower := []rune(strings.ToLower(filepath.Base(path)))
	baseStart := n - len(baseLower)

	const negInf = -1 << 30
	buf, dp, prevJ := acquireFuzzyDP(m)
	defer fuzzyDPBufPool.Put(buf)
	for i := range dp {
		dp[i] = negInf
		prevJ[i] = -1
	}
	dp[0] = 0

	for j := 0; j < n; j++ {
		tr := tLower[j]
		for i := m; i >= 1; i-- {
			if qLower[i-1] != tr || dp[i-1] <= negInf/2 {
				continue
			}

			points := scoreMatch
			if j > 0 {
				prev := tLower[j-1]
				if prev == '/' || prev == '.' || prev == '_' || prev == '-' {
					points += scoreBoundaryBonus
				}
			}
			if j >= baseStart {
				points += scoreBasenameBonus
			}
			if prevJ[i-1] >= 0 && j == prevJ[i-1]+1 {
				points += scoreConsecutiveBonus
			} else if prevJ[i-1] >= 0 {
				gap := j - prevJ[i-1] - 1
				points -= scoreGapPenalty * gap
			}

			newScore := dp[i-1] + points
			if newScore > dp[i] {
				dp[i] = newScore
				prevJ[i] = j
			}
		}
	}

	if dp[m] <= negInf/2 {
		return pathFuzzyResult{}
	}

	return pathFuzzyResult{
		score:    dp[m] + extensionBoostFromLower(qLower, path),
		text:     tLower,
		prevJ:    prevJ,
		queryLen: m,
		ok:       true,
	}
}

func scorePathFuzzyWithIndices(qLower []rune, e pickerEntry) (int, []int, bool) {
	r := computePathFuzzy(qLower, e.RelPath, e.relLower)
	if !r.ok {
		return 0, nil, false
	}
	if r.queryLen == 0 {
		return 0, nil, true
	}
	return r.score, backtrackByteIndices(qLower, r.text, r.prevJ, r.queryLen), true
}

func scorePathFuzzy(query, path string) (int, bool) {
	qLower := []rune(strings.ToLower(query))
	e := pickerEntry{RelPath: path, relLower: []rune(strings.ToLower(path))}
	s, _, ok := scorePathFuzzyWithIndices(qLower, e)
	return s, ok
}

func extensionBoostFromLower(qLower []rune, path string) int {
	ql := string(qLower)
	base := strings.ToLower(filepath.Base(path))
	if strings.HasSuffix(base, "."+ql) {
		return scoreExtensionBoost
	}
	if strings.HasSuffix(base, ql) {
		return scoreExtensionBoost / 2
	}
	return 0
}

func extensionBoost(query, path string) int {
	return extensionBoostFromLower([]rune(strings.ToLower(query)), path)
}

// pathFuzzyIndices returns byte indices in path for highlighting (best-scoring alignment).
func pathFuzzyIndices(query, path string) []int {
	qLower := []rune(strings.ToLower(query))
	e := pickerEntry{RelPath: path, relLower: []rune(strings.ToLower(path))}
	_, indices, ok := scorePathFuzzyWithIndices(qLower, e)
	if !ok {
		return nil
	}
	return indices
}

func backtrackByteIndices(_ []rune, tLower []rune, prevJ []int, m int) []int {
	var runeIdx []int
	j := prevJ[m]
	for i := m; i >= 1; i-- {
		runeIdx = append([]int{j}, runeIdx...)
		if i > 1 {
			j = prevJ[i-1]
		}
	}
	return runeIndicesToByteIndices(tLower, runeIdx)
}

func runeIndicesToByteIndices(text []rune, runeIdx []int) []int {
	set := make(map[int]bool, len(runeIdx))
	for _, ri := range runeIdx {
		set[ri] = true
	}
	var bytes []int
	off := 0
	for ri, r := range text {
		rl := utf8.RuneLen(r)
		if set[ri] {
			for b := 0; b < rl; b++ {
				bytes = append(bytes, off+b)
			}
		}
		off += rl
	}
	return bytes
}

func pickerDisplayPath(relPath string, maxWidth int) string {
	if maxWidth <= 0 {
		return relPath
	}
	if runewidth.StringWidth(relPath) <= maxWidth {
		return relPath
	}
	const ellipsis = "…"
	avail := maxWidth - runewidth.StringWidth(ellipsis)
	if avail < 4 {
		return tailByWidth(relPath, maxWidth)
	}
	return ellipsis + tailByWidth(relPath, avail)
}

func mapIndicesToDisplay(fullPath, display string, fullIndices []int) []int {
	if len(fullIndices) == 0 || display == fullPath {
		return fullIndices
	}
	ellipsis := ""
	trimmed := display
	if strings.HasPrefix(display, "…") {
		ellipsis = "…"
		trimmed = display[len(ellipsis):]
	}
	start := strings.Index(fullPath, trimmed)
	if start < 0 {
		return nil
	}
	ellLen := len(ellipsis)
	var out []int
	for _, bi := range fullIndices {
		if bi < start || bi >= start+len(trimmed) {
			continue
		}
		out = append(out, ellLen+(bi-start))
	}
	return out
}

func tailByWidth(s string, maxWidth int) string {
	if runewidth.StringWidth(s) <= maxWidth {
		return s
	}
	runes := []rune(s)
	for len(runes) > 0 && runewidth.StringWidth(string(runes)) > maxWidth {
		runes = runes[1:]
	}
	return string(runes)
}

func pickerHighlightIndices(query, relPath, display string) []int {
	full := pathFuzzyIndices(query, relPath)
	if display == relPath {
		return full
	}
	return mapIndicesToDisplay(relPath, display, full)
}
