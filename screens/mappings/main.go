package mappings

import (
	"context"
	"log"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bank-data-db/proto/mappings_pb"
	"github.com/bank_data_tui/api"
	"github.com/bank_data_tui/utils"
	"github.com/bank_data_tui/utils/editor"
	"github.com/bank_data_tui/utils/listeditor"
	"github.com/bank_data_tui/utils/repo"
)

var _ listeditor.Delegate[*mappings_pb.Mapping] = &mappingDel{}

type mappingDel struct {
	cache         *repo.Cache
	api           *api.Client
	categoryField *editor.FieldDropdown
	cardField     *editor.FieldDropdown
}

// FilterValue implements [listeditor.Delegate].
func (m *mappingDel) FilterValue(v *mappings_pb.Mapping) string {
	return v.GetName()
}

// InitialFetch implements [listeditor.Delegate].
func (m *mappingDel) InitialFetch() ([]*mappings_pb.Mapping, error) {
	return api.ListAll(func(tok *string) (api.GenericRespList[*mappings_pb.Mapping], error) {
		return m.api.MappingsList(context.Background(), mappings_pb.ReqList_builder{
			PaginationToken: tok,
		}.Build())
	})
}

// NewItem implements [listeditor.Delegate].
func (m *mappingDel) NewItem() *mappings_pb.Mapping {
	return &mappings_pb.Mapping{}
}

func (m *mappingDel) RenderItem(style lipgloss.Style, selected bool, item *mappings_pb.Mapping) string {
	if selected {
		style = style.Underline(true)
	}

	return " " + style.Render(utils.Overflow(item.GetName(), listeditor.WIDTH_LIST-1))
}

type categoriesFetched struct{}
type cardsFetched struct{}

func (m *mappingDel) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			_, err := m.cache.Categories.MaybeLoad(m.api)
			if err != nil {
				log.Panicln(err)
			}

			return categoriesFetched{}
		},
		func() tea.Msg {
			_, err := m.cache.Cards.MaybeLoad(m.api)
			if err != nil {
				log.Panicln(err)
			}

			return cardsFetched{}
		},
	)
}

func (m *mappingDel) Update(msg tea.Msg) tea.Cmd {
	switch msg.(type) {
	case categoriesFetched:
		m.categoryField.SetValues(m.cache.Categories.DropdownValues(false), true)
		return editor.CmdForceReLayout
	case cardsFetched:
		m.cardField.SetValues(m.cache.Cards.DropdownValues(false), true)
		return editor.CmdForceReLayout
	}

	return nil
}

func New(c *api.Client, cache *repo.Cache, w, h int) *listeditor.Model[*mappings_pb.Mapping] {
	m := listeditor.New(
		w, h, "New Mapping", &mappingDel{
			cache: cache,
			api:   c,
		},
	)

	return m
}
