package toast

import (
	"charm.land/lipgloss/v2"
	"github.com/bank_data_tui/styles"
	"github.com/bank_data_tui/utils"
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
