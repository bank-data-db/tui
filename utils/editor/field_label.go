package editor

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bank_data_tui/utils"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type label struct {
	w      int
	prefix string
	text   string
}

func (l *label) SetWidth(w int) {
	l.w = w
}

// View implements [field].
func (l label) View() (string, *tea.Cursor) {
	return l.prefix + utils.Overflow(l.text, l.w), nil
}

// Width implements [field].
func (l label) Width() int {
	return l.w
}

type labelTemplate interface {
	fieldTemplate
	isALabel()
}

type labelGen = fieldGen

func (t labelGen) isALabel() {}

func Label(text string) labelTemplate {
	return labelGen(func(msg protoreflect.Message, ctx rowCtx) (field, int) {
		l := label{text: text, w: lipgloss.Width(text)}
		l.prefix += strings.Repeat("\n", (ctx.row.Height()-1)/2)
		return &l, l.w
	})
}

type spacer struct {
	w int
}

func (s *spacer) SetWidth(w int) {
	s.w = w
}

func (s *spacer) View() (string, *tea.Cursor) {
	return strings.Repeat(" ", s.w), nil
}

func (s *spacer) Width() int {
	return s.w
}

// Make a growth label that just grows and does shit all else
func Spacer() labelTemplate {
	return labelGen(func(msg protoreflect.Message, ctx rowCtx) (field, int) {
		return &spacer{}, -1
	})
}
