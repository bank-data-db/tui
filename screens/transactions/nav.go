package transactions

import (
	tea "charm.land/bubbletea/v2"
	"github.com/bank-data-db/tui/utils/toast"
)

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
	case "c":
		cmd = tea.Sequence(
			tea.SetClipboard(m.items[m.selected].GetDescription()),
			toast.CmdSuccess("Copied"),
		)
	case "ctrl+f":
		cmd = m.newFilterModal()
	}

	if !msg.Mod.Contains(tea.ModAlt) {
		m.forceViewportIntoSel()
	}

	return cmd
}
