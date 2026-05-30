package main

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type focusPanel int

const (
	focusList focusPanel = iota
	focusPreview
)

type uiMode int

const (
	modeNormal uiMode = iota
	modePicker
)

type tickMsg time.Time

type pickerWarmMsg struct {
	files []pickerEntry
}

type pickerFilterMsg struct {
	gen     int
	query   string
	results []pickerMatch
	total   int
}

type model struct {
	homeDir           string
	roots             []*TreeNode
	totalCount        int
	visibleRows       []visibleRow
	expanded          map[string]bool
	cursor            int
	listOffset        int
	focus             focusPanel
	mode              uiMode
	pickerInput            textinput.Model
	pickerAll              []pickerEntry
	pickerResults          []pickerMatch
	pickerMatchTotal       int
	pickerLastQuery        string
	pickerFilterGen        int
	pickerCursor           int
	pickerOffset           int
	pickerPreview          viewport.Model
	pickerPreviewPath      string
	pickerPreviewTitle     string
	pickerPreviewTotalLines int
	preview           viewport.Model
	previewTitle      string
	previewTotalLines int
	width, height     int
	panelHeight       int
	listWidth         int
	previewWidth      int
	listInnerHeight   int
	statusMsg         string
	statusMsgExpiry   time.Time
}

func newModel(homeDir string, roots []*TreeNode, totalCount int) model {
	pi := textinput.New()
	pi.Prompt = "> "
	pi.CharLimit = 256
	pi.Width = 60
	ppv := viewport.New(0, 0)
	vp := viewport.New(0, 0)

	m := model{
		homeDir:       homeDir,
		roots:         roots,
		totalCount:    totalCount,
		expanded:      make(map[string]bool),
		pickerInput:   pi,
		pickerPreview: ppv,
		preview:       vp,
		focus:       focusList,
		mode:        modeNormal,
	}
	m.rebuildVisible()
	m.refreshPreview()
	return m
}

func (m *model) rebuildVisible() {
	m.visibleRows = rebuildVisibleRows(m.roots, m.expanded)
	if m.cursor >= len(m.visibleRows) && len(m.visibleRows) > 0 {
		m.cursor = len(m.visibleRows) - 1
	}
	if len(m.visibleRows) == 0 {
		m.cursor = 0
	} else if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m *model) refreshPreview() {
	if len(m.visibleRows) == 0 {
		m.preview.SetContent("")
		m.previewTitle = ""
		return
	}
	if m.cursor >= len(m.visibleRows) {
		m.cursor = len(m.visibleRows) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	entry := m.visibleRows[m.cursor].node.entry()
	result := buildPreview(entry)
	m.preview.SetContent(result.Content)
	m.previewTitle = result.Title
	m.previewTotalLines = result.TotalLines
	m.preview.GotoTop()
}

func (m *model) selectedEntry() (Entry, bool) {
	if len(m.visibleRows) == 0 || m.cursor >= len(m.visibleRows) {
		return Entry{}, false
	}
	return m.visibleRows[m.cursor].node.entry(), true
}

func (m *model) toggleExpandAtCursor() {
	if m.cursor >= len(m.visibleRows) {
		return
	}
	row := m.visibleRows[m.cursor]
	if !row.node.IsDir {
		return
	}
	if m.expanded[row.node.Path] {
		delete(m.expanded, row.node.Path)
	} else {
		_ = loadChildren(row.node)
		m.expanded[row.node.Path] = true
		m.totalCount = countTreeNodes(m.roots)
		m.invalidatePickerCache()
	}
	m.rebuildVisible()
	m.ensureCursorVisible()
}

func (m *model) collapseAtCursor() {
	if m.cursor >= len(m.visibleRows) {
		return
	}
	row := m.visibleRows[m.cursor]
	if row.depth > 0 {
		for i := m.cursor - 1; i >= 0; i-- {
			if m.visibleRows[i].depth < row.depth {
				m.cursor = i
				delete(m.expanded, m.visibleRows[i].node.Path)
				m.rebuildVisible()
				m.ensureCursorVisible()
				m.refreshPreview()
				return
			}
		}
	}
	if row.node.IsDir && m.expanded[row.node.Path] {
		delete(m.expanded, row.node.Path)
		m.rebuildVisible()
		m.ensureCursorVisible()
	}
}

func (m *model) ensureCursorVisible() {
	if m.cursor < m.listOffset {
		m.listOffset = m.cursor
	}
	if m.cursor >= m.listOffset+m.listInnerHeight {
		m.listOffset = m.cursor - m.listInnerHeight + 1
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) }),
		func() tea.Msg {
			roots := m.roots
			return pickerWarmMsg{files: collectAllFiles(roots)}
		},
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layoutPanels()
		return m, nil

	case tickMsg:
		if m.statusMsg != "" && time.Now().After(m.statusMsgExpiry) {
			m.statusMsg = ""
		}
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })

	case pickerWarmMsg:
		if m.pickerAll == nil {
			m.pickerAll = msg.files
		}
		return m, nil

	case pickerFilterMsg:
		if msg.gen != m.pickerFilterGen {
			return m, nil
		}
		m.applyPickerFilter(msg.query, msg.results, msg.total)
		return m, nil

	case tea.KeyMsg:
		if m.mode == modePicker {
			return m.updatePicker(msg)
		}
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		if cmd := m.handleKey(msg); cmd != nil {
			return m, cmd
		}
		return m, nil
	}

	if m.focus == focusPreview && m.mode == modeNormal {
		var cmd tea.Cmd
		m.preview, cmd = m.preview.Update(msg)
		return m, cmd
	}

	return m, nil
}
