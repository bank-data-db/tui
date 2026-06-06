package editor

import (
	"log"

	tea "charm.land/bubbletea/v2"
	"github.com/bank-data-db/tui/utils/dropdown"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var _ inputField = &FieldDropdown{}

func (i FieldDropdown) ID() string {
	return i.desc.TextName()
}

type FieldDropdown struct {
	dropdown.Model

	required bool

	msg  protoreflect.Message
	desc protoreflect.FieldDescriptor
}

func (i *FieldDropdown) HandleKey(msg tea.KeyPressMsg) (bool, tea.Cmd) {
	if !i.Focused() {
		return false, nil
	}
	switch msg.String() {
	case "up", "down":
		m, cmd := i.Model.Update(msg)
		i.Model = m
		return true, cmd
	case "left", "right":
		if !handleHorizConflict(msg.String() == "right", i.InputCursorPosition(), i.InputCursorAtEnd()) {
			return false, nil
		}
		m, cmd := i.Model.Update(msg)
		i.Model = m
		return true, cmd
	case "enter":
		m, cmd := i.Model.Update(msg)
		i.Model = m
		// Forces a command to be returned
		return false, tea.Batch(cmd, func() tea.Msg { return nil })
	}

	return false, nil
}

func (i *FieldDropdown) SetValues(vals []*dropdown.Value, setFromMsg bool) {
	if !i.required {
		vals = append([]*dropdown.Value{dropdown.EmptyValue}, vals...)
	}
	i.Model.SetValues(vals)
	if setFromMsg {
		i.SetFromMsg(i.msg)
	}
}

// SetFromMsg implements [inputField].
func (i *FieldDropdown) SetFromMsg(m protoreflect.Message) {
	if !m.Has(i.desc) {
		i.SetValue("")
		return
	}

	v := m.Get(i.desc)
	if i.desc.Kind() == protoreflect.EnumKind {
		i.SetValue(
			string(i.desc.Enum().Values().ByNumber(v.Enum()).Name()),
		)
	} else {
		i.SetValue(v.String())
	}
}

func (i *FieldDropdown) SetOnMsg(msg protoreflect.Message) {
	log.Println("Setting dropdown val", i.desc.Name(), i.Value())
	if i.Value() == "" {
		log.Println("\tnvm, clearing")
		msg.Clear(i.desc)
		return
	}

	v, _ := strToProtoVal(i.Value(), i.desc)
	log.Println("\tFr", v)

	msg.Set(i.desc, v)
}

func (i *FieldDropdown) SetWidth(w int) {
	if i.Width() != w {
		panic("Can't set the width to a diff value >:(")
	}
}

// Update implements [inputField].
func (i *FieldDropdown) Update(msg tea.Msg) tea.Cmd {
	m, cmd := i.Model.Update(msg)
	i.Model = m
	return cmd
}

func (i *FieldDropdown) SetBounds(_, h int) {
	i.MaxHeight = h
}

func (i *FieldDropdown) GetErr() error {
	return i.Err()
}

func Dropdown(title string, required bool, protoField string, values []*dropdown.Value) fieldTemplate {
	return fieldGen(func(msg protoreflect.Message, _ rowCtx) (field, int) {
		field := msg.Descriptor().Fields().ByTextName(protoField)
		if field == nil {
			panic("unknown field: " + protoField)
		}

		if !required {
			values = append([]*dropdown.Value{dropdown.EmptyValue}, values...)
		}

		d := &FieldDropdown{
			Model:    dropdown.New(values, title, 0),
			desc:     field,
			required: required,
			msg:      msg,
		}
		return d, d.Width()
	})
}
