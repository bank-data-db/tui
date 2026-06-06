package main

import (
	"log"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bank_data_tui/styles"
	"github.com/bank_data_tui/utils"
)

var (
	STYLE_HEADER_TEXT     = lipgloss.NewStyle().Foreground(styles.COLOR_MAIN).Margin(1)
	STYLE_HEADER_SELECTED = STYLE_HEADER_TEXT.Bold(true).Underline(true)
	STYLE_HEADER          = lipgloss.NewStyle().Border(lipgloss.DoubleBorder(), false, false, true, false).Margin(0, 0, 1, 0).BorderForeground(styles.COLOR_MAIN)
)

const (
	// 1 (margin bottom) + 1 * 2 (padding top & bot) + 1 line of border + 1 line of text
	HEADER_HEIGHT = 1 + 1*2 + 1 + 1
)

var HEADER_SCREENS = []struct {
	s utils.ScreenID
	t string
}{
	{utils.S_TRANS, "Transactions"},
	{utils.S_MAPPINGS, "Mappings"},
	{utils.S_CATEGORIES, "Categories"},
	{utils.S_CARDS, "Cards"},
	{utils.S_UPLOAD, "Upload"},
}

func (m mainApp) renderHeader() string {
	r := []string{}
	for _, h := range HEADER_SCREENS {
		if h.s == m.curFocusedScreen {
			r = append(r, STYLE_HEADER_SELECTED.Render(h.t))
		} else {
			r = append(r, STYLE_HEADER_TEXT.Render(h.t))
		}
	}

	left := r[0]
	right := lipgloss.JoinHorizontal(lipgloss.Top, r[1:]...)

	spacer := strings.Repeat(" ", m.width-lipgloss.Width(right)-lipgloss.Width(left))

	return STYLE_HEADER.Render(lipgloss.JoinHorizontal(lipgloss.Top, left, spacer, right))
}

func (m mainApp) renderTooSmall() string {
	box := lipgloss.NewStyle().Width(m.width).Height(m.height).Align(lipgloss.Center, lipgloss.Center)
	return box.Render("Too Small")
}

func (m mainApp) View() (v tea.View) {
	v.AltScreen = true
	v.WindowTitle = "Bank Data"
	v.MouseMode = tea.MouseModeCellMotion

	if m.width == 0 || m.height == 0 {
		return
	} else if m.width < 50 || m.height < 20 {
		v.SetContent(m.renderTooSmall())
		return
	}

	s, c := m.screenImp.View()
	v.Cursor = c

	w, h := lipgloss.Width(s), lipgloss.Height(s)
	if h > (m.height-HEADER_HEIGHT) || w > m.width {
		v.Cursor = nil
		if h > (m.height - HEADER_HEIGHT) {
			log.Println("Height too big", m.height-HEADER_HEIGHT, h)
		}
		if w > m.width {
			log.Println("Width too big", m.width, w)
		}

		v.SetContent(m.renderTooSmall())
		return
	}

	if m.curFocusedScreen == utils.S_LOGIN {
		padTop := (m.height - h) / 2
		padLeft := (m.width - w) / 2

		if c != nil {
			c.Y += padTop
			c.X += padLeft
		}

		v.SetContent(lipgloss.NewStyle().Padding(padTop, 0, 0, padLeft).Render(s))
		return
	}

	header := m.renderHeader()
	padTop := (m.height - h - HEADER_HEIGHT) / 2
	padLeft := (m.width - w) / 2

	if c != nil {
		c.Y += padTop + HEADER_HEIGHT
		c.X += padLeft
	}

	content := lipgloss.NewLayer(header + "\n" + lipgloss.NewStyle().Padding(padTop, 0, 0, padLeft).Render(s))
	toastData := ""
	toastWidth := int(float64(m.width) * 0.2)
	for _, v := range m.toasts[:min(len(m.toasts), 2)] {
		toastData += v.View(toastWidth) + "\n"
	}
	toasts := lipgloss.NewLayer(strings.TrimRight(toastData, "\n"))
	toasts.Z(20)
	toasts.X(m.width - toastWidth)

	v.SetContent(lipgloss.NewCompositor(content, toasts).Render())

	return v
}
