package main

import (
	"log"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/bank-data-db/tui/screens/cards"
	"github.com/bank-data-db/tui/screens/categories"
	"github.com/bank-data-db/tui/screens/login"
	"github.com/bank-data-db/tui/screens/mappings"
	"github.com/bank-data-db/tui/screens/transactions"
	"github.com/bank-data-db/tui/screens/upload"
	"github.com/bank-data-db/tui/utils"
	"github.com/bank-data-db/tui/utils/toast"
)

func (m *mainApp) switchToScreen(s utils.ScreenID) tea.Cmd {
	if m.curFocusedScreen == s {
		return nil
	}

	m.curFocusedScreen = s
	switch s {
	case utils.S_LOGIN:
		m.screenImp = login.NewScreenLogin(m.api)
	case utils.S_TRANS:
		m.screenImp = transactions.New(m.api, m.cache, m.width, m.height-HEADER_HEIGHT)
	case utils.S_MAPPINGS:
		m.screenImp = mappings.New(m.api, m.cache, m.width, m.height-HEADER_HEIGHT)
	case utils.S_CATEGORIES:
		m.screenImp = categories.New(m.api, m.cache, m.width, m.height-HEADER_HEIGHT)
	case utils.S_CARDS:
		m.screenImp = cards.New(m.api, m.cache, m.width, m.height-HEADER_HEIGHT)
	case utils.S_UPLOAD:
		m.screenImp = upload.New(m.api, m.cache, m.width, m.height-HEADER_HEIGHT)
	}

	return m.screenImp.Init()
}

type clearToastMsg struct{}

func (m *mainApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	batcher := []tea.Cmd{}

	passToChildren := false

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "alt+tab":
			s := m.curFocusedScreen + 1
			if s > utils.S_UPLOAD {
				s = utils.S_TRANS
			}
			batcher = append(batcher, m.switchToScreen(s))
		case "alt+shift+tab":
			s := m.curFocusedScreen - 1
			if s == utils.S_LOGIN {
				s = utils.S_UPLOAD
			}
			batcher = append(batcher, m.switchToScreen(s))
		case "alt+t":
			batcher = append(batcher, m.switchToScreen(utils.S_TRANS))
		case "alt+m":
			batcher = append(batcher, m.switchToScreen(utils.S_MAPPINGS))
		case "alt+c":
			batcher = append(batcher, m.switchToScreen(utils.S_CATEGORIES))
		case "alt+a":
			batcher = append(batcher, m.switchToScreen(utils.S_CARDS))
		case "alt+u", "alt+n":
			batcher = append(batcher, m.switchToScreen(utils.S_UPLOAD))
		default:
			passToChildren = true
		}
	case tea.WindowSizeMsg:
		log.Println("RESIZE", msg.Width)
		m.height = msg.Height
		m.width = msg.Width

		m.screenImp, cmd = m.screenImp.Update(utils.ResizeMessage{
			W: m.width,
			H: m.height - HEADER_HEIGHT,
		})
		batcher = append(batcher, cmd)
	case utils.MsgSwitchScreens:
		batcher = append(batcher, m.switchToScreen(utils.ScreenID(msg)))
	case toast.ToastMsg:
		batcher = append(batcher, func() tea.Msg {
			<-time.After(time.Second * 2)

			return clearToastMsg{}
		})
		m.toasts = append(m.toasts, &msg)
	case clearToastMsg:
		if len(m.toasts) != 0 {
			// sanity check tbh
			m.toasts = m.toasts[1:]
		}
	default:
		passToChildren = true
	}

	if passToChildren {
		m.screenImp, cmd = m.screenImp.Update(msg)
		batcher = append(batcher, cmd)
	}

	return m, tea.Batch(batcher...)
}
