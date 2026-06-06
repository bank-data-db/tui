package cards

import (
	"context"

	"github.com/bank-data-db/proto/bank_svc_pb"
	"github.com/bank-data-db/proto/cards_pb"
	"github.com/bank_data_tui/api"
	"github.com/bank_data_tui/utils/editor"
	"github.com/bank_data_tui/utils/toast"
)

func create(c *api.Client, v *cards_pb.Card) func() (string, error) {
	return func() (string, error) {
		reqNew := &cards_pb.ReqNew{}
		api.CopyTo(reqNew.ProtoReflect(), v.ProtoReflect(), false)
		resp, err := c.CardsNew(context.Background(), reqNew)
		if err != nil {
			return "", err
		}

		toast.Success("Card Created")

		return resp.GetID(), nil
	}
}

func update(c *api.Client, v *cards_pb.Card) func() error {
	return func() error {
		// I know its a patch style impl, but we can also impl it as a POST >:3
		_, err := c.CardsUpdate(context.Background(), v)

		toast.Success("Card Updated")

		return err
	}
}

func delete(c *api.Client, v *cards_pb.Card) func() error {
	return func() error {
		_, err := c.CardDelete(context.Background(), bank_svc_pb.ReqDelete_builder{Id: new(v.GetID())}.Build())

		if err == nil {
			toast.Success("Card Deleted!")
		}

		return err
	}
}

func (c *cardsImpl) NewEditor(w, h int, v *cards_pb.Card) editor.Model {
	return editor.New(
		w, h, v,
		create(c.api, v),
		update(c.api, v),
		delete(c.api, v),
		editor.Layout{
			editor.RowTextInput("Name", "name", true),
		},
	)
}
