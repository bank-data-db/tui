package editor

import (
	"errors"

	tea "charm.land/bubbletea/v2"
	"github.com/bank_data_tui/utils"
	"google.golang.org/protobuf/protoadapt"
)

type ErrRequired struct {
	msg string
}

func (e ErrRequired) Error() string {
	return e.msg
}

type Message interface {
	protoadapt.MessageV2
	GetID() string
	SetID(string)
}

type layoutData struct {
	x int
	// Field ID mapping:
	// >= 0  -> input data
	// < 0   -> buttons
	fieldID  int
	canFocus bool
	grow     bool
}

type EditorMod func(m *Model)

func WithRequireAtLeastOneOf(msg string, fields ...string) EditorMod {
	return func(m *Model) {
		for _, fID := range fields {
			f := m.fieldByID[fID]
			if f == nil {
				panic("RequireAtLeastOneOf: unknown field: " + fID)
			}
		}
		errReq := ErrRequired{msg}

		m.editorValidations = append(m.editorValidations, func(m *Model) {
			has := false
			for _, fID := range fields {
				if m.fieldByID[fID].Value() != "" {
					has = true
					break
				}
			}

			if has {
				var cmpErr = ErrRequired{}
				for _, fID := range fields {
					f := m.fieldByID[fID]
					if errors.As(f.GetErr(), &cmpErr) && cmpErr == errReq {
						f.SetErr(nil)
					}
				}
			} else {
				for _, fID := range fields {
					m.fieldByID[fID].SetErr(errReq)
				}
			}
		})
	}
}

func WithRequireGroup(msg string, fields ...string) EditorMod {
	return func(m *Model) {
		for _, fID := range fields {
			f := m.fieldByID[fID]
			if f == nil {
				panic("RequireGroup: unknown field: " + fID)
			}
		}

		errReq := ErrRequired{msg}

		m.editorValidations = append(m.editorValidations, func(m *Model) {
			hasOne := false
			hasAll := true
			for _, fID := range fields {
				if m.fieldByID[fID].Value() == "" {
					hasAll = false
				} else {
					hasOne = true
				}
			}

			if hasOne == hasAll {
				var cmpErr = ErrRequired{}
				for _, fID := range fields {
					f := m.fieldByID[fID]
					if errors.As(f.GetErr(), &cmpErr) && cmpErr == errReq {
						f.SetErr(nil)
					}
				}
			} else {
				for _, fID := range fields {
					m.fieldByID[fID].SetErr(errReq)
				}
			}

		})
	}
}

func WithAltCreate(msg string, create func() (string, error)) EditorMod {
	return func(m *Model) {
		m.create.altText = msg
		m.create.alt = func() error {
			id, err := create()
			m.item.SetID(id)
			return err
		}
	}
}

func WithAltDelete(msg string, del func() error) EditorMod {
	return func(m *Model) {
		m.del.altText = msg
		m.del.alt = del
	}
}

type action struct {
	altText string
	regular func() error
	alt     func() error
}

func (a action) getF(alt bool) func() error {
	if alt {
		return a.alt
	}
	return a.regular
}

type confirmModal struct {
	text  string
	atYes bool
	cmd   func(m *Model) tea.Cmd
}

type Model struct {
	width  int
	height int

	confirmDial *confirmModal

	item Message

	focusedField int
	fields       []field

	layout     [][]*layoutData
	fieldByID  map[string]inputField
	rowHeights []int

	buttonPad int

	create *action
	update *action
	del    *action

	editorValidations []func(m *Model)
}

func New(
	w int, h int, v Message,
	createFunc func() (string, error),
	updateFunc func() error,
	delFunc func() error,
	layout Layout,
	mods ...EditorMod,
) Model {
	m := Model{
		width:     w,
		height:    h,
		item:      v,
		fields:    []field{},
		fieldByID: map[string]inputField{},
		layout:    [][]*layoutData{},
		create: &action{
			regular: func() error {
				id, err := createFunc()
				v.SetID(id)
				return err
			},
		},
		update: &action{regular: updateFunc},
		del:    &action{regular: delFunc},
	}

	size := [2]int{w, h}
	pr := v.ProtoReflect()
	for i, row := range layout {
		ld := []*layoutData{}
		// +1 for padding
		m.rowHeights = append(m.rowHeights, row.Height()+1)

		j := 0
		for tpl := range row.Items() {
			f, size := tpl.gen(pr, rowCtx{
				pos:  [2]int{i, j},
				size: size,
				row:  row,
			})
			inpF, ok := f.(inputField)
			ld = append(ld, &layoutData{
				fieldID:  len(m.fields),
				canFocus: ok,
				grow:     size == -1,
			})
			if ok {
				inpF.SetFromMsg(pr)
				m.fieldByID[inpF.ID()] = inpF
			}
			valF, ok := f.(validatableField)
			if ok {
				valF.ForceValidate()
			}

			m.fields = append(m.fields, f)
			j++
		}

		m.layout = append(m.layout, ld)
	}

	for _, v := range mods {
		v(&m)
	}

	for _, v := range m.editorValidations {
		v(&m)
	}

	m.focusFirstField()
	if w != 0 {
		m.ResetLayout()
	}

	return m
}

func (m *Model) ResetLayout() {
	hasButtons := len(m.layout) != 0 && m.layout[len(m.layout)-1][0].fieldID < 0

	resize := m.layout
	if hasButtons {
		resize = resize[:len(resize)-1]
	}

	usedHeight := 0

	for i, row := range resize {
		sizeLeft := m.width
		growers := 0
		for _, v := range row {
			if v.grow {
				growers++
			} else {
				sizeLeft -= m.fieldWidth(v.fieldID)
			}
		}

		if growers == 1 && len(row) == 1 {
			f := m.fields[row[0].fieldID]
			f.SetWidth(m.width)
			row[0].x = 0
			m.notifyBounds(f, 0, usedHeight)
		} else if growers != 0 {
			perItemW := sizeLeft / growers
			extraEvery := (sizeLeft % growers) + 1

			curX := 0
			gi := 0
			for _, c := range row {
				c.x = curX
				f := m.fields[c.fieldID]
				m.notifyBounds(f, curX, usedHeight)
				if c.grow {
					w := perItemW
					if extraEvery != 1 && ((gi+1)%extraEvery) == 0 {
						w++
					}
					gi++
					f.SetWidth(w)
				}

				curX += m.fieldWidth(c.fieldID)
			}
		} else if sizeLeft != 0 {
			sizes := []int{}
			for _, v := range row {
				sizes = append(sizes, m.fieldWidth(v.fieldID))
			}

			j := 0
			for off := range utils.EqualSpreadSeq(m.width, sizes) {
				f := m.fields[row[j].fieldID]
				f.SetWidth(sizes[j])
				row[j].x = off
				m.notifyBounds(f, off, usedHeight)
				j++
			}
		}

		usedHeight += m.rowHeights[i]
	}

	saved := m.item.GetID() != ""
	padding := 0
	for p := 2; p >= 0; p-- {
		if btnSizing(saved, p) <= m.width {
			padding = p
			break
		}
	}

	m.resetButtonLayout(m.width, saved, padding)
}

func (m Model) notifyBounds(f field, curX, usedHeight int) {
	b, ok := f.(interestedInBounds)
	if ok {
		b.SetBounds(curX, m.height-usedHeight)
	}
}

func (m Model) FieldByLayout(row, col int) field {
	return m.fields[m.layout[row][col].fieldID]
}

func (m Model) FieldByID(id string) inputField {
	return m.fieldByID[id]
}
