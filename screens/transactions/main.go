package transactions

import (
	"context"
	"log"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bank-data-db/proto/transactions_pb"
	"github.com/bank_data_tui/api"
	"github.com/bank_data_tui/styles"
	"github.com/bank_data_tui/utils"
	"github.com/bank_data_tui/utils/dropdown"
	"github.com/bank_data_tui/utils/repo"
)

type Model struct {
	w, h              int
	selected          int
	viewportOff       int
	items             []*transactions_pb.Transaction
	paginationToken   *string
	api               *api.Client
	cache             *repo.Cache
	loader            spinner.Model
	nextPageLoading   bool
	totalTransactions int
	editRow           *editRow
}

func New(api *api.Client, cache *repo.Cache, w, h int) *Model {
	return &Model{
		w:     w,
		h:     h,
		api:   api,
		cache: cache,
	}
}

type newPageData struct {
	res *transactions_pb.RespList
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.reqNextPage(), func() tea.Msg {
		_, err := m.cache.Categories.MaybeLoad(m.api)
		if err != nil {
			log.Panicln(err)
		}

		return nil
	}, func() tea.Msg {
		_, err := m.cache.Cards.MaybeLoad(m.api)
		if err != nil {
			log.Panicln(err)
		}

		return nil
	})
}

const DE_DUPE_BUFFER = 25

func (m *Model) changeVP(goUp bool) {
	defer m.updateDropdownHeight()

	if goUp {
		if m.viewportOff <= 0 {
			return
		}
		m.viewportOff--
		m.forceSelIntoViewport()
		return
	}

	m.viewportOff++
	visibleItems := len(m.items) - m.viewportOff
	if m.h-visibleItems > 8 {
		m.viewportOff--
	} else {
		m.forceSelIntoViewport()
	}
}

func (m *Model) handleKeyNormal(msg tea.KeyPressMsg) tea.Cmd {
	var cmd tea.Cmd

	switch k := msg.String(); k {
	case "down":
		if m.selected != len(m.items)-1 {
			m.selected++
		}
	case "up":
		if m.selected != 0 {
			m.selected--
		}
	case "end":
		m.selected = len(m.items) - 1
	case "start":
		m.selected = 0
	case "alt+down", "alt+up":
		m.changeVP(k == "alt+up")
	case "e":
		cmd = m.newEditRow()
	}

	if !msg.Mod.Contains(tea.ModAlt) {
		m.forceViewportIntoSel()
	}

	return cmd
}

func (m Model) Update(msg tea.Msg) (utils.Screen, tea.Cmd) {
	var cmd tea.Cmd
	batch := []tea.Cmd{}

	passToChildren := true

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.editRow == nil {
			cmd := m.handleKeyNormal(msg)
			batch = append(batch, cmd)
			passToChildren = false
		} else {
			handled, cmd := m.handleKeyEditMode(msg)
			batch = append(batch, cmd)
			if handled {
				passToChildren = false
			}
		}
	case newPageData:
		if !msg.res.HasPaginationToken() {
			m.paginationToken = nil
		} else {
			m.paginationToken = new(msg.res.GetPaginationToken())
		}
		m.nextPageLoading = false
		m.totalTransactions = int(msg.res.GetTotalCount())
		m.items = append(m.items, msg.res.GetResult()...)
	case utils.ResizeMessage:
		m.w, m.h = msg.W, msg.H
		m.resizeEditRow()
		m.forceViewportIntoSel()
	case tea.MouseWheelMsg:
		if m.editRow == nil {
			switch msg.Button {
			case tea.MouseWheelDown:
				m.changeVP(false)
			case tea.MouseWheelUp:
				m.changeVP(true)
			}
		}
	case dropdown.SelectMsg:
		if m.editRow != nil {
			batch = append(batch, m.editRow.toggleFocus())
		}
	case transUpdated:
		t := m.items[m.selected]
		if v := m.editRow.name.Value(); v == "" {
			t.ClearResolvedName()
		} else {
			t.SetResolvedName(v)
		}
		if v := m.editRow.cat.Value(); v == "" {
			t.ClearResolvedCategoryId()
		} else {
			t.SetResolvedCategoryID(m.editRow.cat.Value())
		}

		m.editRow = nil
	}

	if m.nextPageLoading {
		m.loader, cmd = m.loader.Update(msg)
		batch = append(batch, cmd)
	}

	if m.paginationToken != nil && !m.nextPageLoading && m.indexIsVisible(-LOAD_OFFSET) {
		batch = append(batch, m.reqNextPage())
	}

	if passToChildren && m.editRow != nil {
		if m.editRow.cat.Focused() {
			cat, cmd := m.editRow.cat.Update(msg)
			batch = append(batch, cmd)
			m.editRow.cat = cat
		} else {
			name, cmd := m.editRow.name.Update(msg)
			batch = append(batch, cmd)
			m.editRow.name = name
		}
	}

	return m, tea.Batch(batch...)
}

func (m *Model) forceViewportIntoSel() {
	if len(m.items) <= m.h {
		m.viewportOff = 0
		return
	}

	if m.selected < m.viewportOff {
		m.viewportOff = m.selected
	} else if m.selected > m.viewportOff+m.vpHeight()-1 {
		m.viewportOff = m.selected - (m.vpHeight() - 1)
	}
}

func (m *Model) forceSelIntoViewport() {
	if len(m.items) <= m.h {
		m.viewportOff = 0
		return
	}

	if m.selected < m.viewportOff {
		m.selected = m.viewportOff
	} else if m.selected > m.viewportOff+m.vpHeight()-1 {
		m.selected = m.viewportOff + m.vpHeight() - 1
	}
}

// -nth item to be visible for the next page to be loaded
const LOAD_OFFSET = 5

func (m Model) indexIsVisible(n int) bool {
	if n < 0 {
		n += len(m.items)
	}

	return m.viewportOff <= n && n <= m.viewportOff+m.vpHeight()
}

func (m *Model) reqNextPage() tea.Cmd {
	m.nextPageLoading = true
	m.loader = spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(lipgloss.NewStyle().Foreground(styles.COLOR_MAIN)),
	)

	return tea.Batch(
		func() tea.Msg {
			d, err := m.api.TransactionsList(context.Background(), transactions_pb.ReqList_builder{
				PaginationToken: m.paginationToken,
				OrderBy:         transactions_pb.OrderFieldAuthedAt.Enum(),
				Descending:      new(true),
				PageSize:        new(uint32(100)),
			}.Build())
			if err != nil {
				log.Panicln(err)
			}

			return newPageData{
				res: d,
			}
		},
		m.loader.Tick,
	)
}
