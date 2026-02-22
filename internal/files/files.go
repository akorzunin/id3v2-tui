package files

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/rivo/tview"
)

func Load(list *tview.List, dir string) string {
	list.Clear()

	parentDir := filepath.Dir(dir)
	if parentDir != dir {
		list.AddItem("..", "Go to parent directory", 0, nil)
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return dir
	}

	for _, file := range files {
		if file.IsDir() {
			list.AddItem(file.Name()+"/", "Directory", 0, nil)
		} else if strings.HasSuffix(strings.ToLower(file.Name()), ".mp3") {
			list.AddItem(file.Name(), "MP3 file", 0, nil)
		}
	}

	list.SetTitle("Files - " + filepath.Base(dir))
	return dir
}

func GetSelectedPath(list *tview.List, currentDir string) string {
	if list.GetItemCount() == 0 {
		return ""
	}
	selectedItem := list.GetCurrentItem()
	if selectedItem < 0 {
		return ""
	}
	mainText, _ := list.GetItemText(selectedItem)
	if mainText == "" || strings.HasSuffix(mainText, "/") || mainText == ".." {
		return ""
	}
	return filepath.Join(currentDir, mainText)
}

func IsDirectoryEntry(text string) bool {
	return strings.HasSuffix(text, "/") || text == ".."
}

func ResolveDirectory(currentDir, entry string) string {
	if entry == ".." {
		return filepath.Dir(currentDir)
	}
	return filepath.Join(currentDir, entry)
}
