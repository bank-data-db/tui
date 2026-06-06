package editor

import "github.com/bank_data_tui/utils"

const (
	BTN_SAVE_ID         = -1
	BTN_SAVE_TXT        = "Save"
	BTN_SAVE_TXT_UPDATE = "Update"

	BTN_DEL_ID  = -2
	BTN_DEL_TXT = "Delete"

	BTN_RESET_ID  = -3
	BTN_RESET_TXT = "Reset"
)

// So I think that in the future, I MIGHT add a button field BUT
// It'll be different from these in impl. The way that these scale is good
// for special save buttons at the bottom, but NOT for real inputs

func btnSizing(saved bool, padding int) int {
	c := 2
	base := len(BTN_RESET_TXT)
	if saved {
		base += len(BTN_DEL_TXT) + len(BTN_SAVE_TXT_UPDATE)
		c++
	} else {
		base += len(BTN_SAVE_TXT)
	}

	return c*(padding*2) + base
}

func (m *Model) resetButtonLayout(w int, saved bool, padding int) {
	var ids []int

	if saved {
		ids = []int{
			BTN_SAVE_ID,
			BTN_DEL_ID,
			BTN_RESET_ID,
		}
	} else {
		ids = []int{
			BTN_SAVE_ID,
			BTN_RESET_ID,
		}
	}

	arr := make([]*layoutData, len(ids))
	sizes := make([]int, len(ids))

	for i, id := range ids {
		arr[i] = &layoutData{
			fieldID:  id,
			canFocus: true,
			grow:     false,
		}
		sizes[i] = len(m.btnText(id)) + (padding+1)*2
	}

	i := 0
	for off := range utils.EqualSpreadSeq(w, sizes) {
		arr[i].x = off
		i++
	}

	if !saved {
		arr[1].fieldID = BTN_RESET_ID
	}

	m.buttonPad = padding
	if m.layout[len(m.layout)-1][0].fieldID < 0 {
		m.layout[len(m.layout)-1] = arr
	} else {
		m.layout = append(m.layout, arr)
	}
}
