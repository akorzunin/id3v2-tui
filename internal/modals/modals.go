package modals

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func ShowError(app *tview.Application, root tview.Primitive, msg string) {
	modal := tview.NewModal()
	modal.SetText(msg)
	modal.SetTextColor(tcell.ColorRed)
	modal.AddButtons([]string{"OK"})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		app.SetRoot(root, false)
	})
	app.SetRoot(modal, false)
}

func ShowMessage(app *tview.Application, root tview.Primitive, msg string) {
	modal := tview.NewModal()
	modal.SetText(msg)
	modal.SetTextColor(tcell.ColorGreen)
	modal.AddButtons([]string{"OK"})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		app.SetRoot(root, false)
	})
	app.SetRoot(modal, false)
}
