package editor

import (
	"errors"
	"log"

	tea "charm.land/bubbletea/v2"
	"github.com/bank-data-db/tui/api"
)

func (c Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	batcher := []tea.Cmd{}

	passToChildren := true
	forcePassToChildren := false

	if c.width != 0 {
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			if !c.inButtons(c.focusedField) {
				f := c.fields[c.focusedField]
				inpF, ok := f.(inputField)
				if ok {
					doNotContinue, cmd := inpF.HandleKey(msg)
					batcher = append(batcher, cmd)
					if doNotContinue {
						passToChildren = false
						break
					} else if cmd != nil {
						// Tired lmao
						forcePassToChildren = true
					}
				}
			}

			switch msg.String() {
			case "tab", "shift+tab", "left", "right", "up", "down":
				passToChildren = false

				if c.confirmDial != nil {
					c.confirmDial.atYes = !c.confirmDial.atYes
				} else {
					batcher = append(batcher, c.focusField(c.handleNavKey(msg.String())))
				}
			case "enter", "alt+enter":
				passToChildren = false

				if c.confirmDial != nil {
					if c.confirmDial.atYes {
						batcher = append(batcher, c.confirmDial.cmd(&c))
						c.confirmDial = nil
					} else {
						c.confirmDial = nil
					}
				} else {
					alt := msg.Mod.Contains(tea.ModAlt)

					switch c.focusedField {
					case BTN_SAVE_ID:
						batcher = append(batcher, c.handleSaveEnter(alt))
					case BTN_DEL_ID:
						if !alt || c.del.alt != nil {
							text := "You are about to delete this, are you sure?"
							if alt {
								text = c.del.altText
							}

							c.confirmDial = &confirmModal{
								text: text,
								cmd: func(m *Model) tea.Cmd {
									return m.delete(alt)
								},
							}
						}
					case BTN_RESET_ID:
						c.focusFirstField()
						pr := c.item.ProtoReflect()
						for _, f := range c.fields {
							if f, ok := f.(inputField); ok {
								f.SetFromMsg(pr)
							}
						}
					default:
						batcher = append(batcher, c.focusField(c.handleNavKey("enter")))
					}
				}
			}
		case validationErrMsg:
			for _, v := range msg {
				c.fieldByID[v[0]].SetErr(errors.New(v[1]))
				log.Println("Set err on", v[0], v[1])
			}
		case forceReLayout, MsgItemNew:
			c.ResetLayout()
		}
	}

	if passToChildren || forcePassToChildren {
		updated := false
		for _, f := range c.fieldByID {
			v := f.Value()
			batcher = append(batcher, f.Update(msg))
			if v != f.Value() {
				updated = true
			}
		}
		if updated {
			for _, f := range c.fields {
				if f, ok := f.(validatableField); ok {
					f.ForceValidate()
				}
			}
		}
	}

	for _, val := range c.editorValidations {
		val(&c)
	}

	return c, tea.Batch(batcher...)
}

func (m *Model) Resize(w, h int) {
	m.width, m.height = w, h
	m.ResetLayout()
}

type MsgItemNew string
type MsgItemUpdate string
type MsgItemDel string

type validationErrMsg [][2]string

func newValidationErrMsg(err *api.ValidationErr) validationErrMsg {
	msg := make(validationErrMsg, 0, len(err.Errors))
	for _, v := range err.Errors {
		errMsg := v.GetMessage()
		for _, f := range v.GetFields() {
			msg = append(msg, [2]string{f, errMsg})
		}
	}

	return msg
}

func (m *Model) handleSaveEnter(alt bool) tea.Cmd {
	if !m.valid() {
		return nil
	}

	create := m.item.GetID() == ""
	if alt {
		if create && m.create.alt == nil {
			return nil
		} else if !create && m.update.alt == nil {
			return nil
		}
	}

	if alt {
		m.confirmDial = &confirmModal{
			text:  m.create.altText,
			atYes: false,
			cmd: func(m *Model) tea.Cmd {
				return func() tea.Msg {
					return m.save(alt)
				}
			},
		}
		return nil
	}

	return func() tea.Msg {
		return m.save(alt)
	}
}
