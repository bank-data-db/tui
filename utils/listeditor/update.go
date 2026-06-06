package listeditor

import (
	"slices"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/bank-data-db/tui/utils"
	"github.com/bank-data-db/tui/utils/editor"
)

type ItemNew struct{ Value any }
type ItemUpdate struct{ Value any }

func (m *Model[T]) Update(msg tea.Msg) (utils.Screen, tea.Cmd) {
	batcher := []tea.Cmd{}
	var cmd tea.Cmd
	bubble := true

	switch msg := msg.(type) {
	case initialResp[T]:
		m.items = msg
		cmd = m.list.SetItems(m.categoryItems())
		batcher = append(batcher, cmd)
		m.isLoaded = true
	case editor.MsgItemNew:
		m.curItem.SetID(string(msg))
		m.items = append(m.items, m.curItem)
		batcher = append(batcher, m.list.SetItems(m.categoryItems()))
		m.list.GoToEnd()
		batcher = append(batcher, func() tea.Msg {
			return ItemNew{Value: m.curItem}
		})
	case editor.MsgItemDel:
		i := slices.IndexFunc(m.items, func(c T) bool { return c.GetID() == string(msg) })
		if i != -1 {
			m.items = slices.Delete(m.items, i, i+1)
		}
		batcher = append(batcher, m.list.SetItems(m.categoryItems()))
	case editor.MsgItemUpdate:
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
		m.editor.Resize(msg.W-WIDTH_OFFSET_EDITOR, msg.H)
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

	if a, ok := m.del.(interface{ Update(msg tea.Msg) tea.Cmd }); ok {
		batcher = append(batcher, a.Update(msg))
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
			m.curItem = m.del.NewItem()
			m.resetEditor()
		}
	} else if m.items[i-1].GetID() != m.curItem.GetID() {
		m.curItem = m.items[i-1]
		m.resetEditor()
	}

	return m, tea.Batch(batcher...)
}

func (m *Model[T]) isNewCategory(gi int) bool {
	return gi == 0
}

func (m *Model[T]) resetEditor() {
	m.editor = m.del.NewEditor(m.w-WIDTH_OFFSET_EDITOR, m.h, m.curItem)
}
