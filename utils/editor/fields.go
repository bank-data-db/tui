package editor

import (
	"iter"
	"slices"

	tea "charm.land/bubbletea/v2"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type field interface {
	Width() int
	View() (string, *tea.Cursor)
	SetWidth(int)
}

type inputField interface {
	field

	ID() string

	// Must return false if should be handled by the editor too
	HandleKey(tea.KeyPressMsg) (bool, tea.Cmd)

	Focus() tea.Cmd
	Blur()

	// Due to interfaces, we MUST have pointers ://
	Update(tea.Msg) tea.Cmd

	Value() string

	SetOnMsg(protoreflect.Message)
	SetFromMsg(protoreflect.Message)

	// For external validation
	SetErr(err error)
	GetErr() error
}

type validatableField interface {
	ForceValidate()
}

type interestedInBounds interface {
	SetBounds(w, h int)
}

type rowCtx struct {
	// indexes of your position within the layout
	pos [2]int
	// size of the editor
	size [2]int
	// parent row ptr
	row RowLike
}

type fieldGen func(msg protoreflect.Message, ctx rowCtx) (field, int)

func (f fieldGen) gen(msg protoreflect.Message, ctx rowCtx) (field, int) {
	return f(msg, ctx)
}

type fieldTemplate interface {
	gen(msg protoreflect.Message, ctx rowCtx) (field, int)
}

type RowLike interface {
	Height() int
	Items() iter.Seq[fieldTemplate]
}

type Row []fieldTemplate

func (r Row) Height() int {
	return 3
}
func (r Row) Items() iter.Seq[fieldTemplate] {
	return slices.Values(r)
}

// A row that only allows labels, but in turn has a smaller height
type LabelRow []labelTemplate

func (r LabelRow) Height() int {
	return 1
}
func (r LabelRow) Items() iter.Seq[fieldTemplate] {
	return func(yield func(fieldTemplate) bool) {
		for _, v := range r {
			if !yield(v) {
				return
			}
		}
	}
}

type Layout []RowLike
