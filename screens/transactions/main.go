package transactions

import (
	"log"
	"slices"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bank_data_tui/api"
	"github.com/bank_data_tui/styles"
	"github.com/bank_data_tui/utils"
	"github.com/bank_data_tui/utils/repo"
)

type Model struct {
	w, h              int
	selected          int
	viewportOff       int
	items             []*api.Transaction
	hasHitLastPage    bool
	lastDataPage      int
	api               *api.APIClient
	cache             *repo.Cache
	loader            spinner.Model
	nextPageLoading   bool
	totalTransactions int
}

func New(api *api.APIClient, cache *repo.Cache, w, h int) *Model {
	return &Model{
		w:     w,
		h:     h,
		api:   api,
		cache: cache,
	}
}

type newPageData struct {
	*api.RespPages[[]*api.Transaction]
	page     int
	override bool
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.forceRequestPage(1), func() tea.Msg {
		_, err := m.cache.EasyCategories(m.api)
		if err != nil {
			panic(err)
		}

		return nil
	})
}

const DE_DUPE_BUFFER = 25

func (m *Model) changeVP(goUp bool) {
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

func (m Model) Update(msg tea.Msg) (utils.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
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
		}

		if !msg.Mod.Contains(tea.ModAlt) {
			m.forceViewportIntoSel()
		}
	case newPageData:
		if msg.override {
			m.items = msg.Data
		} else {
			if m.lastDataPage+1 != msg.page {
				break
			}
			sl := slices.DeleteFunc(msg.Data, func(vb *api.Transaction) bool {
				for _, va := range m.items[max(len(m.items)-DE_DUPE_BUFFER, 0):] {
					if va.ID == vb.ID {
						return true
					}
				}
				return false
			})

			m.items = append(m.items, sl...)
		}

		if len(msg.Data) != 50 {
			m.hasHitLastPage = true
		}

		m.nextPageLoading = false
		m.lastDataPage = msg.page
		m.totalTransactions = msg.Total
	case utils.ResizeMessage:
		m.w, m.h = msg.W, msg.H
		m.forceViewportIntoSel()
	case tea.MouseWheelMsg:
		switch msg.Button {
		case tea.MouseWheelDown:
			m.changeVP(false)
		case tea.MouseWheelUp:
			m.changeVP(true)
		}
	}

	batch := []tea.Cmd{}
	var cmd tea.Cmd
	if m.nextPageLoading {
		m.loader, cmd = m.loader.Update(msg)
		batch = append(batch, cmd)
	}

	if !m.hasHitLastPage && !m.nextPageLoading && m.indexIsVisible(-LOAD_OFFSET) {
		batch = append(batch, m.reqPage(m.lastDataPage+1))
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

	log.Println("sel", m.selected, "vpheight", m.vpHeight(), "off", m.viewportOff, "lastItem", m.viewportOff+m.vpHeight())

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

const REQ_DEDUPE_PERIOD = 1 * time.Minute

func (m *Model) reqPage(n int) tea.Cmd {
	if m.nextPageLoading {
		return nil
	}

	return m.forceRequestPage(n)
}

func (m *Model) forceRequestPage(n int) tea.Cmd {
	log.Println("Loading")
	m.nextPageLoading = true
	m.loader = spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(lipgloss.NewStyle().Foreground(styles.COLOR_MAIN)),
	)

	return tea.Batch(
		func() tea.Msg {
			d, err := m.api.TransactionsFetch(api.TOR_AUTH, n, false)
			if err != nil {
				log.Panicln(err)
			}

			return newPageData{
				RespPages: d,
				page:      n,
			}
		},
		m.loader.Tick,
	)
}
