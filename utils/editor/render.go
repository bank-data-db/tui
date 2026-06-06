package editor

import (
	"log"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bank_data_tui/styles"
	"github.com/bank_data_tui/utils"
	"github.com/charmbracelet/x/ansi"
)

func (m Model) valid() bool {
	for _, f := range m.fields {
		f, ok := f.(inputField)
		if ok {
			if f.GetErr() != nil {
				return false
			}
		}
	}
	return true
}

func (m Model) renderButton(text string, selected bool, evil bool, disabled bool) string {
	s := styles.STYLE_FIELD.Padding(1, m.buttonPad)

	if !selected {
	} else if disabled {
		s = s.Background(styles.COLOR_DISABLED)
	} else if evil {
		s = s.Background(styles.COLOR_WRONG)
	} else {
		s = s.Background(styles.COLOR_MAIN)
	}

	if selected {
		if disabled {
			s = s.BorderForeground(styles.COLOR_MAIN)
		} else {
			s = s.BorderForeground(styles.COLOR_SECONDARY)
		}
	} else if disabled {
		s = s.BorderForeground(styles.COLOR_DISABLED)
	}

	return s.Render(text)
}

func (m Model) buttonLayer(y int) *lipgloss.Layer {
	parentLayer := lipgloss.NewLayer("")

	parentLayer.Y(y)

	for _, v := range m.layout[len(m.layout)-1] {
		saveBtn := v.fieldID == BTN_SAVE_ID

		l := lipgloss.NewLayer(m.renderButton(
			m.btnText(v.fieldID),
			m.focusedField == v.fieldID,
			!saveBtn,
			saveBtn && !m.valid(),
		))
		l.X(v.x)
		parentLayer.AddLayers(l)
	}

	return parentLayer
}

func (m Model) View() (string, *tea.Cursor) {
	if m.width == 0 {
		return "", nil
	}

	compose := lipgloss.NewCompositor()
	var cur *tea.Cursor

	usedHeight := 0
	for i, row := range m.layout[:len(m.layout)-1] {
		for _, ld := range row {
			f := m.fields[ld.fieldID]
			d, c := f.View()
			if c != nil {
				c.X += ld.x
				c.Y += usedHeight
				cur = c
			}

			l := lipgloss.NewLayer(d)
			l.Y(usedHeight)
			l.X(ld.x)
			l.Z(len(m.layout) - i)
			compose.AddLayers(l)
		}

		usedHeight += m.rowHeights[i]
	}

	compose.AddLayers(m.buttonLayer(usedHeight))

	if m.confirmDial != nil {
		w := m.width - 4
		s := lipgloss.NewStyle().Border(lipgloss.BlockBorder()).Width(w).Align(lipgloss.Center)

		btnRow, _ := utils.JoinHorizontalEqualSpread(
			w-4,
			m.renderButton("Yes!", m.confirmDial.atYes, true, false),
			m.renderButton("No :(", !m.confirmDial.atYes, false, false),
		)
		log.Println(ansi.Strip(btnRow))
		c := s.Render(
			"\n",
			m.confirmDial.text,
			"\n",
			btnRow,
			"\n",
		)

		l := lipgloss.NewLayer(c)
		l.X(2)
		l.Y((int(float64(m.height)*0.7) - lipgloss.Height(c)) / 2)
		l.Z(900)

		compose.AddLayers(l)
	}

	return lipgloss.NewCanvas(m.width, m.height).Compose(compose).Render(), cur
}
