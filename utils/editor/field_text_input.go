package editor

import (
	"errors"
	"log"
	"strconv"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bank-data-db/tui/styles"
	"github.com/bank-data-db/tui/utils"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var _ inputField = &FieldTextInput{}

func (i FieldTextInput) ID() string {
	return i.desc.TextName()
}

func strToProtoVal(s string, f protoreflect.FieldDescriptor) (protoreflect.Value, error) {
	switch f.Kind() {
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		v, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return protoreflect.Value{}, err
		}
		return protoreflect.ValueOfInt32(int32(v)), nil
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return protoreflect.Value{}, err
		}
		return protoreflect.ValueOfInt64(int64(v)), nil
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		v, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			return protoreflect.Value{}, err
		}
		return protoreflect.ValueOfUint32(uint32(v)), nil
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		v, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return protoreflect.Value{}, err
		}
		return protoreflect.ValueOfUint64(uint64(v)), nil
	case protoreflect.FloatKind:
		v, err := strconv.ParseFloat(s, 32)
		if err != nil {
			return protoreflect.Value{}, err
		}
		return protoreflect.ValueOfFloat32(float32(v)), nil
	case protoreflect.DoubleKind:
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return protoreflect.Value{}, err
		}
		return protoreflect.ValueOfFloat64(float64(v)), nil
	case protoreflect.StringKind:
		return protoreflect.ValueOfString(s), nil
	case protoreflect.BytesKind:
		return protoreflect.ValueOfBytes([]byte(s)), nil
	case protoreflect.EnumKind:
		n := f.Enum().Values().ByName(protoreflect.Name(s))
		if n == nil {
			panic("unknown value provided to enum field")
		}
		return protoreflect.ValueOfEnum(n.Number()), nil
	}

	panic("Unsupported Data Type")
}

type FieldTextInput struct {
	textinput.Model

	desc  protoreflect.FieldDescriptor
	title string

	width int

	rowMode bool
	StyleCB func(m *textinput.Model, cur lipgloss.Style) lipgloss.Style
}

func (t *FieldTextInput) Focus() tea.Cmd {
	t.CursorEnd()
	return t.Model.Focus()
}

func (t *FieldTextInput) HandleKey(msg tea.KeyPressMsg) (bool, tea.Cmd) {
	if !t.Focused() {
		return false, nil
	}

	switch k := msg.String(); k {
	case "left", "right":
		pos := t.Position()
		if handleHorizConflict(k == "right", pos, pos == lipgloss.Width(t.Value())) {
			cmd := t.Update(msg)
			return true, cmd
		}
	}

	return false, nil
}

func (t *FieldTextInput) SetWidth(w int) {
	t.width = w
	if !t.rowMode {
		// extra -1 for the bullshit "extra" space at the end
		t.Model.SetWidth(w - 4 - 1)
	}
}

func (t FieldTextInput) Width() int {
	return t.width
}

func (t FieldTextInput) viewRowMode(style lipgloss.Style) (string, *tea.Cursor) {
	err := ""
	if t.Err != nil {
		err = styles.S_TEXT_ERR.Render(t.Err.Error())
	}
	field := style.Render(t.Model.View())

	res, offsets := utils.JoinHorizontalWithSpacer(
		t.width, 1,
		t.title,
		utils.Overflow(
			err,
			t.width-lipgloss.Width(t.title)-lipgloss.Width(field)-2,
		)+" ",
		field,
	)

	c := t.Cursor()
	if c != nil {
		c.X += offsets[len(offsets)-1] + 2
		c.Y += 1
	}

	return res, c
}

func (t FieldTextInput) viewInlineMode(style lipgloss.Style) (string, *tea.Cursor) {
	style = style.BorderTop(false)

	str := utils.RenderHeader(t.width, style.GetBorderTopForeground(), t.title) + "\n"
	if t.Err != nil {
		style = style.BorderBottom(false)
	}

	str += style.Render(t.Model.View())

	if t.Err != nil {
		str += utils.RenderErrFooter(t.width, style.GetBorderTopForeground(), t.Err)
	}

	c := t.Cursor()
	if c != nil {
		c.X += 2
		c.Y += 1
	}

	return str, c
}

func (t FieldTextInput) View() (string, *tea.Cursor) {
	if t.width == 0 {
		return "", nil
	}

	fieldStyle := styles.STYLE_FIELD
	if t.Focused() {
		if t.StyleCB != nil {
			fieldStyle = t.StyleCB(&t.Model, fieldStyle)
		} else {
			fieldStyle = fieldStyle.BorderForeground(styles.COLOR_SECONDARY)
		}
	}

	if t.rowMode {
		return t.viewRowMode(fieldStyle)
	}

	return t.viewInlineMode(fieldStyle)
}

func (t *FieldTextInput) Update(msg tea.Msg) tea.Cmd {
	m, cmd := t.Model.Update(msg)
	t.Model = m

	return cmd
}

func (t FieldTextInput) SetOnMsg(msg protoreflect.Message) {
	if t.Value() == "" {
		log.Println("\tActually, nvm clearing")
		msg.Clear(t.desc)
		return
	}

	v, err := strToProtoVal(t.Value(), t.desc)
	if err != nil {
		panic(err)
	}
	msg.Set(t.desc, v)
}

func (t *FieldTextInput) ForceValidate() {
	if t.Validate != nil {
		t.Err = t.Validate(t.Value())
	}
}

func (t *FieldTextInput) SetErr(err error) {
	t.Err = err
}

func (t *FieldTextInput) GetErr() error {
	return t.Err
}

func (t *FieldTextInput) SetFromMsg(pr protoreflect.Message) {
	if !pr.Has(t.desc) {
		t.SetValue("")
		return
	}

	t.SetValue(pr.Get(t.desc).String())
}

type TextMod func(m *FieldTextInput)

func WithTextValidation(f func(s string) *string) TextMod {
	return func(t *FieldTextInput) {
		og := t.Validate
		t.Validate = func(s string) error {
			if og != nil {
				err := og(s)
				if err != nil {
					return err
				}
			}
			if err := f(s); err != nil {
				return errors.New(*err)
			}
			return nil
		}
	}
}

func WithTextSize(size int) TextMod {
	return func(t *FieldTextInput) {
		t.Model.SetWidth(size)
	}
}

func WithStyleCB(f func(m *textinput.Model, cur lipgloss.Style) lipgloss.Style) TextMod {
	return func(m *FieldTextInput) {
		m.StyleCB = f
	}
}

func newTextInput(title string, required bool, mods []TextMod, msg protoreflect.Message, fieldName string) *FieldTextInput {
	stdInp := textinput.New()
	stdInp.Prompt = ""
	stdInp.Blur()
	stdInp.SetVirtualCursor(false)
	stdInp.SetStyles(textinput.Styles{
		Focused: textinput.StyleState{
			Text:        lipgloss.Style{},
			Placeholder: styles.S_TEXT_DISABLED,
			Suggestion:  styles.S_TEXT_DISABLED,
		},
		Blurred: textinput.StyleState{
			Text:        styles.S_TEXT_DISABLED,
			Placeholder: styles.S_TEXT_DISABLED,
		},
		Cursor: styles.TI_CURSOR,
	})

	stdInp.SetWidth(15)

	if required {
		stdInp.Validate = func(s string) error {
			if s == "" {
				return ErrRequired{"Required"}
			}
			return nil
		}
	}

	field := msg.Descriptor().Fields().ByTextName(fieldName)
	if field == nil {
		panic("unknown field: " + fieldName)
	}

	m := &FieldTextInput{
		Model: stdInp,
		desc:  field,
		title: title,
	}

	for _, mod := range mods {
		mod(m)
	}

	m.width = m.Model.Width() + 4 + 1

	switch field.Kind() {
	case protoreflect.BoolKind:
		panic("Don't use text input on a bool")
	case protoreflect.Int32Kind, protoreflect.Sint32Kind,
		protoreflect.Uint32Kind, protoreflect.Int64Kind,
		protoreflect.Sint64Kind, protoreflect.Uint64Kind,
		protoreflect.Sfixed32Kind, protoreflect.Fixed32Kind,
		protoreflect.Sfixed64Kind, protoreflect.Fixed64Kind:
		WithTextValidation(func(s string) *string {
			if s == "" {
				return nil
			}
			_, err := strconv.Atoi(s)
			if err != nil {
				return new("Not a number")
			}
			return nil
		})(m)
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		WithTextValidation(func(s string) *string {
			if s == "" {
				return nil
			}
			_, err := strconv.ParseFloat(s, 32)
			if err != nil {
				return new("Not a decimal")
			}
			return nil
		})(m)
	}

	return m
}

func TextInput(title string, grow bool, protoField string, required bool, mods ...TextMod) fieldTemplate {
	return fieldGen(func(msg protoreflect.Message, _ rowCtx) (field, int) {
		m := newTextInput(title, required, mods, msg, protoField)
		size := m.Width()
		if grow {
			size = -1
		}

		m.rowMode = false

		return m, size
	})
}

func RowTextInput(title string, protoField string, required bool, mods ...TextMod) Row {
	return Row{
		fieldGen(func(msg protoreflect.Message, _ rowCtx) (field, int) {
			m := newTextInput(title, required, mods, msg, protoField)
			m.rowMode = true

			return m, -1
		}),
	}
}
