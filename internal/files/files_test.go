package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rivo/tview"
)

func TestLoadFiles(t *testing.T) {
	if _, err := os.Stat("test"); os.IsNotExist(err) {
		t.Skip("test directory not found")
	}

	list := tview.NewList()
	dir := Load(list, "test")

	if dir != "test" {
		t.Errorf("expected dir 'test', got '%s'", dir)
	}
}

func TestGetSelectedPath(t *testing.T) {
	list := tview.NewList()
	list.AddItem("test.mp3", "MP3 file", 0, nil)
	list.AddItem("other.mp3", "MP3 file", 0, nil)

	path := GetSelectedPath(list, "/home/user/music")
	if path != "/home/user/music/test.mp3" {
		t.Errorf("expected '/home/user/music/test.mp3', got '%s'", path)
	}
}

func TestGetSelectedPathEmpty(t *testing.T) {
	list := tview.NewList()

	path := GetSelectedPath(list, "/home/user/music")
	if path != "" {
		t.Errorf("expected empty path, got '%s'", path)
	}
}

func TestGetSelectedPathDirectory(t *testing.T) {
	list := tview.NewList()
	list.AddItem("subdir/", "Directory", 0, nil)

	path := GetSelectedPath(list, "/home/user/music")
	if path != "" {
		t.Errorf("expected empty path for directory, got '%s'", path)
	}
}

func TestIsDirectoryEntry(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"folder/", true},
		{"..", true},
		{"file.mp3", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := IsDirectoryEntry(tt.input)
			if result != tt.expected {
				t.Errorf("IsDirectoryEntry(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestResolveDirectory(t *testing.T) {
	tests := []struct {
		currentDir string
		entry      string
		expected   string
	}{
		{"/home/user/music", "..", "/home/user"},
		{"/home/user/music", "subfolder", "/home/user/music/subfolder"},
		{"/home", "..", "/"},
	}

	for _, tt := range tests {
		t.Run(tt.currentDir+"/"+tt.entry, func(t *testing.T) {
			result := ResolveDirectory(tt.currentDir, tt.entry)
			if result != tt.expected {
				t.Errorf("ResolveDirectory(%q, %q) = %q, expected %q", tt.currentDir, tt.entry, result, tt.expected)
			}
		})
	}
}

func TestFilePathOperations(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"test.mp3", "test.mp3"},
		{"/path/to/test.mp3", "test.mp3"},
		{"../test.mp3", "test.mp3"},
		{"/path/to/file", "file"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := filepath.Base(tt.input)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestLoadWithTestDirectory(t *testing.T) {
	testDir := "../../test"
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Skip("test directory not found")
	}

	list := tview.NewList()
	dir := Load(list, testDir)

	if dir != testDir {
		t.Errorf("expected dir '%s', got '%s'", testDir, dir)
	}

	count := list.GetItemCount()
	if count == 0 {
		t.Error("expected items in list, got 0")
	}

	foundMP3 := false
	for i := 0; i < count; i++ {
		mainText, _ := list.GetItemText(i)
		if strings.HasSuffix(mainText, ".mp3") {
			foundMP3 = true
			break
		}
	}
	if !foundMP3 {
		t.Error("expected to find at least one .mp3 file in test directory")
	}
}

func TestLoadWithNonExistentDirectory(t *testing.T) {
	list := tview.NewList()
	dir := Load(list, "/nonexistent/directory")

	if dir != "/nonexistent/directory" {
		t.Errorf("expected original dir to be returned, got '%s'", dir)
	}
}

func TestLoadParentDirectory(t *testing.T) {
	testDir := "../../test"
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Skip("test directory not found")
	}

	list := tview.NewList()
	Load(list, testDir)

	count := list.GetItemCount()
	if count == 0 {
		t.Error("expected items in list")
	}

	hasParentDir := false
	for i := 0; i < count; i++ {
		mainText, _ := list.GetItemText(i)
		if mainText == ".." {
			hasParentDir = true
			break
		}
	}
	if !hasParentDir {
		t.Error("expected '..' entry for non-root directory")
	}
}
