package ui

import (
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"id3v2-tui/internal/theme"
)

type SaveCallback func(filePath, trackName, artist, album, coverPath string) (string, error)
type GetRootFunc func() tview.Primitive
type ShowErrorFunc func(msg string)
type ShowMessageFunc func(msg string)
type GetFormFunc func() *tview.Form
type GetFileListFunc func() *tview.List
type GetCurrentDirFunc func() string
type SetCurrentDirFunc func(string)
type GetFocusIndexFunc func() int
type SetFocusIndexFunc func(int)

type UIContext struct {
	App           *tview.Application
	GetRoot       GetRootFunc
	ShowError     ShowErrorFunc
	ShowMessage   ShowMessageFunc
	GetForm       GetFormFunc
	GetFileList   GetFileListFunc
	GetCurrentDir GetCurrentDirFunc
	SetCurrentDir SetCurrentDirFunc
	GetFocusIndex GetFocusIndexFunc
	SetFocusIndex SetFocusIndexFunc
	SaveMetadata  SaveCallback
	CurrentFile   string
}

func CreateFileBrowser(ctx *UIContext) *tview.List {
	list := tview.NewList()
	list.SetBorder(true).SetTitle("Files")
	list.SetMainTextColor(theme.Text)
	list.SetSecondaryTextColor(theme.TextDim)
	list.SetTitleColor(theme.Primary)
	list.SetBorderColor(theme.Primary)
	return list
}

func CreateMetadataForm(directMode bool, ctx *UIContext) *tview.Form {
	form := tview.NewForm()
	form.SetBorder(true).SetTitle("Metadata Editor")
	form.SetTitleColor(theme.Secondary)
	form.SetBorderColor(theme.Primary)
	form.SetLabelColor(theme.TextDim)
	form.SetFieldTextColor(theme.Text)
	form.SetFieldBackgroundColor(theme.Secondary)

	form.AddInputField("Track Name", "", 40, nil, nil)
	form.AddInputField("Artist", "", 40, nil, nil)
	form.AddInputField("Album", "", 40, nil, nil)
	form.AddInputField("Cover Image Path", "", 40, nil, nil)

	form.AddButton("Save", func() {
		var filePath string
		if directMode {
			filePath = ctx.CurrentFile
		} else {
			fileList := ctx.GetFileList()
			selectedItem := fileList.GetCurrentItem()
			if selectedItem >= 0 {
				mainText, _ := fileList.GetItemText(selectedItem)
				if mainText != "" && !strings.HasSuffix(mainText, "/") && mainText != ".." {
					filePath = filepath.Join(ctx.GetCurrentDir(), mainText)
				}
			}
		}

		if filePath == "" {
			return
		}

		trackName := form.GetFormItemByLabel("Track Name").(*tview.InputField).GetText()
		artist := form.GetFormItemByLabel("Artist").(*tview.InputField).GetText()
		album := form.GetFormItemByLabel("Album").(*tview.InputField).GetText()
		coverPath := form.GetFormItemByLabel("Cover Image Path").(*tview.InputField).GetText()

		diff, err := ctx.SaveMetadata(filePath, trackName, artist, album, coverPath)
		if err != nil {
			ctx.ShowError(err.Error())
		} else if diff != "" {
			ctx.ShowMessage("Metadata saved successfully!\n\n" + diff)
		} else {
			ctx.ShowMessage("No changes detected")
		}
	})

	form.AddButton("Clear", func() {
		form.GetFormItemByLabel("Track Name").(*tview.InputField).SetText("")
		form.GetFormItemByLabel("Artist").(*tview.InputField).SetText("")
		form.GetFormItemByLabel("Album").(*tview.InputField).SetText("")
		form.GetFormItemByLabel("Cover Image Path").(*tview.InputField).SetText("")
	})

	form.SetButtonsAlign(tview.AlignCenter)
	form.SetButtonBackgroundColor(theme.Secondary)
	form.SetButtonTextColor(theme.Text)

	return form
}

func CreateInputCapture(directMode bool, ctx *UIContext) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'q' {
			ctx.App.Stop()
			return nil
		}
		if event.Key() == tcell.KeyTab {
			form := ctx.GetForm()
			fileList := ctx.GetFileList()
			numFormItems := form.GetFormItemCount()
			numButtons := form.GetButtonCount()
			totalItems := numFormItems + numButtons

			if !directMode {
				totalItems++
			}

			focusIndex := ctx.GetFocusIndex()
			focusIndex = (focusIndex + 1) % totalItems
			ctx.SetFocusIndex(focusIndex)

			if !directMode && focusIndex == 0 {
				ctx.App.SetFocus(fileList)
			} else {
				idx := focusIndex
				if !directMode {
					idx--
				}
				if idx < numFormItems {
					ctx.App.SetFocus(form.GetFormItem(idx))
				} else {
					ctx.App.SetFocus(form.GetButton(idx - numFormItems))
				}
			}
			return nil
		}
		if event.Key() == tcell.KeyBacktab {
			form := ctx.GetForm()
			fileList := ctx.GetFileList()
			numFormItems := form.GetFormItemCount()
			numButtons := form.GetButtonCount()
			totalItems := numFormItems + numButtons

			if !directMode {
				totalItems++
			}

			focusIndex := ctx.GetFocusIndex()
			focusIndex--
			if focusIndex < 0 {
				focusIndex = totalItems - 1
			}
			ctx.SetFocusIndex(focusIndex)

			if !directMode && focusIndex == 0 {
				ctx.App.SetFocus(fileList)
			} else {
				idx := focusIndex
				if !directMode {
					idx--
				}
				if idx < 0 {
					ctx.App.SetFocus(form.GetButton(numButtons - 1))
				} else if idx < numFormItems {
					ctx.App.SetFocus(form.GetFormItem(idx))
				} else {
					ctx.App.SetFocus(form.GetButton(idx - numFormItems))
				}
			}
			return nil
		}
		if event.Key() == tcell.KeyEsc {
			form := ctx.GetForm()
			form.GetFormItemByLabel("Track Name").(*tview.InputField).SetText("")
			form.GetFormItemByLabel("Artist").(*tview.InputField).SetText("")
			form.GetFormItemByLabel("Album").(*tview.InputField).SetText("")
			form.GetFormItemByLabel("Cover Image Path").(*tview.InputField).SetText("")
			return nil
		}
		return event
	}
}

func PopulateForm(form *tview.Form, trackName, artist, album, coverPath string) {
	form.GetFormItemByLabel("Track Name").(*tview.InputField).SetText(trackName)
	form.GetFormItemByLabel("Artist").(*tview.InputField).SetText(artist)
	form.GetFormItemByLabel("Album").(*tview.InputField).SetText(album)
	form.GetFormItemByLabel("Cover Image Path").(*tview.InputField).SetText(coverPath)
}

func CreateStatusBar(text string) *tview.TextView {
	statusBar := tview.NewTextView().
		SetText(text).
		SetTextColor(theme.TextDim).
		SetTextAlign(tview.AlignCenter)
	statusBar.SetBorder(false)
	return statusBar
}

func CreateFileListWrapper(fileList *tview.List) *tview.Frame {
	title := "[" + theme.HexPrimary + "]Files[" + theme.HexText + "]"
	fileListWrapper := tview.NewFrame(fileList).
		AddText(title, true, tview.AlignLeft, theme.Primary).
		AddText("", false, tview.AlignLeft, theme.Text)
	fileListWrapper.SetBorder(true).SetTitle("Files")
	return fileListWrapper
}

func CreateFormWrapper(form *tview.Form, title string) *tview.Frame {
	coloredTitle := "[" + theme.HexPrimary + "]" + title + "[" + theme.HexText + "]"
	formWrapper := tview.NewFrame(form).
		AddText(" "+coloredTitle+" ", true, tview.AlignLeft, theme.Primary).
		AddText("", false, tview.AlignLeft, theme.Text)
	return formWrapper
}
