package main

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	First    key.Binding
	Last     key.Binding
	Search   key.Binding
	Expand   key.Binding
	Collapse key.Binding
	CopyPath key.Binding
	Tab      key.Binding
	Quit     key.Binding
	ScrollUp key.Binding
	ScrollDown key.Binding
	HalfPageDown key.Binding
	HalfPageUp key.Binding
}

var keys = keyMap{
	Up:       key.NewBinding(key.WithKeys("k", "up")),
	Down:     key.NewBinding(key.WithKeys("j", "down")),
	First:    key.NewBinding(key.WithKeys("g")),
	Last:     key.NewBinding(key.WithKeys("G")),
	Search:   key.NewBinding(key.WithKeys("/")),
	Expand:   key.NewBinding(key.WithKeys("l", "right", "enter")),
	Collapse: key.NewBinding(key.WithKeys("h", "left")),
	CopyPath: key.NewBinding(key.WithKeys("y")),
	Tab:      key.NewBinding(key.WithKeys("tab")),
	Quit:     key.NewBinding(key.WithKeys("q", "ctrl+c")),
	ScrollUp:       key.NewBinding(key.WithKeys("k", "up")),
	ScrollDown:     key.NewBinding(key.WithKeys("j", "down")),
	HalfPageDown:   key.NewBinding(key.WithKeys("d")),
	HalfPageUp:     key.NewBinding(key.WithKeys("u")),
}
