package utils

import (
	"image/color"
	"iter"
	"slices"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bank-data-db/proto/categories_pb"
	"github.com/bank-data-db/tui/styles"
	"github.com/charmbracelet/x/ansi"
)

func JoinHorizontal2(w int, a, b string) string {
	res, _ := JoinHorizontalWithSpacer(w, 1, a, b)
	return res
}

func JoinHorizontalWithSpacer(w, spacerIndex int, str ...string) (string, []int) {
	offsets := make([]int, len(str))
	widths := 0
	for _, s := range str {
		widths += lipgloss.Width(s)
	}

	str = slices.Insert(str, spacerIndex, strings.Repeat(" ", w-widths))
	off := 0
	for i, v := range str {
		if spacerIndex == i {
			off += lipgloss.Width(v)
			continue
		} else if i > spacerIndex {
			i--
		}

		offsets[i] = off
		off += lipgloss.Width(v)
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, str...), offsets
}

// Returns [offset, spacer size]
func EqualSpreadSeq(w int, sizes []int) iter.Seq2[int, int] {
	if len(sizes) == 0 {
		return func(yield func(int, int) bool) {
			yield(0, 0)
		}
	} else if len(sizes) == 1 {
		leftover := w - sizes[0]
		return func(yield func(int, int) bool) {
			off := leftover / 2
			yield(off, off)
		}
	}

	widths := 0
	for _, s := range sizes {
		widths += s
	}

	leftover := w - widths
	perSlice := leftover / (len(sizes) - 1)
	addExtraSpaceEvery := (leftover % len(sizes)) + 1
	curOff := 0

	return func(yield func(int, int) bool) {
		for i, size := range sizes {
			spacerSize := perSlice
			if addExtraSpaceEvery != 1 && (i+1)%addExtraSpaceEvery == 0 {
				spacerSize++
			}

			if !yield(curOff, spacerSize) {
				return
			}

			curOff += size + spacerSize
			i++
		}
	}
}

// Returns a rendered str, and offsets from the left
func JoinHorizontalEqualSpread(w int, str ...string) (string, []int) {
	if len(str) == 0 {
		return "", nil
	} else if len(str) == 1 {
		usedW := lipgloss.Width(str[0])
		leftover := w - usedW

		return lipgloss.NewStyle().PaddingLeft(leftover / 2).PaddingRight(leftover - leftover/2).Render(str[0]), []int{leftover / 2}
	}

	widths := make([]int, len(str))
	for i, v := range str {
		widths[i] = lipgloss.Width(v)
	}

	res := make([]string, 0, len(str)+len(str)-1)
	offsets := make([]int, len(str))
	i := 0
	for off, spaceSize := range EqualSpreadSeq(w, widths) {
		offsets[i] = off
		res = append(res, str[i])
		res = append(res, strings.Repeat(" ", spaceSize))

		i++
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, res...), offsets
}

// Reports if any in sl is true
func Any[T any](sl iter.Seq[T], cond func(T) bool) bool {
	for v := range sl {
		if cond(v) {
			return true
		}
	}

	return false
}

// Reports if all in sl is true
func All[T any](sl iter.Seq[T], cond func(T) bool) bool {
	for v := range sl {
		if !cond(v) {
			return false
		}
	}

	return true
}

func Overflow(str string, maxWidth int) string {
	if lipgloss.Width(str) <= maxWidth {
		return str
	}

	lines := strings.Split(str, "\n")
	for i := range lines {
		lines[i] = ansi.Truncate(lines[i], maxWidth, "…")
	}

	return strings.Join(lines, "\n")
}

type ResizeMessage struct {
	W, H int
}

type Screen interface {
	Update(msg tea.Msg) (Screen, tea.Cmd)
	View() (string, *tea.Cursor)
	Init() tea.Cmd
}

type ScreenID int

const (
	S_LOGIN ScreenID = iota
	S_TRANS
	S_MAPPINGS
	S_CATEGORIES
	S_CARDS
	S_UPLOAD
)

type MsgSwitchScreens ScreenID

func GoToScreen(s ScreenID) {
	GlobalMessage <- MsgSwitchScreens(s)
}

func CmdGoToScreen(s ScreenID) tea.Cmd {
	return func() tea.Msg {
		return MsgSwitchScreens(s)
	}
}

// USE SPARINGLY
var GlobalMessage = make(chan tea.Msg)

func RenderCategory(s lipgloss.Style, w int, color bool, c *categories_pb.Category) string {
	if color {
		_, err := strconv.ParseUint(c.GetColor(), 16, 32)

		if err == nil {
			s = s.Foreground(lipgloss.Color("#" + c.GetColor()))
		}
	}

	res := "[" + c.GetIcon() + "] " + c.GetName()
	if w != -1 {
		res = Overflow(res, w)
	}
	return s.Render(res)
}

func RenderHeader(w int, color color.Color, text string) string {
	c := lipgloss.NewStyle().Foreground(color)
	w -= 6
	t := Overflow(text, w)
	return c.Render("╔╡ ") + t + c.Render(" ╞"+strings.Repeat("═", w-lipgloss.Width(t))+"╗")
}

func RenderFooter(w int, color color.Color, text string) string {
	left, right := "", ""
	if w >= 6 {
		left, right = "╡", "╞"
		w -= 2
	}

	w -= 4

	c := lipgloss.NewStyle().Foreground(color)
	t := Overflow(text, w)
	return "\n" + c.Render("╚"+left+" ") + t + c.Render(" "+right+strings.Repeat("═", w-lipgloss.Width(t))+"╝")
}

func RenderErrFooter(w int, color color.Color, err error) string {
	return RenderFooter(
		w, color, styles.S_TEXT_ERR.Render(err.Error()),
	)
}
