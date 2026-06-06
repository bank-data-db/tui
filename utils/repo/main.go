package repo

import (
	"context"
	"slices"

	"charm.land/lipgloss/v2"
	"github.com/bank-data-db/proto/cards_pb"
	"github.com/bank-data-db/proto/categories_pb"
	"github.com/bank-data-db/tui/api"
	"github.com/bank-data-db/tui/utils"
	"github.com/bank-data-db/tui/utils/dropdown"
)

func NewCache() *Cache {
	return &Cache{
		Categories: &cacher[categories_pb.Category, *categories_pb.Category]{
			fetch: func(c *api.Client, tok *string) (api.GenericRespList[*categories_pb.Category], error) {
				return c.CategoriesList(context.Background(), categories_pb.ReqList_builder{
					PageSize:        new(uint32(100)),
					PaginationToken: tok,
				}.Build())
			},
			dropdown: func(c *categories_pb.Category) *dropdown.Value {
				return &dropdown.Value{
					Display: utils.RenderCategory(
						lipgloss.NewStyle(),
						-1,
						true, c,
					),
					Value:       c.GetID(),
					DisplayText: c.GetName(),
					SearchText:  "[" + c.GetIcon() + "] " + c.GetName(),
				}
			},
		},
		Cards: &cacher[cards_pb.Card, *cards_pb.Card]{
			fetch: func(c *api.Client, tok *string) (api.GenericRespList[*cards_pb.Card], error) {
				return c.CardsList(context.Background(), cards_pb.ReqList_builder{
					PageSize:        new(uint32(100)),
					PaginationToken: tok,
				}.Build())
			},
			dropdown: func(c *cards_pb.Card) *dropdown.Value {
				return &dropdown.Value{
					Display: c.GetName(),
					Value:   c.GetID(),
				}
			},
		},
	}
}

type Cache struct {
	Categories *cacher[categories_pb.Category, *categories_pb.Category]
	Cards      *cacher[cards_pb.Card, *cards_pb.Card]
}

type cacher[T any, PT interface {
	*T
	GetID() string
}] struct {
	Data     []PT
	fetch    func(c *api.Client, tok *string) (api.GenericRespList[PT], error)
	dropdown func(PT) *dropdown.Value
}

func (d *cacher[T, PT]) DeleteByID(id string) {
	for i, v := range d.Data {
		if v.GetID() == id {
			d.Data = slices.Delete(d.Data, i, i+1)
			return
		}
	}
}

func (d *cacher[T, PT]) DropdownValues(includeNone bool) []*dropdown.Value {
	l := len(d.Data)
	if includeNone {
		l++
	}
	arr := make([]*dropdown.Value, l)
	if includeNone {
		arr[0] = dropdown.EmptyValue
	}
	for i, v := range d.Data {
		tI := i
		if includeNone {
			tI++
		}
		arr[tI] = d.dropdown(v)
	}

	return arr
}

func (d *cacher[T, PT]) MaybeLoad(c *api.Client) ([]PT, error) {
	if d.Data == nil {
		return d.Refetch(c)
	}

	return d.Data, nil
}

func (d *cacher[T, PT]) Refetch(c *api.Client) ([]PT, error) {
	res, err := api.ListAll(func(tok *string) (api.GenericRespList[PT], error) {
		return d.fetch(c, tok)
	})
	if err != nil {
		return nil, err
	}

	d.Data = res
	return d.Data, nil
}

func (s *cacher[T, PT]) ByID(id string) PT {
	for _, v := range s.Data {
		if v.GetID() == id {
			return v
		}
	}

	return nil
}

func (s *cacher[T, PT]) Easy(c *api.Client) ([]PT, error) {
	if s.Data == nil {
		return s.Refetch(c)
	}
	return s.Data, nil
}

func (s *cacher[T, PT]) EasyByID(c *api.Client, id string) (PT, error) {
	if s.Data == nil {
		_, err := s.Refetch(c)
		if err != nil {
			return nil, err
		}
	}

	return s.ByID(id), nil
}
