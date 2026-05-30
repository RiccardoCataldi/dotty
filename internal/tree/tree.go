package tree

import "github.com/riccardo/dotty/internal/scan"

type Row struct {
	Node     *scan.TreeNode
	Depth    int
	Expanded bool
}

func Visible(roots []*scan.TreeNode, expanded map[string]bool) []Row {
	return flatten(roots, expanded, 0)
}

func FindByRelPath(rows []Row, relPath string) int {
	for i, row := range rows {
		if row.Node.RelPath == relPath {
			return i
		}
	}
	return -1
}

func flatten(nodes []*scan.TreeNode, expanded map[string]bool, depth int) []Row {
	var rows []Row
	for _, node := range nodes {
		exp := expanded[node.Path]
		rows = append(rows, Row{Node: node, Depth: depth, Expanded: exp})
		if node.IsDir && exp {
			rows = append(rows, flatten(node.Children, expanded, depth+1)...)
		}
	}
	return rows
}
