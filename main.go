package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bogem/id3v2"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Metadata struct {
	TrackName string
	Artist    string
	Album     string
	CoverPath string
}

type App struct {
	app         *tview.Application
	fileList    *tview.List
	form        *tview.Form
	pages       *tview.Pages
	root        tview.Primitive
	metadata    *Metadata
	currentDir  string
	currentFile string
	focusIndex  int
}

func NewApp() *App {
	return &App{
		metadata: &Metadata{},
	}
}

func (a *App) runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: %w", string(output), err)
	}
	return string(output), nil
}

type ffprobeFormat struct {
	Tags struct {
		Title  string `json:"title"`
		Artist string `json:"artist"`
		Album  string `json:"album"`
	} `json:"tags"`
}

type ffprobeOutput struct {
	Format ffprobeFormat `json:"format"`
}

func (a *App) readMetadata(filePath string) error {
	output, err := a.runCommand("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", filePath)
	if err != nil {
		a.metadata = &Metadata{}
		return nil
	}

	a.metadata = &Metadata{}

	var probe ffprobeOutput
	if err := json.Unmarshal([]byte(output), &probe); err != nil {
		return nil
	}

	a.metadata.TrackName = probe.Format.Tags.Title
	a.metadata.Artist = probe.Format.Tags.Artist
	a.metadata.Album = probe.Format.Tags.Album

	return nil
}

func (a *App) saveMetadata(filePath string) error {
	if a.metadata.CoverPath != "" {
		_, err := a.runCommand("ffmpeg", "-i", filePath, "-i", a.metadata.CoverPath,
			"-map", "0:0", "-map", "1:0", "-c:v", "copy", "-id3v2_version", "3",
			"-metadata:s:v", "title=Album cover", "-metadata:s:v", "comment=Cover (front)",
			"-metadata", "title="+a.metadata.TrackName,
			"-metadata", "artist="+a.metadata.Artist,
			"-metadata", "album="+a.metadata.Album,
			"-c:a", "copy", "-y", filePath+".tmp.mp3")
		if err != nil {
			return fmt.Errorf("failed to set cover: %w", err)
		}
		os.Rename(filePath+".tmp.mp3", filePath)
		return nil
	}

	tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer tag.Close()

	if a.metadata.TrackName != "" {
		tag.SetTitle(a.metadata.TrackName)
	}
	if a.metadata.Artist != "" {
		tag.SetArtist(a.metadata.Artist)
	}
	if a.metadata.Album != "" {
		tag.SetAlbum(a.metadata.Album)
	}

	if err := tag.Save(); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	return nil
}

func (a *App) loadFiles(dir string) {
	a.fileList.Clear()
	a.currentDir = dir

	parentDir := filepath.Dir(dir)
	if parentDir != dir {
		a.fileList.AddItem("..", "Go to parent directory", 0, nil)
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, file := range files {
		if file.IsDir() {
			a.fileList.AddItem(file.Name()+"/", "Directory", 0, nil)
		} else if strings.HasSuffix(strings.ToLower(file.Name()), ".mp3") {
			a.fileList.AddItem(file.Name(), "MP3 file", 0, nil)
		}
	}

	a.fileList.SetTitle("Files - " + filepath.Base(dir))
}

func (a *App) createFileBrowser() *tview.List {
	a.fileList = tview.NewList()
	a.fileList.SetBorder(true).SetTitle("Files")

	currentDir, _ := os.Getwd()
	a.currentDir = currentDir
	a.loadFiles(currentDir)

	return a.fileList
}

func (a *App) createMetadataForm(directMode bool) *tview.Form {
	form := tview.NewForm()
	form.SetBorder(true).SetTitle("Metadata Editor")

	form.AddInputField("Track Name", "", 40, nil, nil)
	form.AddInputField("Artist", "", 40, nil, nil)
	form.AddInputField("Album", "", 40, nil, nil)
	form.AddInputField("Cover Image Path", "", 40, nil, nil)

	form.AddButton("Save", func() {
		var filePath string
		if directMode {
			filePath = a.currentFile
		} else {
			selectedItem := a.fileList.GetCurrentItem()
			if selectedItem >= 0 {
				mainText, _ := a.fileList.GetItemText(selectedItem)
				if mainText != "" && !strings.HasSuffix(mainText, "/") && mainText != ".." {
					filePath = filepath.Join(a.currentDir, mainText)
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

		a.metadata.TrackName = trackName
		a.metadata.Artist = artist
		a.metadata.Album = album
		a.metadata.CoverPath = coverPath

		err := a.saveMetadata(filePath)
		if err != nil {
			a.showError(err.Error())
		} else {
			a.showMessage("Metadata saved successfully!")
		}
	})

	form.AddButton("Clear", func() {
		form.GetFormItemByLabel("Track Name").(*tview.InputField).SetText("")
		form.GetFormItemByLabel("Artist").(*tview.InputField).SetText("")
		form.GetFormItemByLabel("Album").(*tview.InputField).SetText("")
		form.GetFormItemByLabel("Cover Image Path").(*tview.InputField).SetText("")
	})

	form.SetButtonsAlign(tview.AlignCenter)
	form.SetButtonBackgroundColor(tcell.ColorDarkSlateGray)
	form.SetButtonTextColor(tcell.ColorWhite)

	return form
}

func (a *App) showError(msg string) {
	modal := tview.NewModal()
	modal.SetText(msg)
	modal.SetTextColor(tcell.ColorRed)
	modal.AddButtons([]string{"OK"})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		a.app.SetRoot(a.root, false)
	})
	a.app.SetRoot(modal, false)
}

func (a *App) showMessage(msg string) {
	modal := tview.NewModal()
	modal.SetText(msg)
	modal.SetTextColor(tcell.ColorGreen)
	modal.AddButtons([]string{"OK"})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		a.app.SetRoot(a.root, false)
	})
	a.app.SetRoot(modal, false)
}

func (a *App) Run(filePath string) error {
	a.app = tview.NewApplication()

	if filePath != "" {
		return a.runDirectEdit(filePath)
	}

	a.fileList = a.createFileBrowser()
	a.form = a.createMetadataForm(false)

	statusBar := tview.NewTextView().
		SetText("↑↓ Navigate | Enter: Open | Tab/Shift+Tab: Cycle | Esc: Clear | q: Quit").
		SetTextColor(tcell.ColorDarkGray).
		SetTextAlign(tview.AlignCenter)
	statusBar.SetBorder(false)

	fileListWrapper := tview.NewFrame(a.fileList).
		AddText(" [red]Files[white] ", true, tview.AlignLeft, tcell.ColorRed).
		AddText("", false, tview.AlignLeft, tcell.ColorWhite)
	fileListWrapper.SetBorder(true).SetTitle("Files")

	formWrapper := tview.NewFrame(a.form).
		AddText(" [green]Metadata Editor[white] ", true, tview.AlignLeft, tcell.ColorGreen).
		AddText("", false, tview.AlignLeft, tcell.ColorWhite)
	formWrapper.SetBorder(true)

	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(fileListWrapper, 0, 1, true).
			AddItem(formWrapper, 0, 2, false), 0, 1, true).
		AddItem(statusBar, 1, 0, false)

	a.fileList.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		if strings.HasSuffix(mainText, "/") || mainText == ".." {
			newDir := filepath.Join(a.currentDir, mainText)
			if mainText == ".." {
				newDir = filepath.Dir(a.currentDir)
			}
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

		form := a.form
		form.GetFormItemByLabel("Track Name").(*tview.InputField).SetText(a.metadata.TrackName)
		form.GetFormItemByLabel("Artist").(*tview.InputField).SetText(a.metadata.Artist)
		form.GetFormItemByLabel("Album").(*tview.InputField).SetText(a.metadata.Album)
		form.GetFormItemByLabel("Cover Image Path").(*tview.InputField).SetText(a.metadata.CoverPath)

		a.focusIndex = 1
		a.app.SetFocus(a.form.GetFormItem(0))
	})

	mainFlex.SetInputCapture(a.createInputCapture(false))

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

	a.form = a.createMetadataForm(true)

	statusBar := tview.NewTextView().
		SetText("Tab: Cycle fields | Enter: Save | Esc: Clear | q: Quit").
		SetTextColor(tcell.ColorDarkGray).
		SetTextAlign(tview.AlignCenter)
	statusBar.SetBorder(false)

	formWrapper := tview.NewFrame(a.form).
		AddText(" [green]Editing: "+filepath.Base(absPath)+"[white] ", true, tview.AlignLeft, tcell.ColorGreen).
		AddText("", false, tview.AlignLeft, tcell.ColorWhite)
	formWrapper.SetBorder(true)

	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(formWrapper, 0, 1, true).
		AddItem(statusBar, 1, 0, false)

	a.form.GetFormItemByLabel("Track Name").(*tview.InputField).SetText(a.metadata.TrackName)
	a.form.GetFormItemByLabel("Artist").(*tview.InputField).SetText(a.metadata.Artist)
	a.form.GetFormItemByLabel("Album").(*tview.InputField).SetText(a.metadata.Album)
	a.form.GetFormItemByLabel("Cover Image Path").(*tview.InputField).SetText(a.metadata.CoverPath)

	mainFlex.SetInputCapture(a.createInputCapture(true))

	a.root = mainFlex
	a.app.SetRoot(mainFlex, true)
	a.app.SetFocus(a.form.GetFormItem(0))

	return a.app.Run()
}

func (a *App) createInputCapture(directMode bool) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'q' {
			a.app.Stop()
			return nil
		}
		if event.Key() == tcell.KeyTab {
			numFormItems := a.form.GetFormItemCount()
			numButtons := a.form.GetButtonCount()
			totalItems := numFormItems + numButtons

			if !directMode {
				totalItems++
			}

			a.focusIndex = (a.focusIndex + 1) % totalItems

			if !directMode && a.focusIndex == 0 {
				a.app.SetFocus(a.fileList)
			} else {
				idx := a.focusIndex
				if !directMode {
					idx--
				}
				if idx < numFormItems {
					a.app.SetFocus(a.form.GetFormItem(idx))
				} else {
					a.app.SetFocus(a.form.GetButton(idx - numFormItems))
				}
			}
			return nil
		}
		if event.Key() == tcell.KeyBacktab {
			numFormItems := a.form.GetFormItemCount()
			numButtons := a.form.GetButtonCount()
			totalItems := numFormItems + numButtons

			if !directMode {
				totalItems++
			}

			a.focusIndex--
			if a.focusIndex < 0 {
				a.focusIndex = totalItems - 1
			}

			if !directMode && a.focusIndex == 0 {
				a.app.SetFocus(a.fileList)
			} else {
				idx := a.focusIndex
				if !directMode {
					idx--
				}
				if idx < 0 {
					a.app.SetFocus(a.form.GetButton(numButtons - 1))
				} else if idx < numFormItems {
					a.app.SetFocus(a.form.GetFormItem(idx))
				} else {
					a.app.SetFocus(a.form.GetButton(idx - numFormItems))
				}
			}
			return nil
		}
		if event.Key() == tcell.KeyEsc {
			a.form.GetFormItemByLabel("Track Name").(*tview.InputField).SetText("")
			a.form.GetFormItemByLabel("Artist").(*tview.InputField).SetText("")
			a.form.GetFormItemByLabel("Album").(*tview.InputField).SetText("")
			a.form.GetFormItemByLabel("Cover Image Path").(*tview.InputField).SetText("")
			return nil
		}
		return event
	}
}

func main() {
	filePath := ""
	if len(os.Args) > 1 {
		filePath = os.Args[1]
	}

	app := NewApp()
	if err := app.Run(filePath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
