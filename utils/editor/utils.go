package editor

import (
	tea "charm.land/bubbletea/v2"
)


// returns true if the model should handle it
func handleHorizConflict(right bool, pos int, atEnd bool) bool {
	if right && atEnd {
		return false
	} else if !right && pos == 0 {
		return false
	}

	return true
}

type forceReLayout struct{}

func CmdForceReLayout() tea.Msg {
	return forceReLayout{}
}
