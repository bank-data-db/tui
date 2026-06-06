package categories

import (
	"context"
	"strconv"

	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
	"github.com/bank-data-db/proto/bank_svc_pb"
	"github.com/bank-data-db/proto/categories_pb"
	"github.com/bank_data_tui/api"
	"github.com/bank_data_tui/utils/editor"
	"github.com/bank_data_tui/utils/toast"
	"github.com/charmbracelet/x/ansi"
	"github.com/rivo/uniseg"
)

func verifyColor(s string) *string {
	if len(s) != 6 {
		return new("Needs a hex color (no #)")
	}
	if _, err := strconv.ParseUint(s, 16, 64); err != nil {
		return new("Not a valid color")
	}

	return nil
}

func create(c *api.Client, v *categories_pb.Category) func() (string, error) {
	return func() (string, error) {
		reqNew := &categories_pb.ReqNew{}
		api.CopyTo(reqNew.ProtoReflect(), v.ProtoReflect(), false)
		resp, err := c.CategoriesNew(context.Background(), reqNew)
		if err != nil {
			return "", err
		}

		toast.Success("Category Created")

		return resp.GetID(), nil
	}
}

func update(c *api.Client, v *categories_pb.Category) func() error {
	return func() error {
		// I know its a patch style impl, but we can also impl it as a POST >:3
		_, err := c.CategoriesUpdate(context.Background(), v)

		toast.Success("Category Updated")

		return err
	}
}

func delete(c *api.Client, v *categories_pb.Category) func() error {
	return func() error {
		_, err := c.CategoriesDelete(context.Background(), bank_svc_pb.ReqDelete_builder{Id: new(v.GetID())}.Build())

		if err == nil {
			toast.Success("Category Deleted!")
		}

		return err
	}
}

func (c *categoryImpl) NewEditor(w, h int, v *categories_pb.Category) editor.Model {
	return editor.New(
		w, h, v,
		create(c.api, v),
		update(c.api, v),
		delete(c.api, v),
		editor.Layout{
			editor.RowTextInput("Name", "name", true),
			editor.RowTextInput(
				"Color", "color", true,
				editor.WithTextValidation(verifyColor),
				editor.WithStyleCB(func(m *textinput.Model, cur lipgloss.Style) lipgloss.Style {
					if verifyColor(m.Value()) != nil {
						return cur
					}
					return cur.Border(lipgloss.BlockBorder()).BorderForeground(lipgloss.Color("#" + m.Value()))
				}),
			),
			editor.RowTextInput("Icon", "icon", true, editor.WithTextValidation(func(s string) *string {
				if uniseg.GraphemeClusterCount(ansi.Strip(s)) != 1 || lipgloss.Width(s) != 1 {
					return new("Must be 1 character")
				}

				return nil
			})),
		},
	)
}
