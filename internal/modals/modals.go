package modals

import (
	"github.com/rivo/tview"

	"id3v2-tui/internal/theme"
)

func ShowError(app *tview.Application, root tview.Primitive, msg string) {
	modal := tview.NewModal()
	modal.SetText(msg)
	modal.SetTextColor(theme.Error)
	modal.AddButtons([]string{"OK"})
	modal.SetButtonBackgroundColor(theme.Secondary)
	modal.SetButtonTextColor(theme.Text)
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		app.SetRoot(root, false)
	})
	app.SetRoot(modal, false)
}

func ShowMessage(app *tview.Application, root tview.Primitive, msg string) {
	modal := tview.NewModal()
	modal.SetText(msg)
	modal.SetTextColor(theme.Text)
	modal.AddButtons([]string{"OK"})
	modal.SetButtonBackgroundColor(theme.Secondary)
	modal.SetButtonTextColor(theme.Text)
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		app.SetRoot(root, false)
	})
	app.SetRoot(modal, false)
}
