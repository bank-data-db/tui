package listeditor

import (
	"log"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bank-data-db/tui/utils/editor"
)

const (
	WIDTH_LIST               = 20
	WIDTH_EDITOR_SPLIT_SPACE = 2
	WIDTH_OFFSET_EDITOR      = WIDTH_LIST + 1 + WIDTH_EDITOR_SPLIT_SPACE + WIDTH_EDITOR_SPLIT_SPACE // border + margin + padding
)

type Delegate[T any] interface {
	NewEditor(w, h int, v T) editor.Model
	RenderItem(baseStyle lipgloss.Style, selected bool, item T) string
	NewItem() T
	InitialFetch() ([]T, error)
	FilterValue(T) string
}

type Item interface {
	GetID() string
	SetID(v string)
}

type Model[T Item] struct {
	del Delegate[T]

	list     list.Model
	spin     spinner.Model
	isLoaded bool
	newItem  CreateNewItemKey

	items   []T
	curItem T

	editor editor.Model

	w, h int
}

func New[T Item](
	w, h int,
	newItemText string,
	del Delegate[T],
) *Model[T] {
	m := &Model[T]{
		del:      del,
		spin:     spinner.Model{},
		isLoaded: false,
		newItem:  CreateNewItemKey(newItemText),
		curItem:  del.NewItem(),
		items:    []T{},
		editor:   editor.Model{},
		w:        w,
		h:        h,
	}

	m.list = list.New([]list.Item{m.newItem}, itemDel[T](m.del.RenderItem), WIDTH_LIST, h)
	m.list.KeyMap = listKeyMap
	m.list.SetShowTitle(false)
	m.list.SetShowHelp(false)
	m.list.SetShowStatusBar(false)
	m.list.FilterInput.SetVirtualCursor(false)
	m.list.FilterInput.Prompt = "Filter: "

	return m
}

type initialResp[T any] []T

func (m *Model[T]) Init() tea.Cmd {
	m.resetEditor()

	batcher := []tea.Cmd{
		func() tea.Msg {
			res, err := m.del.InitialFetch()
			if err != nil {
				log.Panicln("Can't do initial fetch:", err)
			}

			return initialResp[T](res)
		},
		m.spin.Tick,
	}

	if a, ok := m.del.(interface{ Init() tea.Cmd }); ok {
		batcher = append(batcher, a.Init())
	}

	return tea.Batch(batcher...)
}

func (m Model[T]) View() (string, *tea.Cursor) {
	if !m.isLoaded {
		return m.spin.View(), nil
	}

	l := m.list.View()
	e, cur := m.editor.View()
	if cur != nil {
		cur.X += WIDTH_OFFSET_EDITOR
	}

	if m.list.FilterState() == list.Filtering {
		cur = m.list.FilterInput.Cursor()
		cur.X += 2
	}

	listL := lipgloss.NewLayer(l)
	editL := lipgloss.NewLayer(e)
	editL.X(WIDTH_OFFSET_EDITOR)
	splitBar := lipgloss.NewLayer(
		strings.Repeat("║\n", m.h),
	)
	splitBar.X(WIDTH_LIST + 1 + WIDTH_EDITOR_SPLIT_SPACE)

	return lipgloss.NewCanvas(m.w, m.h).Compose(
		lipgloss.NewCompositor(listL, editL, splitBar),
	).Render(), cur
}
