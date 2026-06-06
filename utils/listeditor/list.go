package listeditor

import (
	"io"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bank-data-db/tui/styles"
)

type itemProxy[T Item] struct {
	fv string
	v  T
}

func (ip itemProxy[T]) FilterValue() string {
	return ip.fv
}

func (m *Model[T]) categoryItems() []list.Item {
	arr := make([]list.Item, len(m.items)+1)
	arr[0] = m.newItem
	for i, v := range m.items {
		arr[i+1] = itemProxy[T]{fv: m.del.FilterValue(v), v: v}
	}

	return arr
}

type CreateNewItemKey string

func (ni CreateNewItemKey) FilterValue() string { return string(ni) }

type itemDel[T Item] func(baseStyle lipgloss.Style, selected bool, item T) string

func (i itemDel[T]) Height() int                               { return 1 }
func (i itemDel[T]) Spacing() int                              { return 1 }
func (i itemDel[T]) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (r itemDel[T]) Render(w io.Writer, m list.Model, index int, item list.Item) {
	style := lipgloss.NewStyle().Foreground(styles.COLOR_MAIN)
	if m.Index() == index {
		style = style.Underline(true).Foreground(styles.COLOR_SECONDARY)
	}

	if i, ok := item.(CreateNewItemKey); ok {
		w.Write(
			[]byte(" " + style.Render("| "+string(i))),
		)
		return
	}

	str := r(style, m.GlobalIndex() == index, item.(itemProxy[T]).v)
	w.Write([]byte(str))
}
