package transactions

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bank-data-db/tui/styles"
	"github.com/bank-data-db/tui/utils"
	"github.com/bank-data-db/tui/utils/dropdown"
)

type filterModal struct {
	resolved dropdown.Model
	cards    dropdown.Model
}

func (f *filterModal) toggleFocus() tea.Cmd {
	if f.resolved.Focused() {
		f.resolved.Blur()
		return f.cards.Focus()
	}

	f.cards.Blur()
	return f.resolved.Focus()
}

func (m *Model) newFilterModal() tea.Cmd {
	resolved := dropdown.New([]*dropdown.Value{
		dropdown.EmptyValue,
		{
			Display: "Resolved Only",
			Value:   "t",
		},
		{
			Display: "Not Resolved Only",
			Value:   "f",
		},
	}, "Resolved", 10)
	if m.reqBuilder.Resolved != nil {
		if *m.reqBuilder.Resolved {
			resolved.SetValue("t")
		} else {
			resolved.SetValue("f")
		}
	}

	cards := dropdown.New(m.cache.Cards.DropdownValues(true), "Cards", 10)
	if m.reqBuilder.CardId != nil {
		cards.SetValue(*m.reqBuilder.CardId)
	}

	m.filterModal = &filterModal{
		resolved: resolved,
		cards:    cards,
	}

	return m.filterModal.resolved.Focus()
}

const FILTER_MODAL_HEIGHT = 9

func (m Model) viewFilterModal() (utils.LayerGroup, *tea.Cursor) {
	resolved, rCur := m.filterModal.resolved.View()
	cards, cCur := m.filterModal.cards.View()
	rw := lipgloss.Width(resolved)
	cw := lipgloss.Width(cards)

	rl := lipgloss.NewLayer(resolved)
	cl := lipgloss.NewLayer(cards)

	w := max(rw, cw)
	rl.X((w-rw)/2 + 2)
	rl.Z(10)
	rl.Y(1)
	cl.X((w-cw)/2 + 2)
	cl.Y(5)
	cl.Z(9)

	bg := lipgloss.NewLayer(
		styles.STYLE_FIELD.Padding(2).Width(
			w + 4,
		).Height(FILTER_MODAL_HEIGHT).Render(),
	)

	return utils.LayerGroup{bg, rl, cl}, utils.Coalesce(
		utils.CursorLayer(rCur, rl),
		utils.CursorLayer(cCur, cl),
	)
}

func (m *Model) handleKeyModal(msg tea.KeyPressMsg) (bool, tea.Cmd) {
	switch k := msg.String(); k {
	case "esc":
		switch m.filterModal.resolved.Value() {
		case "":
			m.reqBuilder.Resolved = nil
		case "t":
			m.reqBuilder.Resolved = new(true)
		case "f":
			m.reqBuilder.Resolved = new(false)
		}

		if m.filterModal.cards.Value() == "" {
			m.reqBuilder.CardId = nil
		} else {
			m.reqBuilder.CardId = new(m.filterModal.cards.Value())
		}

		m.filterModal = nil
		// Reset
		m.items = nil
		m.paginationToken = nil
		m.nextPageLoading = true
		m.totalTransactions = 0
		m.selected = 0
		m.viewportOff = 0

		return true, m.reqNextPage()
	case "tab", "shift+tab":
		cmd := m.filterModal.toggleFocus()
		return true, cmd
	}
	return false, nil
}
