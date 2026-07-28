package scan

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

var loadMu sync.Mutex

// DefaultBlocklist skips noisy or sensitive dot-directories under $HOME.
var DefaultBlocklist = map[string]bool{
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

func (n *TreeNode) ToEntry() Entry {
	return Entry{Name: n.RelPath, Path: n.Path, IsDir: n.IsDir}
}

func Dotfiles(homeDir string) ([]*TreeNode, int, error) {
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
			if DefaultBlocklist[name] {
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
	return roots, CountNodes(roots), nil
}

func LoadChildren(node *TreeNode) error {
	if node == nil || !node.IsDir {
		return nil
	}

	loadMu.Lock()
	defer loadMu.Unlock()
	if node.loaded {
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
		a, b := nodes[i], nodes[j]
		if a == nil {
			return false
		}
		if b == nil {
			return true
		}
		return a.RelPath < b.RelPath
	})
}

func CountNodes(nodes []*TreeNode) int {
	n := 0
	for _, node := range nodes {
		n++
		if node.loaded {
			n += CountNodes(node.Children)
		}
	}
	return n
}
