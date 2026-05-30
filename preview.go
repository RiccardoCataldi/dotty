package main

import (
	"fmt"
	"os"
	"strings"
)

const (
	maxPreviewSize  = 512 * 1024
	maxPreviewLines = 500
)

type previewResult struct {
	Content    string
	Title      string
	TotalLines int
}

func buildPreview(entry Entry) previewResult {
	if entry.IsDir {
		return buildDirPreview(entry)
	}
	return buildFilePreview(entry)
}

func binaryPreviewResult(name string, size int64) previewResult {
	msg := fmt.Sprintf("Binary file cannot be previewed (%s)", formatFileSize(size))
	return previewResult{
		Content:    styleDim.Render(msg),
		Title:      name,
		TotalLines: 1,
	}
}

func formatFileSize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for d := n / unit; d >= unit; d /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}

func buildDirPreview(entry Entry) previewResult {
	return previewResult{
		Content: styleDim.Render("Directory — press Enter to expand, then select a file to preview"),
		Title:   entry.Name,
	}
}

func buildFilePreview(entry Entry) previewResult {
	info, err := os.Stat(entry.Path)
	if err != nil {
		return previewResult{
			Content: styleError.Render("Unable to read file"),
			Title:   entry.Name,
		}
	}

	if info.Size() > maxPreviewSize {
		return previewResult{
			Content: styleDim.Render("File too large to preview (> 512 KB)"),
			Title:   entry.Name,
		}
	}

	if isBinaryExtension(entry.Path) {
		return binaryPreviewResult(entry.Name, info.Size())
	}

	data, err := os.ReadFile(entry.Path)
	if err != nil {
		return previewResult{
			Content: styleError.Render("Unable to read file"),
			Title:   entry.Name,
		}
	}

	if len(data) == 0 {
		return previewResult{
			Content: styleDim.Render("(empty file)"),
			Title:   entry.Name,
		}
	}

	if isBinaryData(data) {
		return binaryPreviewResult(entry.Name, info.Size())
	}

	content := sanitizePreviewContent(string(data))
	if strings.TrimSpace(content) == "" {
		return previewResult{
			Content: styleDim.Render("(empty file)"),
			Title:   entry.Name,
		}
	}

	lines := strings.Split(content, "\n")
	totalLines := len(lines)
	truncated := false
	remaining := 0

	if totalLines > maxPreviewLines {
		remaining = totalLines - maxPreviewLines
		lines = lines[:maxPreviewLines]
		truncated = true
	}

	var rendered []string
	for i, line := range lines {
		lineNum := fmt.Sprintf("%4d", i+1)
		highlighted := highlightLine(line)
		rendered = append(rendered, styleLineNumber.Render(lineNum)+"  "+highlighted)
	}

	if truncated {
		msg := fmt.Sprintf("... (%d more lines not shown)", remaining)
		rendered = append(rendered, styleDim.Render(msg))
	}

	return previewResult{
		Content:   strings.Join(rendered, "\n"),
		Title:     entry.Name,
		TotalLines: totalLines,
	}
}

func highlightLine(line string) string {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "--") {
		return styleComment.Render(line)
	}
	if strings.HasPrefix(trimmed, "[[") || strings.HasPrefix(trimmed, "[") {
		return styleComment.Render(line)
	}
	return line
}

func previewTitleWithPercent(title string, yOffset, visibleLines, totalLines int) string {
	if totalLines <= visibleLines || visibleLines <= 0 {
		return title
	}
	maxScroll := totalLines - visibleLines
	if maxScroll <= 0 {
		return title
	}
	pct := (yOffset * 100) / maxScroll
	if pct > 100 {
		pct = 100
	}
	return fmt.Sprintf("%s  %d%%", title, pct)
}

func renderListItem(row visibleRow, selected bool) string {
	indent := strings.Repeat("  ", row.depth)

	var marker string
	switch {
	case selected:
		marker = "▸"
	case row.node.IsDir:
		if row.expanded {
			marker = "▾"
		} else {
			marker = "▸"
		}
	default:
		marker = " "
	}

	name := row.node.Name
	line := indent + marker + " " + name

	if selected {
		return styleActiveItem.Render(line)
	}
	if row.node.IsDir {
		return styleDirItem.Render(line)
	}
	return line
}

