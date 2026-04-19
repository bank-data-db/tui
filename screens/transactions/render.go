package transactions

import (
	"slices"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bank_data_tui/api"
	"github.com/bank_data_tui/styles"
	"github.com/bank_data_tui/utils"
	"github.com/shadiestgoat/colorutils"
)

const COL_SPLIT = "│"

func (m *Model) cols() []int {
	icon := 2 // the icon doesn't have a spacer around COLO_SPLIT
	amt := 8
	date := 10

	leftover := m.w - icon - amt - date - lipgloss.Width("  "+COL_SPLIT)*4

	catName := int(float64(leftover) * 0.3)

	return []int{date, amt, leftover - catName, catName, icon}
}

func (m Model) renderRow(t *api.Transaction, selected bool) string {
	rowStyle := lipgloss.NewStyle()
	if selected {
		rowStyle = rowStyle.Background(styles.COLOR_MAIN).Bold(true).Italic(true)
	}

	cols := m.cols()
	str := make([]string, len(cols))
	var cat *api.Category

	if t.ResolvedCategoryID != nil {
		i := slices.IndexFunc(m.cache.Categories, func(c *api.Category) bool {
			return c.ID == *t.ResolvedCategoryID
		})
		if i != -1 {
			cat = m.cache.Categories[i]
			str[4] = m.cache.Categories[i].Icon
		}
	}
	if cat != nil {
		c, err := strconv.ParseInt(cat.Color, 16, 64)
		if err != nil {
			c = 0xffffff
		}

		lum := colorutils.RelativeLuminosity(uint8(c>>16), uint8((c>>8)&0xff), uint8(c&0xff))

		fg := "#ffffff"
		if colorutils.ContrastRatio(lum, colorutils.RelativeLuminosity(0, 0, 0)) > colorutils.ContrastRatio(lum, colorutils.RelativeLuminosity(0xff, 0xff, 0xff)) {
			fg = "#000000"
		}

		if !selected {
			rowStyle = lipgloss.NewStyle().Background(
				lipgloss.Color("#" + cat.Color),
			).Foreground(lipgloss.Color(
				fg,
			))
		}

		str[4] = rowStyle.Width(cols[4]).AlignHorizontal(lipgloss.Center).Render(str[4])
		str[3] = cat.Name
	}

	str[0] = t.AuthedAt.Format("02/01/2006")
	str[1] = lipgloss.NewStyle().Width(cols[1]).AlignHorizontal(lipgloss.Right).Render(
		strconv.FormatFloat(t.Amount, 'f', 2, 64),
	)

	if t.ResolvedName != nil {
		str[2] = *t.ResolvedName
	} else {
		str[2] = t.Desc
	}

	for i, w := range cols {
		str[i] = rowStyle.Width(w).Render(
			utils.Overflow(str[i], w),
		)
	}

	colSplitter := rowStyle.Render(" " + COL_SPLIT + " ")

	return strings.Join(str, colSplitter)
}

func (m Model) vpHeight() int {
	return m.h - 2
}

func (m Model) View() (string, *tea.Cursor) {
	if m.h == 0 || len(m.items) == 0 {
		return "", nil
	}

	headers := []string{"Date", "Amount", "Name", "Category"}
	cols := m.cols()
	for i, v := range headers {
		headers[i] = lipgloss.NewStyle().Width(cols[i]).Render(utils.Overflow(v, cols[i]))
	}

	header := strings.Join(headers, " "+COL_SPLIT+" ")

	items := m.items[m.viewportOff:]
	items = items[:min(len(items), m.vpHeight())]

	if len(items) == 0 {
		return "No Items here!", nil
	}

	rows := header + "\n"
	for i, v := range items {
		rows += m.renderRow(v, m.selected == m.viewportOff+i) + "\n"
	}

	lastRowItems := []string{"Total Transactions: " + strconv.Itoa(m.totalTransactions)}
	if m.nextPageLoading {
		loading := m.loader.View()
		lastRowItems = append(
			lastRowItems,
			loading+lipgloss.NewStyle().Faint(true).Render(" Loading"),
		)
	}

	rows = rows[:len(rows)-1]

	if m.hasHitLastPage && m.h-len(items) > 4 {
		rows += "\n\n\n" + lipgloss.PlaceHorizontal(m.w, lipgloss.Center, "No More Transactions!")
	}

	res, _ := utils.JoinHorizontalWithSpacer(m.w, 1, lastRowItems...)

	return rows + strings.Repeat("\n", m.h-strings.Count(rows, "\n")-1) + res, nil
}
