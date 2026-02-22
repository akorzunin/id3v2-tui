package files

import (
	"os"
	"path/filepath"
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
