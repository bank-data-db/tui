package editor

import (
	"errors"
	"log"

	tea "charm.land/bubbletea/v2"
	"github.com/bank-data-db/tui/api"
	"github.com/bank-data-db/tui/utils/toast"
)

func (m *Model) save(alt bool) tea.Msg {
	create := m.item.GetID() == ""
	var f func() error
	if create {
		f = m.create.getF(alt)
	} else {
		f = m.update.getF(alt)
	}

	pr := m.item.ProtoReflect()
	for _, v := range m.fieldByID {
		v.SetOnMsg(pr)
	}

	err := f()
	if err != nil {
		if valErr, ok := errors.AsType[*api.ValidationErr](err); ok {
			return newValidationErrMsg(valErr)
		} else {
			toast.Error("Failed to save")
			log.Println(err)
			return nil
		}
	}

	if create {
		return MsgItemNew(m.item.GetID())
	} else {
		return MsgItemUpdate(m.item.GetID())
	}
}

func (m *Model) delete(alt bool) tea.Cmd {
	return func() tea.Msg {
		err := m.del.getF(alt)()
		if err != nil {
			toast.Error("Failed to delete")
			log.Println(err)
			return nil
		}

		return MsgItemDel(m.item.GetID())
	}
}
