package categories

import (
	"slices"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bank-data-db/proto/categories_pb"
	"github.com/bank_data_tui/api"
	"github.com/bank_data_tui/utils"
	"github.com/bank_data_tui/utils/editor"
	"github.com/bank_data_tui/utils/listeditor"
	"github.com/bank_data_tui/utils/repo"
)

var _ listeditor.Delegate[*categories_pb.Category] = &categoryImpl{}

type categoryImpl struct {
	api *api.Client
	cache *repo.Cache
}

func (m *categoryImpl) InitialFetch() ([]*categories_pb.Category, error) {
	return m.cache.Categories.MaybeLoad(m.api)
}

func (m *categoryImpl) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case editor.MsgItemDel:
		m.cache.Categories.DeleteByID(string(msg))
	case listeditor.ItemNew:
		m.cache.Categories.Data = append(m.cache.Categories.Data, msg.Value.(*categories_pb.Category))
	case listeditor.ItemUpdate:
		cat := msg.Value.(*categories_pb.Category)
		i := slices.IndexFunc(m.cache.Categories.Data, func(c *categories_pb.Category) bool {
			return c.GetID() == cat.GetID()
		})
		if i == -1 {
			m.cache.Categories.Data = append(m.cache.Categories.Data, cat)
		} else {
			m.cache.Categories.Data[i] = cat
		}
	}

	return nil
}

func (m *categoryImpl) NewItem() *categories_pb.Category {
	return &categories_pb.Category{}
}

func (m *categoryImpl) RenderItem(s lipgloss.Style, sel bool, cat *categories_pb.Category) string {
	return " " + utils.RenderCategory(s, listeditor.WIDTH_LIST-1, !sel, cat)
}

func (m categoryImpl) FilterValue(cat *categories_pb.Category) string {
	return cat.GetIcon() + " " + cat.GetName()
}

func New(c *api.Client, cache *repo.Cache, w, h int) *listeditor.Model[*categories_pb.Category] {
	return listeditor.New(
		w, h, "New Category",
		&categoryImpl{
			api:   c,
			cache: cache,
		},
	)
}
