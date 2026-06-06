package transactions

import (
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bank-data-db/proto/transactions_pb"
	"github.com/bank-data-db/tui/styles"
	"github.com/bank-data-db/tui/utils"
	"github.com/shadiestgoat/colorutils"
)

const COL_SPLIT = "│"

func (m *Model) cols() []int {
	icon := 2
	amt := 8
	date := 10

	leftover := m.w - icon - amt - date - lipgloss.Width("  "+COL_SPLIT)*5

	catName := int(float64(leftover) * 0.2)
	cardName := int(float64(leftover) * 0.1)

	return []int{date, amt, leftover - catName - cardName, cardName, catName, icon}
}

func (m Model) genericRenderRow(t *transactions_pb.Transaction, rowStyle lipgloss.Style, name string, catRender string) string {
	cols := m.cols()
	str := make([]string, len(cols)-1) // icon is rendered using a diff setup

	str[0] = time.UnixMilli(t.GetAuthedAt()).Format("02/01/2006")
	str[1] = lipgloss.NewStyle().Width(cols[1]).AlignHorizontal(lipgloss.Right).Render(
		strconv.FormatFloat(t.GetAmount(), 'f', 2, 64),
	)
	str[2] = name

	if card := m.cache.Cards.ByID(t.GetCardID()); card != nil {
		str[3] = card.GetName()
	}

	for i, w := range cols[:4] {
		s := rowStyle
		if i == 2 {
			// name gets rendered by parent
			s = lipgloss.NewStyle()
		}
		str[i] = s.Width(w).Render(
			utils.Overflow(str[i], w),
		)
	}

	lastW := cols[4] + cols[5] + 3
	str[4] = lipgloss.NewStyle().Width(lastW).Render(utils.Overflow(catRender, lastW))

	return strings.Join(str, rowStyle.Render(" "+COL_SPLIT+" "))
}

func (m Model) renderRow(t *transactions_pb.Transaction, selected bool) string {
	var rowStyle lipgloss.Style
	cols := m.cols()

	catName, catIcon := "", ""

	if cat := m.cache.Categories.ByID(t.GetResolvedCategoryID()); cat != nil {
		c, err := strconv.ParseInt(cat.GetColor(), 16, 64)
		if err != nil {
			c = 0xffffff
		}

		lum := colorutils.RelativeLuminosity(uint8(c>>16), uint8((c>>8)&0xff), uint8(c&0xff))

		fg := "#ffffff"
		if colorutils.ContrastRatio(lum, colorutils.RelativeLuminosity(0, 0, 0)) > colorutils.ContrastRatio(lum, colorutils.RelativeLuminosity(0xff, 0xff, 0xff)) {
			fg = "#000000"
		}

		rowStyle = lipgloss.NewStyle().Background(
			lipgloss.Color("#" + cat.GetColor()),
		).Foreground(lipgloss.Color(
			fg,
		))

		catName, catIcon = cat.GetName(), cat.GetIcon()
	}

	if selected {
		rowStyle = rowStyle.Background(styles.COLOR_MAIN).Bold(true).Italic(true)
	}

	catRes := rowStyle.Width(cols[4]).Render(
		utils.Overflow(catName, cols[4]),
	) + rowStyle.Render(" "+COL_SPLIT+" ") + rowStyle.Width(cols[5]).AlignHorizontal(lipgloss.Center).Render(
		catIcon,
	)

	name := ""
	if t.HasResolvedName() {
		name = t.GetResolvedName()
	} else {
		name = t.GetDescription()
	}

	return m.genericRenderRow(t, rowStyle, rowStyle.Width(cols[2]).Render(
		utils.Overflow(name, cols[2]),
	), catRes)
}

func (m Model) renderEditRow(t *transactions_pb.Transaction) (string, *lipgloss.Layer, *tea.Cursor) {
	cv, cur := m.editRow.cat.View()
	nameCur := m.editRow.name.Cursor()
	cols := m.cols()
	nameOff := cols[0] + cols[1] + 3*2
	catOff := nameOff + cols[2] + cols[3] + 3*2 - 2

	l := lipgloss.NewLayer(cv)
	l.X(catOff)
	l.Z(999)
	l.Y(-1)

	if cur != nil {
		cur.X += catOff
		cur.Y -= 1
	} else if nameCur != nil {
		cur = nameCur
		cur.X += nameOff
	}

	if cur != nil {
		cur.Y += m.selected - m.viewportOff + 1
	}

	return m.genericRenderRow(t, lipgloss.NewStyle().Background(styles.COLOR_MAIN), m.editRow.name.View(), ""), l, cur
}

func (m Model) vpHeight() int {
	return m.h - 2
}

func (m Model) View() (string, *tea.Cursor) {
	if m.h == 0 {
		return "", nil
	}
	if len(m.items) == 0 {
		return lipgloss.NewStyle().Height(m.h).Width(m.w).Align(
			lipgloss.Center, lipgloss.Center,
		).Render("No Transactions!"), nil
	}

	headers := []string{"Date", "Amount", "Name", "Card", "Category"}
	cols := m.cols()
	for i, v := range headers {
		w := cols[i]
		if i == len(headers)-1 {
			for _, v := range cols[i+1:] {
				w += v
			}
		}

		headers[i] = lipgloss.NewStyle().Width(w).Render(utils.Overflow(v, w))
	}

	header := strings.Join(headers, " "+COL_SPLIT+" ")

	items := m.items[m.viewportOff:]
	items = items[:min(len(items), m.vpHeight())]

	if len(items) == 0 {
		return "No Items here!", nil
	}

	var cur *tea.Cursor
	comp := lipgloss.NewCompositor(lipgloss.NewLayer(header))
	for i, v := range items {
		selected := m.selected == m.viewportOff+i
		var l *lipgloss.Layer
		if m.editRow != nil && selected {
			cont, layer, c := m.renderEditRow(v)
			l = lipgloss.NewLayer(cont)
			cur = c
			l.AddLayers(layer)
		} else {
			l = lipgloss.NewLayer(m.renderRow(v, selected))
		}

		l.Y(i + 1)
		l.Z(len(items) - i)
		comp.AddLayers(l)
	}

	lastRowItems := []string{"Total Transactions: " + strconv.Itoa(m.totalTransactions)}
	if m.nextPageLoading {
		loading := m.loader.View()
		lastRowItems = append(
			lastRowItems,
			loading+lipgloss.NewStyle().Faint(true).Render(" Loading"),
		)
	}

	if m.paginationToken == nil && m.h-len(items) > 4 {
		l := lipgloss.NewLayer(lipgloss.PlaceHorizontal(m.w, lipgloss.Center, "No More Transactions!"))
		l.Y(m.h - 3)
		comp.AddLayers(l)
	}

	res, _ := utils.JoinHorizontalWithSpacer(m.w, 1, lastRowItems...)
	l := lipgloss.NewLayer(res)
	l.Y(m.h - 1)

	comp.AddLayers(l)

	return lipgloss.NewCanvas(m.w, m.h).Compose(comp).Render(), cur
}
