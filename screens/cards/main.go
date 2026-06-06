package cards

import (
	"slices"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bank-data-db/proto/cards_pb"
	"github.com/bank-data-db/tui/api"
	"github.com/bank-data-db/tui/utils"
	"github.com/bank-data-db/tui/utils/editor"
	"github.com/bank-data-db/tui/utils/listeditor"
	"github.com/bank-data-db/tui/utils/repo"
)

var _ listeditor.Delegate[*cards_pb.Card] = &cardsImpl{}

type cardsImpl struct {
	api   *api.Client
	cache *repo.Cache
}

func (m *cardsImpl) InitialFetch() ([]*cards_pb.Card, error) {
	return m.cache.Cards.MaybeLoad(m.api)
}

func (m *cardsImpl) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case editor.MsgItemDel:
		m.cache.Cards.DeleteByID(string(msg))
	case listeditor.ItemNew:
		m.cache.Cards.Data = append(m.cache.Cards.Data, msg.Value.(*cards_pb.Card))
	case listeditor.ItemUpdate:
		card := msg.Value.(*cards_pb.Card)
		i := slices.IndexFunc(m.cache.Cards.Data, func(c *cards_pb.Card) bool {
			return c.GetID() == card.GetID()
		})
		if i == -1 {
			m.cache.Cards.Data = append(m.cache.Cards.Data, card)
		} else {
			m.cache.Cards.Data[i] = card
		}
	}

	return nil
}

func (m *cardsImpl) NewItem() *cards_pb.Card {
	return &cards_pb.Card{}
}

func (m *cardsImpl) RenderItem(s lipgloss.Style, sel bool, card *cards_pb.Card) string {
	return " " + s.Render(utils.Overflow(card.GetName(), listeditor.WIDTH_LIST-1))
}

func (m cardsImpl) FilterValue(card *cards_pb.Card) string {
	return card.GetName()
}

func New(c *api.Client, cache *repo.Cache, w, h int) *listeditor.Model[*cards_pb.Card] {
	return listeditor.New(
		w, h, "New Card",
		&cardsImpl{
			api:   c,
			cache: cache,
		},
	)
}
