package transactions

import (
	"context"
	"log"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/bank-data-db/proto/transactions_pb"
	"github.com/bank-data-db/tui/styles"
	"github.com/bank-data-db/tui/utils/dropdown"
	"github.com/bank-data-db/tui/utils/toast"
)

type editRow struct {
	name    textinput.Model
	cat     dropdown.Model
	oldName *string
	// Old cat id
	oldCat *string
}

func (e *editRow) toggleFocus() tea.Cmd {
	if e.cat.Focused() {
		e.cat.Blur()
		return e.name.Focus()
	}

	e.name.Blur()
	return e.cat.Focus()
}

func (m *Model) newEditRow() tea.Cmd {
	nameInput := textinput.New()
	nameInput.SetStyles(textinput.Styles{
		Focused: textinput.StyleState{
			Placeholder: styles.S_TEXT_DISABLED,
		},
		Blurred: textinput.StyleState{
			Text:        styles.S_TEXT_DISABLED,
			Placeholder: styles.S_TEXT_DISABLED,
		},
		Cursor: styles.TI_CURSOR,
	})
	nameInput.SetVirtualCursor(false)
	nameInput.Prompt = ""
	cmd := nameInput.Focus()

	catInput := dropdown.New(
		m.cache.Categories.DropdownValues(true),
		"Category",
		m.h-(m.selected-m.viewportOff),
	)
	v := m.items[m.selected]

	var oldName, oldCat *string

	if v.HasResolvedName() {
		oldName = new(v.GetResolvedName())
		nameInput.SetValue(v.GetResolvedName())
	}
	if v.HasResolvedCategoryID() {
		oldCat = new(v.GetResolvedCategoryID())
		catInput.SetValue(v.GetResolvedCategoryID())
	}

	m.editRow = &editRow{
		name:    nameInput,
		cat:     catInput,
		oldName: oldName,
		oldCat:  oldCat,
	}

	m.resizeEditRow()

	return cmd
}

func (m *Model) resizeEditRow() {
	if m.editRow == nil {
		return
	}

	cols := m.cols()
	m.editRow.name.SetWidth(cols[2] - 1)
	m.editRow.cat.SetWidth(cols[4] + 3 + cols[5] + 2)
	m.updateDropdownHeight()
}

func (m *Model) updateDropdownHeight() {
	if m.editRow == nil {
		return
	}

	m.editRow.cat.MaxHeight = m.h - (m.selected - m.viewportOff - 5)
}

type transUpdated struct{}

func (m *Model) submitEditRow() tea.Msg {
	v := m.items[m.selected]
	_, err := m.api.TransactionsUpdate(context.Background(), transactions_pb.ReqUpdate_builder{
		Id:                 new(v.GetID()),
		ResolvedName:       new(m.editRow.name.Value()),
		ResolvedCategoryId: new(m.editRow.cat.Value()),
	}.Build())
	if err != nil {
		toast.Error("Error Updating Transaction")
		log.Println("Error Updating transaction", err)
		return nil
	}

	toast.Success("Transaction Updated!")

	return transUpdated{}
}

func (m *Model) handleKeyEditMode(msg tea.KeyPressMsg) (bool, tea.Cmd) {
	var ti interface {
		Value() string
		SetValue(string)
		Blur()
	} = &m.editRow.name

	if m.editRow.cat.Focused() {
		ti = &m.editRow.cat
	}

	switch k := msg.String(); k {
	case "esc":
		cur := ti.Value()
		old := m.editRow.oldName
		if m.editRow.cat.Focused() {
			old = m.editRow.oldCat
		}

		if (cur == "" && old == nil) || (old != nil && *old == cur) {
			m.editRow = nil
		} else {
			if old == nil {
				ti.SetValue("")
			} else {
				ti.SetValue(*old)
			}
		}

		return true, nil
	case "alt+esc":
		m.editRow = nil
		return true, nil
	case "enter":
		if !m.editRow.cat.Focused() {
			return true, m.submitEditRow
		}
	case "alt+enter":
		return true, m.submitEditRow
	case "tab", "shift+tab":
		return true, m.editRow.toggleFocus()
	}

	return false, nil
}
