package editor

import (
	"strconv"

	tea "charm.land/bubbletea/v2"
)

func (c *Model) focusFirstField() tea.Cmd {
	for _, row := range c.layout {
		for _, f := range row {
			if f.canFocus {
				return c.focusField(f.fieldID)
			}
		}
	}

	return nil
}

func (c *Model) focusField(f int) tea.Cmd {
	if !c.inButtons(c.focusedField) {
		oldF, ok := c.fields[c.focusedField].(inputField)

		if ok {
			oldF.Blur()
		}
	}

	c.focusedField = f
	if c.inButtons(c.focusedField) {
		return nil
	}

	newF, ok := c.fields[c.focusedField].(inputField)
	var cmd tea.Cmd
	if ok {
		cmd = newF.Focus()
	}

	return cmd
}

func (c Model) inButtons(i int) bool {
	return i < 0
}

func (c Model) handleNavKey(key string) int {
	switch key {
	case "tab", "right":
		return c.navKeyHorizontal(1)
	case "shift+tab", "left":
		return c.navKeyHorizontal(-1)
	case "enter":
		return c.navKeyHorizontal(1)
	case "down":
		return c.navKeyVertical(1)
	case "up":
		return c.navKeyVertical(-1)
	}

	panic("unknown nav key? " + key)
}

// x, y
func (c Model) rowCol(f int) (int, int) {
	for y, row := range c.layout {
		for x, c := range row {
			if c.fieldID == f {
				return x, y
			}
		}
	}

	return 0, 0
}

func (c Model) navKeyHorizontal(dir int) int {
	cx, cy := c.rowCol(c.focusedField)

	y := cy
	x := cx + dir
	for {
		for x >= 0 && x < len(c.layout[y]) {
			f := c.layout[y][x]
			if f.canFocus {
				return f.fieldID
			}
			x += dir
		}
		y += dir
		if y == len(c.layout) {
			y = 0
		} else if y < 0 {
			y = len(c.layout) - 1
		}
		if dir > 0 {
			x = 0
		} else {
			x = len(c.layout[y]) - 1
		}
	}
}

func (c Model) btnText(id int) string {
	switch id {
	case BTN_SAVE_ID:
		if c.item.GetID() == "" {
			return BTN_SAVE_TXT
		}
		return BTN_SAVE_TXT_UPDATE
	case BTN_RESET_ID:
		return BTN_RESET_TXT
	case BTN_DEL_ID:
		return BTN_DEL_TXT
	default:
		panic("Unknown button id: " + strconv.Itoa(id))
	}
}

func (c Model) fieldWidth(f int) int {
	if c.inButtons(f) {
		for _, v := range c.layout[len(c.layout)-1] {
			if v.fieldID == f {
				baseSize := c.buttonPad*2 + 2
				return len(c.btnText(v.fieldID)) + baseSize
			}
		}
	}

	return c.fields[f].Width()
}

func (c Model) navKeyVertical(dir int) int {
	cx, cy := c.rowCol(c.focusedField)

	y := cy + dir

	// first, we check if the next row can be done using the x coords
	// We ignore if it wraps around bc if it wraps around it'd be weird to follow x coords
	if y >= 0 && y < len(c.layout) {
		// the extra 4 is for comfort
		// Think of this scendario:
		// |       |    1   |
		// |     | 2 |   3  |
		//
		// Going 1->2 would be weird

		curFullX := c.layout[cy][cx].x + 4

		for _, ld := range c.layout[y] {
			if !ld.canFocus {
				continue
			}
			if curFullX >= ld.x && curFullX <= ld.x+c.fieldWidth(ld.fieldID) {
				return ld.fieldID
			}
		}
	}

	// So clearly no "perfect" match, so lets just find A match

	for {
		if y < 0 {
			y = len(c.layout) - 1
		} else if y == len(c.layout) {
			y = 0
		}

		for x := 0; x < len(c.layout[y]); x++ {
			f := c.layout[y][x]
			if f.canFocus {
				return f.fieldID
			}
		}
		y += dir
	}
}
