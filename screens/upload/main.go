package upload

import (
	"context"
	"log"
	"os"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bank-data-db/proto/bank_svc_pb"
	"github.com/bank_data_tui/api"
	"github.com/bank_data_tui/styles"
	"github.com/bank_data_tui/utils"
	"github.com/bank_data_tui/utils/dropdown"
	"github.com/bank_data_tui/utils/filepicker"
	"github.com/bank_data_tui/utils/repo"
)

type Model struct {
	api           *api.Client
	cache         *repo.Cache
	cardPicker    dropdown.Model
	filepicker    filepicker.Model
	uploadingPath string
	err           error
	spin          spinner.Model
	w, h          int

	focusedField int
}

const INP_PADDING = 5

type uploaded struct {
	err error
}

func New(api *api.Client, cache *repo.Cache, w, h int) *Model {
	m := &Model{
		api: api, cache: cache,
		w: w, h: h,
		spin: spinner.New(spinner.WithStyle(styles.S_TEXT_HIGHLIGHT)),
		cardPicker: dropdown.New(
			[]*dropdown.Value{}, "Card", h-2,
		),
	}

	fp := filepicker.New(w, h-4, []string{"tsv", "csv"})
	m.filepicker = fp

	m.resetCardVals()
	m.focusOn(0)

	return m
}

func (m *Model) resetCardVals() {
	m.cardPicker.SetValues(m.cache.Cards.DropdownValues(false))
	m.cardPicker.SetWidth(m.w)
}

type cardsFetched struct{}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.filepicker.Init(), func() tea.Msg {
		_, err := m.cache.Cards.MaybeLoad(m.api)
		if err != nil {
			log.Panicln(err)
		}
		return cardsFetched{}
	}, func() tea.Msg {
		return m.focusOn(0)
	})
}

func (m Model) View() (string, *tea.Cursor) {
	canvas := lipgloss.NewCanvas(m.w, m.h)

	if m.uploadingPath == "" {
		cp, cur := m.cardPicker.View()
		fp, fpCur := m.filepicker.View()
		if fpCur != nil {
			cur = fpCur
			cur.Y += 4
		}

		fpL := lipgloss.NewLayer(fp)
		fpL.Y(4)

		comp := lipgloss.NewCompositor(fpL, lipgloss.NewLayer(cp))

		return canvas.Compose(comp).Render(), cur
	}

	var res string
	if m.err != nil {
		res = lipgloss.JoinVertical(
			lipgloss.Center,
			"!! "+styles.S_TEXT_WRONG.Render("Error Uploading")+" !!",
			m.uploadingPath,
			"",
			styles.S_TEXT_WRONG.Render(m.err.Error()),
		)
	} else {
		spin := m.spin.View()
		res = lipgloss.JoinVertical(
			lipgloss.Center,
			spin+" "+styles.S_TEXT_HIGHLIGHT_SECONDARY.Render("Uploading...")+" "+spin,
			"",
			m.uploadingPath,
		)
	}

	res = lipgloss.NewStyle().Width(m.w).AlignHorizontal(lipgloss.Center).Render(res)
	canvas.Compose(lipgloss.NewLayer(res))

	return canvas.Render(), nil
}

func (m Model) Update(msg tea.Msg) (utils.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case utils.ResizeMessage:
		m.w, m.h = msg.W, msg.H
		m.filepicker.SetSize(msg.W, msg.H)
		return m, nil
	case uploaded:
		var cmd tea.Cmd
		if msg.err != nil {
			m.err = msg.err
			cmd = clearErrCMD
		} else {
			cmd = utils.CmdGoToScreen(utils.S_TRANS)
		}

		return m, cmd
	case clearErr:
		m.uploadingPath = ""
		m.err = nil
		return m, nil
	case cardsFetched:
		m.resetCardVals()
	case filepicker.FileSelected:
		m.uploadingPath = msg.Path

		return m, tea.Batch(func() tea.Msg {
			f, err := os.ReadFile(msg.Path)
			if err != nil {
				return uploaded{err: err}
			}

			req := &bank_svc_pb.ReqBankSheet{}
			req.SetBankSheet(f)
			req.SetCardID(m.cardPicker.Value())

			_, err = m.api.UploadBankSheet(context.Background(), req)
			return uploaded{err: err}
		}, m.spin.Tick)
	case tea.KeyPressMsg:
		switch k := msg.String(); k {
		case "shift+up":
			if m.focusedField == 0 {
				cmd := m.focusOn(2)
				return m, cmd
			} else {
				cmd := m.focusOn(m.focusedField - 1)
				return m, cmd
			}
		case "shift+down", "shift+tab":
			if m.focusedField == 2 {
				cmd := m.focusOn(0)
				return m, cmd
			} else {
				cmd := m.focusOn(m.focusedField + 1)
				return m, cmd
			}
		}
	case dropdown.SelectMsg:
		cmd := m.focusOn(1)
		return m, cmd
	}

	if m.uploadingPath == "" {
		// the key press can only go to the focused field, but the rest of the messages should be free flowing
		if _, ok := msg.(tea.KeyPressMsg); ok {
			if m.focusedField == 0 {
				cp, cmd := m.cardPicker.Update(msg)
				m.cardPicker = cp
				return m, cmd
			} else {
				fp, cmd := m.filepicker.Update(msg)
				m.filepicker = fp
				return m, cmd
			}
		} else {
			fp, fpCmd := m.filepicker.Update(msg)
			m.filepicker = fp
			cp, cmd := m.cardPicker.Update(msg)
			m.cardPicker = cp
			return m, tea.Batch(fpCmd, cmd)
		}
	} else {
		spin, cmd := m.spin.Update(m)
		m.spin = spin
		return m, cmd
	}
}

func (m *Model) focusOn(f int) tea.Cmd {
	if m.focusedField == 0 {
		m.cardPicker.Blur()
	}

	m.focusedField = f

	switch f {
	case 0:
		m.filepicker.Blur()
		return m.cardPicker.Focus()
	case 1:
		return m.filepicker.FocusText()
	case 2:
		return m.filepicker.FocusFileArea()
	}

	return nil
}

type clearErr struct{}

func clearErrCMD() tea.Msg {
	<-time.After(5 * time.Second)
	return clearErr{}
}
