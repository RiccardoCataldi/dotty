package app

import (
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/riccardo/dotty/internal/clipboard"
)

func (m *Model) handleKey(msg tea.KeyMsg) tea.Cmd {
	switch {
	case msg.String() == "esc":
		return nil

	case key.Matches(msg, keys.Search):
		return m.openPicker()

	case key.Matches(msg, keys.Tab):
		if m.focus == focusList {
			m.focus = focusPreview
		} else {
			m.focus = focusList
		}
		return nil
	}

	if m.focus == focusPreview {
		switch {
		case key.Matches(msg, keys.ScrollDown):
			m.preview.LineDown(1)
		case key.Matches(msg, keys.ScrollUp):
			m.preview.LineUp(1)
		case key.Matches(msg, keys.HalfPageDown):
			m.preview.HalfViewDown()
		case key.Matches(msg, keys.HalfPageUp):
			m.preview.HalfViewUp()
		}
		return nil
	}

	entry, ok := m.selectedEntry()

	switch {
	case key.Matches(msg, keys.Down):
		if m.cursor < len(m.visibleRows)-1 {
			m.cursor++
			m.ensureCursorVisible()
			m.refreshPreview()
		}
	case key.Matches(msg, keys.Up):
		if m.cursor > 0 {
			m.cursor--
			m.ensureCursorVisible()
			m.refreshPreview()
		}
	case key.Matches(msg, keys.First):
		if len(m.visibleRows) > 0 {
			m.cursor = 0
			m.listOffset = 0
			m.refreshPreview()
		}
	case key.Matches(msg, keys.Last):
		if len(m.visibleRows) > 0 {
			m.cursor = len(m.visibleRows) - 1
			m.ensureCursorVisible()
			m.refreshPreview()
		}
	case key.Matches(msg, keys.Expand):
		m.toggleExpandAtCursor()
		m.refreshPreview()
	case key.Matches(msg, keys.Collapse):
		m.collapseAtCursor()
		m.refreshPreview()
	case key.Matches(msg, keys.CopyPath):
		if ok && !entry.IsDir {
			msgText, err := clipboard.CopyPath(entry.Path)
			if err != nil {
				m.statusMsg = err.Error()
			} else {
				m.statusMsg = msgText
			}
			m.statusMsgExpiry = time.Now().Add(3 * time.Second)
		}
	}
	return nil
}
