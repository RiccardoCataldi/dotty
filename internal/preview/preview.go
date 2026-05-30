package preview

import (
	"fmt"
	"os"
	"strings"

	"github.com/riccardo/dotty/internal/scan"
	"github.com/riccardo/dotty/internal/tree"
	"github.com/riccardo/dotty/internal/ui"
)

const (
	maxPreviewSize  = 512 * 1024
	maxPreviewLines = 500
)

type Result struct {
	Content    string
	Title      string
	TotalLines int
}

func Build(entry scan.Entry) Result {
	if entry.IsDir {
		return buildDir(entry)
	}
	return buildFile(entry)
}

func binaryResult(name string, size int64) Result {
	msg := fmt.Sprintf("Binary file cannot be previewed (%s)", formatFileSize(size))
	return Result{
		Content:    ui.Dim.Render(msg),
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

func buildDir(entry scan.Entry) Result {
	return Result{
		Content: ui.Dim.Render("Directory — press Enter to expand, then select a file to preview"),
		Title:   entry.Name,
	}
}

func buildFile(entry scan.Entry) Result {
	info, err := os.Stat(entry.Path)
	if err != nil {
		return Result{
			Content: ui.Error.Render("Unable to read file"),
			Title:   entry.Name,
		}
	}

	if info.Size() > maxPreviewSize {
		return Result{
			Content: ui.Dim.Render("File too large to preview (> 512 KB)"),
			Title:   entry.Name,
		}
	}

	if IsBinaryExtension(entry.Path) {
		return binaryResult(entry.Name, info.Size())
	}

	data, err := os.ReadFile(entry.Path)
	if err != nil {
		return Result{
			Content: ui.Error.Render("Unable to read file"),
			Title:   entry.Name,
		}
	}

	if len(data) == 0 {
		return Result{
			Content: ui.Dim.Render("(empty file)"),
			Title:   entry.Name,
		}
	}

	if IsBinaryData(data) {
		return binaryResult(entry.Name, info.Size())
	}

	content := sanitizeContent(string(data))
	if strings.TrimSpace(content) == "" {
		return Result{
			Content: ui.Dim.Render("(empty file)"),
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
		rendered = append(rendered, ui.LineNumber.Render(lineNum)+"  "+highlighted)
	}

	if truncated {
		msg := fmt.Sprintf("... (%d more lines not shown)", remaining)
		rendered = append(rendered, ui.Dim.Render(msg))
	}

	return Result{
		Content:    strings.Join(rendered, "\n"),
		Title:      entry.Name,
		TotalLines: totalLines,
	}
}

func highlightLine(line string) string {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "--") {
		return ui.Comment.Render(line)
	}
	if strings.HasPrefix(trimmed, "[[") || strings.HasPrefix(trimmed, "[") {
		return ui.Comment.Render(line)
	}
	return line
}

func TitleWithPercent(title string, yOffset, visibleLines, totalLines int) string {
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

func RenderListItem(row tree.Row, selected bool) string {
	indent := strings.Repeat("  ", row.Depth)

	var marker string
	switch {
	case selected:
		marker = "▸"
	case row.Node.IsDir:
		if row.Expanded {
			marker = "▾"
		} else {
			marker = "▸"
		}
	default:
		marker = " "
	}

	name := row.Node.Name
	line := indent + marker + " " + name

	if selected {
		return ui.ActiveItem.Render(line)
	}
	if row.Node.IsDir {
		return ui.DirItem.Render(line)
	}
	return line
}
