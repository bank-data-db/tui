package mappings

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/bank-data-db/proto/mappings_pb"
	"github.com/bank_data_tui/api"
	"github.com/bank_data_tui/styles"
	"github.com/bank_data_tui/utils/dropdown"
	"github.com/bank_data_tui/utils/editor"
	"github.com/bank_data_tui/utils/toast"
	"golang.org/x/exp/constraints"
)

var amtModeMatch = []*dropdown.Value{}

func init() {
	// [display, search, enum]
	v := [][3]any{
		{"=", "==", mappings_pb.AmountMatchModeExact},
		{">", ">", mappings_pb.AmountMatchModeGt},
		{"≥", ">=", mappings_pb.AmountMatchModeGte},
		{"<", "<", mappings_pb.AmountMatchModeLt},
		{"≤", "<=", mappings_pb.AmountMatchModeLte},
	}

	amtModeMatch = make([]*dropdown.Value, len(v))
	bold := lipgloss.NewStyle().Bold(true)

	for i, v := range v {
		d := v[0].(string)
		amtModeMatch[i] = &dropdown.Value{
			Display:     bold.Render(d) + " value",
			Value:       mappings_pb.AmountMatchMode_name[int32(v[2].(mappings_pb.AmountMatchMode))],
			DisplayText: d,
			SearchText:  v[1].(string),
		}
	}
}

func fmtHeading(title string) string {
	return lipgloss.NewStyle().Bold(true).Render(title) + " " + lipgloss.NewStyle().Faint(true).Render("(at least 1 required)")
}

func runNumericToast[T constraints.Integer](tpl string, num T) {
	toast.Success(
		fmt.Sprintf(
			tpl,
			lipgloss.NewStyle().Bold(true).Foreground(styles.COLOR_MAIN).Render(fmt.Sprint(num)),
		),
	)
}

func create(c *api.Client, v *mappings_pb.Mapping) func() (string, error) {
	return func() (string, error) {
		reqNew := &mappings_pb.ReqNew{}
		api.CopyTo(reqNew.ProtoReflect(), v.ProtoReflect(), false)
		resp, err := c.MappingsNew(context.Background(), reqNew)
		if err != nil {
			return "", err
		}

		runNumericToast("Mapped %v Transactions", resp.GetMappedTransactions())

		return resp.GetID(), nil
	}
}

func update(c *api.Client, v *mappings_pb.Mapping) func() error {
	return func() error {
		// I know its a patch style impl, but we can also impl it as a POST >:3
		req := &mappings_pb.ReqUpdate{}
		api.CopyTo(req.ProtoReflect(), v.ProtoReflect(), true, "match_amount_mode", "match_amount")
		if !v.HasMatchAmount() || !v.HasMatchAmountMode() {
			req.SetMatchAmount(mappings_pb.PatchDouble_builder{
				Delete: new(true),
			}.Build())
			req.SetMatchAmountMode(mappings_pb.PatchAmountMode_builder{
				Delete: new(true),
			}.Build())
		} else {
			req.SetMatchAmount(mappings_pb.PatchDouble_builder{
				Value: new(v.GetMatchAmount()),
			}.Build())
			req.SetMatchAmountMode(mappings_pb.PatchAmountMode_builder{
				Value: new(v.GetMatchAmountMode()),
			}.Build())
		}

		log.Println("Map", v)
		log.Println("ReqUp", req)

		_, err := c.MappingsUpdate(context.Background(), req)

		if err == nil {
			toast.Success("Mapping Updated")
		}

		return err
	}
}

func delete(c *api.Client, v *mappings_pb.Mapping, orphan bool) func() error {
	return func() error {
		resp, err := c.MappingDelete(context.Background(), mappings_pb.ReqDelete_builder{
			Id:                 new(v.GetID()),
			OrphanTransactions: new(orphan),
		}.Build())

		if !orphan && err == nil {
			runNumericToast("Mapping Deleted; Un-Mapped %v Transactions", resp.GetAffectedTransactions())
		}

		return err
	}
}

func (c *mappingDel) NewEditor(w, h int, v *mappings_pb.Mapping) editor.Model {
	e := editor.New(
		w, h, v,
		create(c.api, v),
		update(c.api, v),
		delete(c.api, v, false),
		editor.Layout{
			editor.Row{
				editor.TextInput("Name", true, "name", true),
				editor.TextInput("Priority", false, "priority", false, editor.WithTextSize(10)),
			},
			editor.LabelRow{editor.Label(fmtHeading("Matchers")), editor.Spacer()},
			editor.Row{
				editor.TextInput("Text", true, "match_text", false, editor.WithTextValidation(func(s string) *string {
					_, err := regexp.CompilePOSIX(s)
					if err != nil {
						return new("Invalid Regex")
					}

					return nil
				})),
				editor.Label("&&"),
				editor.TextInput("Amount", false, "match_amount", false, editor.WithTextValidation(func(s string) *string {
					parts := strings.SplitN(s, ".", 2)
					if len(parts) < 2 {
						return nil
					}
					if len(parts[1]) > 2 {
						return new("Too precise")
					}
					return nil
				}), editor.WithTextSize(10)),
				editor.Dropdown("Operator", false, "match_amount_mode", amtModeMatch),
				editor.Label("&&"),
				editor.Dropdown("Card", false, "match_card_id", c.cache.Cards.DropdownValues(false)),
			},
			editor.LabelRow{editor.Label(fmtHeading("Result")), editor.Spacer()},
			editor.Row{
				editor.TextInput("Name", true, "result_name", false),
				editor.Dropdown("Category", false, "result_category_id", c.cache.Categories.DropdownValues(false)),
			},
		},
		editor.WithRequireAtLeastOneOf("Matcher Required", "match_text", "match_amount", "match_amount_mode", "match_card_id"),
		editor.WithRequireAtLeastOneOf("Result Required", "result_name", "result_category_id"),
		editor.WithRequireGroup("Need N & OP", "match_amount", "match_amount_mode"),
		editor.WithAltDelete(
			"You are about to delete a mapping but keep it's affects. Are you sure?", delete(c.api, v, true),
		),
	)

	c.categoryField = e.FieldByID("result_category_id").(*editor.FieldDropdown)
	c.cardField = e.FieldByID("match_card_id").(*editor.FieldDropdown)

	return e
}
