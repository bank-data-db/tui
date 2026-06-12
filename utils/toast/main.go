package toast

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bank-data-db/tui/styles"
	"github.com/bank-data-db/tui/utils"
)

type ToastMsg struct {
	Msg string
	Err bool
}

var toastStyle = lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).Padding(1).Align(lipgloss.Center)

func (t ToastMsg) View(w int) string {
	s := toastStyle
	if t.Err {
		s = s.BorderForeground(styles.COLOR_WRONG)
	} else {
		s = s.BorderForeground(styles.COLOR_MAIN)
	}

	return s.Width(w).Render(t.Msg)
}

func Success(msg string) {
	utils.GlobalMessage <- ToastMsg{
		Msg: msg,
		Err: false,
	}
}

func Error(msg string) {
	utils.GlobalMessage <- ToastMsg{
		Msg: msg,
		Err: true,
	}
}

func CmdSuccess(msg string) tea.Cmd {
	return func() tea.Msg {
		return ToastMsg{
			Msg: msg,
			Err: false,
		}
	}
}

func CmdError(msg string) tea.Cmd {
	return func() tea.Msg {
		return ToastMsg{
			Msg: msg,
			Err: true,
		}
	}
}
