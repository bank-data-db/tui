package dropdown

import (
	"log"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bank_data_tui/styles"
	"github.com/bank_data_tui/utils"
	"github.com/charmbracelet/x/ansi"
)

type Value struct {
	Display string
	Value   string
	// Optional: a text-only representation of this value
	DisplayText string
	// Optional: a text to use for filtering
	SearchText string
}

var EmptyValue = &Value{
	Display:     "<none>",
	DisplayText: "",
	Value:       "",
	SearchText:  "none",
}

type Model struct {
	inp            textinput.Model
	MaxHeight      int
	vals           []*Value
	curSel         int
	vOff           int
	cur            int
	valToI         map[string]int
	filteredValues []*Value
	title          string
}

func New(vals []*Value, title string, maxHeight int) Model {
	m := Model{
		inp:       textinput.New(),
		MaxHeight: maxHeight,
		cur:       -1,
		curSel:    -1,
		title:     title,
	}

	m.inp.Prompt = ""
	m.inp.SetVirtualCursor(false)
	m.inp.SetStyles(textinput.Styles{
		Focused: textinput.StyleState{
			Text:        styles.S_TEXT_HIGHLIGHT,
			Placeholder: styles.S_TEXT_DISABLED,
		},
		Blurred: textinput.StyleState{
			Text:        styles.S_TEXT_DISABLED,
			Placeholder: styles.S_TEXT_DISABLED,
		},
		Cursor: styles.TI_CURSOR,
	})

	m.SetValues(vals)

	return m
}

func (m *Model) Focus() tea.Cmd {
	if m.cur == -1 {
		if len(m.filteredValues) == 0 {
			m.curSel = -1
		} else {
			m.curSel = 0
		}
	} else {
		m.curSel = m.cur
	}
	m.filteredValues = m.vals

	return m.inp.Focus()
}

func (m Model) Focused() bool {
	return m.inp.Focused()
}

func (m *Model) Blur() {
	// Errs are set explicitly, so focus should not change this
	err := m.inp.Err
	m.inp.SetValue("")
	m.inp.Err = err
	m.inp.Blur()
}

func (m *Model) SetValues(vals []*Value) {
	var curVal, curSelVal string
	if m.cur != -1 {
		curVal = m.vals[m.cur].Value
	}
	if m.curSel != -1 {
		curSelVal = m.filteredValues[m.curSel].Value
	}

	m.valToI = map[string]int{}
	m.cur = -1
	m.curSel = -1
	longest := 0
	m.filteredValues = vals
	m.vals = vals

	for i, v := range vals {
		if w := lipgloss.Width(v.Display); w > longest {
			longest = w
		}
		if v != EmptyValue {
			if v.DisplayText == "" {
				v.DisplayText = ansi.Strip(v.Display)
			}
			if v.SearchText == "" {
				v.SearchText = v.DisplayText
			}
			v.SearchText = strings.ToLower(v.SearchText)
		}
		if v.Value == curVal {
			m.cur = i
		}
		if v.Value == curSelVal {
			m.curSel = i
		}
		if _, ok := m.valToI[v.Value]; ok {
			log.Panicf("Oh-oh.. duplicate value: '%s'\n", v.Value)
		}

		m.valToI[v.Value] = i
	}

	m.inp.SetWidth(longest)
	if m.cur == -1 {
		m.inp.Placeholder = ""
	} else {
		m.inp.Placeholder = m.vals[m.cur].DisplayText
	}
	if m.curSel == -1 {
		m.inp.SetValue("")
	}
}

func (m Model) Values() []*Value {
	return m.vals
}

func (m *Model) Value() string {
	if m.cur == -1 {
		return ""
	}
	v := m.vals[m.cur]
	// ptr comparison!!
	if v == EmptyValue {
		return ""
	}

	return v.Value
}

func (m *Model) SetValue(nv string) {
	m.cur = m.valToI[nv]
	m.inp.Placeholder = m.vals[m.valToI[nv]].DisplayText
	m.inp.SetValue("")
}

// the extra heigh introduced by text input, styling, etc
const nonOptionHeight = 1 + 2 + 1 // text input + border top/bottom + divider

type SelectMsg string

func (m *Model) adjustVP(up bool) {
	if up {
		if m.vOff != 0 {
			m.vOff--
		}
	} else {
		m.vOff++
		if m.vOff+m.MaxHeight >= len(m.filteredValues)+nonOptionHeight {
			m.vOff--
		}
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if !m.Focused() {
			return m, nil
		}

		switch k := msg.String(); k {
		case "down":
			m.curSel++
			if m.curSel >= len(m.filteredValues) {
				m.curSel--
			}
			return m, nil
		case "up":
			if m.curSel != 0 {
				m.curSel--
			}
			return m, nil
		case "alt+down":
			m.adjustVP(false)
			return m, nil
		case "alt+up":
			m.adjustVP(true)
			return m, nil
		case "enter":
			if m.curSel != -1 {
				val := m.filteredValues[m.curSel]
				m.SetValue(val.Value)

				return m, func() tea.Msg { return SelectMsg(val.Value) }
			}
		}
	case tea.MouseWheelMsg:
		switch msg.Button {
		case tea.MouseWheelDown:
			m.adjustVP(false)
		case tea.MouseWheelUp:
			m.adjustVP(true)
		}
	}

	oldV := m.inp.Value()
	inpM, cmd := m.inp.Update(msg)
	m.inp = inpM

	nv := m.inp.Value()
	if oldV != nv {
		filterTargets := m.vals
		if len(nv) > len(oldV) {
			// we added more text, so we can filter out the currently filtered items
			filterTargets = m.filteredValues
		}

		fv := make([]*Value, 0, len(m.filteredValues))
		searchStr := strings.ToLower(nv)
		curSelV := ""
		if m.curSel != -1 {
			curSelV = m.filteredValues[m.curSel].Value
		}
		m.curSel = -1
		for _, v := range filterTargets {
			off := 0
			found := true

			for _, sr := range searchStr {
				found = false
				log.Printf("Searching for '%s' in '%s'\n", string([]rune{sr}), v.SearchText[off:])
				for i, r := range v.SearchText[off:] {
					if sr == r {
						off += i + 1
						found = true
						break
					}
				}
				if !found {
					break
				}
			}
			if found {
				if v.Value == curSelV {
					m.curSel = len(fv)
				}
				fv = append(fv, v)
			}
		}

		m.filteredValues = fv
		if m.curSel == -1 && len(fv) != 0 {
			m.curSel = 0
		}
	}

	return m, cmd
}

func (m *Model) SetWidth(w int) {
	m.inp.SetWidth(w - 4 - 1)
}

func (m Model) Width() int {
	return m.inp.Width() + 4 + 1 // + 1 bc of the stupid extra bs
}

func (m Model) View() (string, *tea.Cursor) {
	s := styles.STYLE_FIELD.BorderTop(false)

	c := m.inp.Cursor()
	if c != nil {
		c.X += 2
		c.Y += 1
	}

	availHeight := m.MaxHeight - nonOptionHeight
	if m.inp.Focused() {
		s = s.BorderForeground(styles.COLOR_SECONDARY)
	}

	header := utils.RenderHeader(m.Width(), s.GetBorderTopForeground(), m.title) + "\n"
	str := ""
	if !m.Focused() && m.inp.Err != nil {
		s = s.BorderBottom(false)
		str += utils.RenderErrFooter(
			m.Width(), s.GetBorderTopForeground(), m.inp.Err,
		)
	}

	if !m.Focused() || availHeight <= 0 || len(m.filteredValues) == 0 {
		return header + s.Render(m.inp.View()) + str, c
	}

	str += s.BorderBottom(false).Render(m.inp.View())

	str += "\n" + lipgloss.NewStyle().Foreground(styles.COLOR_MAIN).Render(
		// x2 of padding, x1 for extra space at the end of the textinput
		"╟"+strings.Repeat("─", m.inp.Width()+2+1)+"╢",
	) + "\n"

	options := ""
	for i, v := range m.filteredValues[m.vOff:min(len(m.filteredValues), m.vOff+availHeight)] {
		s := lipgloss.NewStyle()
		left, right := "╟⋲", "⋺╢"

		if m.curSel == m.vOff+i {
			s = s.Foreground(styles.COLOR_SECONDARY)
		} else if m.valToI[v.Value] == m.cur {
			s = s.Foreground(styles.COLOR_MAIN)
		} else {
			left, right = "║ ", " ║"
			s = s.Foreground(styles.COLOR_MAIN)
		}

		disp := utils.Overflow(v.Display, m.inp.Width())

		// double space for the last part for the extra space at the end of text input
		options += "\n" + s.Render(left) + disp + strings.Repeat(" ", m.inp.Width()-lipgloss.Width(disp)+1) + s.Render(right)
	}
	options += "\n" + lipgloss.NewStyle().Foreground(styles.COLOR_MAIN).Render("╚"+strings.Repeat("═", m.inp.Width()+2+1)+"╝")

	if len(options) != 0 {
		options = options[1:]
	}

	return header + str + options, c
}

func (m *Model) SetErr(err error) {
	m.inp.Err = err
}
func (m *Model) Err() error {
	return m.inp.Err
}

func (m Model) InputCursorPosition() int {
	return m.inp.Position()
}

func (m Model) InputCursorAtEnd() bool {
	return m.inp.Position() == lipgloss.Width(m.inp.Value())
}
