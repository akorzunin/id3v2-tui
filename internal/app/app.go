package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rivo/tview"

	"id3v2-tui/internal/files"
	"id3v2-tui/internal/metadata"
	"id3v2-tui/internal/modals"
	"id3v2-tui/internal/ui"
)

type App struct {
	app         *tview.Application
	fileList    *tview.List
	form        *tview.Form
	pages       *tview.Pages
	root        tview.Primitive
	meta        *metadata.Metadata
	currentDir  string
	currentFile string
	focusIndex  int
}

func NewApp() *App {
	return &App{
		meta: &metadata.Metadata{},
	}
}

func (a *App) getRoot() tview.Primitive {
	return a.root
}

func (a *App) showError(msg string) {
	modals.ShowError(a.app, a.root, msg)
}

func (a *App) showMessage(msg string) {
	modals.ShowMessage(a.app, a.root, msg)
}

func (a *App) getForm() *tview.Form {
	return a.form
}

func (a *App) getFileList() *tview.List {
	return a.fileList
}

func (a *App) getCurrentDir() string {
	return a.currentDir
}

func (a *App) setCurrentDir(dir string) {
	a.currentDir = dir
}

func (a *App) getFocusIndex() int {
	return a.focusIndex
}

func (a *App) setFocusIndex(idx int) {
	a.focusIndex = idx
}

func (a *App) saveMetadata(filePath, trackName, artist, album, coverPath string) error {
	a.meta.TrackName = trackName
	a.meta.Artist = artist
	a.meta.Album = album
	a.meta.CoverPath = coverPath
	return metadata.Save(filePath, a.meta)
}

func (a *App) readMetadata(filePath string) error {
	meta, err := metadata.Read(filePath)
	if err != nil {
		a.meta = &metadata.Metadata{}
		return nil
	}
	a.meta = meta
	return nil
}

func (a *App) loadFiles(dir string) {
	a.currentDir = files.Load(a.fileList, dir)
}

func (a *App) Run(filePath string) error {
	a.app = tview.NewApplication()

	if filePath != "" {
		return a.runDirectEdit(filePath)
	}

	ctx := &ui.UIContext{
		App:           a.app,
		GetRoot:       a.getRoot,
		ShowError:     a.showError,
		ShowMessage:   a.showMessage,
		GetForm:       a.getForm,
		GetFileList:   a.getFileList,
		GetCurrentDir: a.getCurrentDir,
		SetCurrentDir: a.setCurrentDir,
		GetFocusIndex: a.getFocusIndex,
		SetFocusIndex: a.setFocusIndex,
		SaveMetadata:  a.saveMetadata,
	}

	a.fileList = ui.CreateFileBrowser(ctx)
	a.form = ui.CreateMetadataForm(false, ctx)

	currentDir, _ := os.Getwd()
	a.currentDir = currentDir
	a.loadFiles(currentDir)

	statusBar := ui.CreateStatusBar("↑↓ Navigate | Enter: Open | Tab/Shift+Tab: Cycle | Esc: Clear | q: Quit")
	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(a.fileList, 0, 1, true).
			AddItem(a.form, 0, 2, false), 0, 1, true).
		AddItem(statusBar, 1, 0, false)

	a.fileList.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		if files.IsDirectoryEntry(mainText) {
			newDir := files.ResolveDirectory(a.currentDir, mainText)
			if _, err := os.Stat(newDir); err == nil {
				a.loadFiles(newDir)
			}
			return
		}

		selectedPath := filepath.Join(a.currentDir, mainText)
		err := a.readMetadata(selectedPath)
		if err != nil {
			a.showError(err.Error())
			return
		}

		ui.PopulateForm(a.form, a.meta.TrackName, a.meta.Artist, a.meta.Album, a.meta.CoverPath)

		a.focusIndex = 1
		a.app.SetFocus(a.form.GetFormItem(0))
	})

	mainFlex.SetInputCapture(ui.CreateInputCapture(false, ctx))

	a.root = mainFlex
	a.app.SetRoot(mainFlex, true)
	a.app.SetFocus(a.fileList)

	return a.app.Run()
}

func (a *App) runDirectEdit(filePath string) error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("cannot access file: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file")
	}

	a.currentFile = absPath
	a.currentDir = filepath.Dir(absPath)

	a.readMetadata(absPath)

	ctx := &ui.UIContext{
		App:           a.app,
		GetRoot:       a.getRoot,
		ShowError:     a.showError,
		ShowMessage:   a.showMessage,
		GetForm:       a.getForm,
		GetFileList:   a.getFileList,
		GetCurrentDir: a.getCurrentDir,
		SetCurrentDir: a.setCurrentDir,
		GetFocusIndex: a.getFocusIndex,
		SetFocusIndex: a.setFocusIndex,
		SaveMetadata:  a.saveMetadata,
		CurrentFile:   absPath,
	}

	a.form = ui.CreateMetadataForm(true, ctx)

	statusBar := ui.CreateStatusBar("Tab: Cycle fields | Enter: Save | Esc: Clear | q: Quit")
	formWrapper := ui.CreateFormWrapper(a.form, "Editing: "+filepath.Base(absPath))

	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(formWrapper, 0, 1, true).
		AddItem(statusBar, 1, 0, false)

	ui.PopulateForm(a.form, a.meta.TrackName, a.meta.Artist, a.meta.Album, a.meta.CoverPath)

	mainFlex.SetInputCapture(ui.CreateInputCapture(true, ctx))

	a.root = mainFlex
	a.app.SetRoot(mainFlex, true)
	a.app.SetFocus(a.form.GetFormItem(0))

	return a.app.Run()
}

func (a *App) GetMetadata() *metadata.Metadata {
	return a.meta
}

func (a *App) SetMetadata(m *metadata.Metadata) {
	a.meta = m
}

func IsMP3File(filename string) bool {
	return strings.HasSuffix(strings.ToLower(filename), ".mp3")
}
