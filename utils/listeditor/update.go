package listeditor

import (
	"slices"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/bank_data_tui/utils"
	"github.com/bank_data_tui/utils/editor"
)

type ItemNew struct{ Value any }
type ItemUpdate struct{ Value any }

func (m *Model[T, PT]) Update(msg tea.Msg) (utils.Screen, tea.Cmd) {
	batcher := []tea.Cmd{}
	var cmd tea.Cmd
	bubble := true

	switch msg := msg.(type) {
	case initialResp[PT]:
		m.items = msg
		cmd = m.list.SetItems(m.categoryItems())
		batcher = append(batcher, cmd)
		m.isLoaded = true
	case editor.ItemNew:
		m.curItem.SetID(string(msg))
		m.items = append(m.items, m.curItem)
		batcher = append(batcher, m.list.SetItems(m.categoryItems()))
		m.list.GoToEnd()
		batcher = append(batcher, func() tea.Msg {
			return ItemNew{Value: m.curItem}
		})
	case editor.ItemDel:
		i := slices.IndexFunc(m.items, func(c PT) bool { return c.GetID() == string(msg) })
		if i != -1 {
			m.items = slices.Delete(m.items, i, i+1)
		}
		batcher = append(batcher, m.list.SetItems(m.categoryItems()))
	case editor.ItemUpdate:
		batcher = append(batcher, func() tea.Msg {
			return ItemUpdate{Value: m.curItem}
		})
	case tea.KeyPressMsg:
		switch msg.String() {
		case "alt+up":
			bubble = false
			m.list.CursorUp()
		case "alt+down":
			bubble = false
			m.list.CursorDown()
		}
	case utils.ResizeMessage:
		m.w, m.h = msg.W, msg.H

		m.list.SetHeight(msg.H)
		m.editor.SetWidth(msg.W - WIDTH_OFFSET_EDITOR)
	case tea.MouseWheelMsg:
		bubble = false

		switch msg.Button {
		case tea.MouseWheelUp:
			m.list.CursorUp()
		case tea.MouseWheelDown:
			m.list.CursorDown()
		}
	}

	if !m.isLoaded {
		m.spin, cmd = m.spin.Update(msg)
		batcher = append(batcher, cmd)
	}

	if a, ok := m.Abstraction.(interface{ Update(msg tea.Msg) }); ok {
		a.Update(msg)
	}

	if bubble {
		forList := false
		if km, ok := msg.(tea.KeyPressMsg); ok {
			forList = doesKeyMatchList(km, m.list)
		}

		m.list, cmd = m.list.Update(msg)
		batcher = append(batcher, cmd)

		if !forList && m.list.FilterState() != list.Filtering {
			m.editor, cmd = m.editor.Update(msg)
			batcher = append(batcher, cmd)
		}
	}

	i := m.list.GlobalIndex()
	if m.isNewCategory(i) {
		if m.curItem.GetID() != "" {
			m.curItem = new(T)
			m.resetEditor()
			batcher = append(batcher, m.editor.Init())
		}
	} else if m.items[i-1].GetID() != m.curItem.GetID() {
		m.curItem = m.items[i-1]
		m.resetEditor()
		batcher = append(batcher, m.editor.Init())
	}

	return m, tea.Batch(batcher...)
}

func (m *Model[T, PT]) isNewCategory(gi int) bool {
	return gi == 0
}

func (m *Model[T, PT]) categoryItems() []list.Item {
	arr := make([]list.Item, len(m.items)+1)
	arr[0] = m.newItem
	for i, v := range m.items {
		arr[i+1] = v
	}

	return arr
}

func (m *Model[item, PT]) resetEditor() {
	m.editor = m.NewEditor(m.w, m.h, m.curItem)
}
