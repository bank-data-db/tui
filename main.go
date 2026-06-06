package main

import (
	"fmt"
	"log"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/bank-data-db/tui/api"

	"github.com/bank-data-db/tui/screens/login"
	"github.com/bank-data-db/tui/utils"
	"github.com/bank-data-db/tui/utils/repo"
	"github.com/bank-data-db/tui/utils/toast"
	"github.com/joho/godotenv"
)

type mainApp struct {
	curFocusedScreen utils.ScreenID
	screenImp        utils.Screen

	width  int
	height int

	cache *repo.Cache
	api   *api.Client

	toasts []*toast.ToastMsg
}

func (m mainApp) Init() tea.Cmd {
	cmd := m.screenImp.Init()
	return cmd
}

func main() {
	f, err := os.Create("logs/log.log")
	if err != nil {
		panic(err)
	}
	log.SetOutput(f)
	defer f.Close()
	godotenv.Load()

	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		panic("no API url :( (set API_URL env)")
	}

	api, err := api.NewClient(apiURL)
	if err != nil {
		panic("Is the server down? Err: " + err.Error())
	}

	app := &mainApp{
		curFocusedScreen: utils.S_LOGIN,
		screenImp:        login.NewScreenLogin(api),
		api:              api,
		cache:            repo.NewCache(),
	}
	user, pass := os.Getenv("USERNAME"), os.Getenv("PASSWORD")

	if user != "" && pass != "" {
		err := app.api.Login(user, pass)
		if err != nil {
			panic(err)
		}
		// Slower startup times bc this init will run in the global init
		app.switchToScreen(utils.S_TRANS)
	}

	p := tea.NewProgram(app)

	go func() {
		for {
			msg := <-utils.GlobalMessage
			p.Send(msg)
		}
	}()

	if _, err := p.Run(); err != nil {
		fmt.Println(err)
	}
}
