package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var defaultBlocklist = map[string]bool{
	".cache":   true,
	".local":   true,
	".mozilla": true,
	".dbus":    true,
	".pki":     true,
	".gnupg":   true,
	".var":     true,
	".snap":    true,
}

type Entry struct {
	Name  string
	Path  string
	IsDir bool
}

type TreeNode struct {
	Name     string
	RelPath  string
	Path     string
	IsDir    bool
	Children []*TreeNode
	loaded   bool
}

type visibleRow struct {
	node     *TreeNode
	depth    int
	expanded bool
}

func (n *TreeNode) entry() Entry {
	return Entry{Name: n.RelPath, Path: n.Path, IsDir: n.IsDir}
}

func scanDotfiles(homeDir string) ([]*TreeNode, int, error) {
	entries, err := os.ReadDir(homeDir)
	if err != nil {
		return nil, 0, err
	}

	var roots []*TreeNode
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, ".") || name == "." || name == ".." {
			continue
		}

		fullPath := filepath.Join(homeDir, name)
		info, err := e.Info()
		if err != nil {
			continue
		}

		if info.IsDir() {
			if defaultBlocklist[name] {
				continue
			}
			roots = append(roots, &TreeNode{
				Name:    name + "/",
				RelPath: name + "/",
				Path:    fullPath,
				IsDir:   true,
			})
		} else {
			roots = append(roots, &TreeNode{
				Name:    name,
				RelPath: name,
				Path:    fullPath,
				IsDir:   false,
			})
		}
	}

	sortTreeNodes(roots)
	return roots, countTreeNodes(roots), nil
}

func loadChildren(node *TreeNode) error {
	if !node.IsDir || node.loaded {
		return nil
	}

	entries, err := os.ReadDir(node.Path)
	if err != nil {
		return err
	}

	for _, e := range entries {
		childName := e.Name()
		childPath := filepath.Join(node.Path, childName)
		info, err := e.Info()
		if err != nil {
			continue
		}

		child := &TreeNode{
			Name:    childName,
			RelPath: node.RelPath + childName,
			Path:    childPath,
			IsDir:   info.IsDir(),
		}
		if info.IsDir() {
			child.Name += "/"
			child.RelPath += "/"
		}
		node.Children = append(node.Children, child)
	}

	sortTreeNodes(node.Children)
	node.loaded = true
	return nil
}

func sortTreeNodes(nodes []*TreeNode) {
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].RelPath < nodes[j].RelPath
	})
}

func countTreeNodes(nodes []*TreeNode) int {
	n := 0
	for _, node := range nodes {
		n++
		if node.loaded {
			n += countTreeNodes(node.Children)
		}
	}
	return n
}
