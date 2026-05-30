package main

func rebuildVisibleRows(roots []*TreeNode, expanded map[string]bool) []visibleRow {
	return flattenTree(roots, expanded, 0)
}

func flattenTree(nodes []*TreeNode, expanded map[string]bool, depth int) []visibleRow {
	var rows []visibleRow
	for _, node := range nodes {
		exp := expanded[node.Path]
		rows = append(rows, visibleRow{node: node, depth: depth, expanded: exp})
		if node.IsDir && exp {
			rows = append(rows, flattenTree(node.Children, expanded, depth+1)...)
		}
	}
	return rows
}

func findRowByRelPath(rows []visibleRow, relPath string) int {
	for i, row := range rows {
		if row.node.RelPath == relPath {
			return i
		}
	}
	return -1
}
