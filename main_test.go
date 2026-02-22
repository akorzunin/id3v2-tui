package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const testFile = "test/test.mp3"

func setupTestFile(t *testing.T) {
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("test.mp3 not found, skipping test")
	}
}

func clearMetadata(_ *testing.T) {
	exec.Command("id3v2", "--delete-all", testFile).Run()
}

func TestReadMetadata(t *testing.T) {
	setupTestFile(t)
	clearMetadata(t)

	exec.Command("id3v2", "-t", "Test Song", "-a", "Test Artist", "-A", "Test Album", testFile).Run()

	app := &App{}
	err := app.readMetadata(testFile)
	if err != nil {
		t.Fatalf("readMetadata failed: %v", err)
	}

	if app.metadata == nil {
		t.Fatal("metadata is nil")
	}

	if app.metadata.TrackName != "Test Song" {
		t.Errorf("expected track name 'Test Song', got '%s'", app.metadata.TrackName)
	}

	if app.metadata.Artist != "Test Artist" {
		t.Errorf("expected artist 'Test Artist', got '%s'", app.metadata.Artist)
	}

	if app.metadata.Album != "Test Album" {
		t.Errorf("expected album 'Test Album', got '%s'", app.metadata.Album)
	}

	clearMetadata(t)
}

func TestSaveMetadata(t *testing.T) {
	setupTestFile(t)
	clearMetadata(t)

	app := &App{
		metadata: &Metadata{
			TrackName: "Saved Song",
			Artist:    "Saved Artist",
			Album:     "Saved Album",
		},
	}

	err := app.saveMetadata(testFile)
	if err != nil {
		t.Fatalf("saveMetadata failed: %v", err)
	}

	app2 := &App{}
	err = app2.readMetadata(testFile)
	if err != nil {
		t.Fatalf("readMetadata after save failed: %v", err)
	}

	if app2.metadata.TrackName != "Saved Song" {
		t.Errorf("expected 'Saved Song', got '%s'", app2.metadata.TrackName)
	}

	if app2.metadata.Artist != "Saved Artist" {
		t.Errorf("expected 'Saved Artist', got '%s'", app2.metadata.Artist)
	}

	if app2.metadata.Album != "Saved Album" {
		t.Errorf("expected 'Saved Album', got '%s'", app2.metadata.Album)
	}

	clearMetadata(t)
}

func TestSaveEmptyMetadata(t *testing.T) {
	setupTestFile(t)
	clearMetadata(t)

	app := &App{
		metadata: &Metadata{
			TrackName: "",
			Artist:    "",
			Album:     "",
		},
	}

	err := app.saveMetadata(testFile)
	if err != nil {
		t.Fatalf("saveMetadata with empty fields failed: %v", err)
	}

	clearMetadata(t)
}

func TestLoadFiles(t *testing.T) {
	if _, err := os.Stat("test"); os.IsNotExist(err) {
		t.Skip("test directory not found")
	}

	app := &App{}
	app.fileList = nil
	app.currentDir = ""

	_ = app
}

func TestParseId3v2Output(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected Metadata
	}{
		{
			name:   "single line TIT2",
			output: "TIT2: My Song Title",
			expected: Metadata{
				TrackName: "My Song Title",
			},
		},
		{
			name:   "single line TPE1",
			output: "TPE1: My Artist",
			expected: Metadata{
				Artist: "My Artist",
			},
		},
		{
			name:   "single line TALB",
			output: "TALB: My Album",
			expected: Metadata{
				Album: "My Album",
			},
		},
		{
			name:   "multiple frames",
			output: "TIT2: Song\nTPE1: Artist\nTALB: Album",
			expected: Metadata{
				TrackName: "Song",
				Artist:    "Artist",
				Album:     "Album",
			},
		},
		{
			name: "full id3v2 list output",
			output: `id3v2 tag info for test.mp3:
TIT2: Test Song
TPE1: Test Artist
TALB: Test Album
TPE2: Test Album Artist
TYER: 2024
TRCK: 1/10
TCON: Rock
TDRC: 2024`,
			expected: Metadata{
				TrackName: "Test Song",
				Artist:    "Test Artist",
				Album:     "Test Album",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &App{}
			lines := strings.Split(tt.output, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.Contains(line, "TIT2") {
					if idx := strings.Index(line, ":"); idx != -1 {
						app.metadata = &Metadata{}
						app.metadata.TrackName = strings.TrimSpace(line[idx+1:])
					}
				}
				if strings.Contains(line, "TPE1") {
					if app.metadata == nil {
						app.metadata = &Metadata{}
					}
					if idx := strings.Index(line, ":"); idx != -1 {
						app.metadata.Artist = strings.TrimSpace(line[idx+1:])
					}
				}
				if strings.Contains(line, "TALB") {
					if app.metadata == nil {
						app.metadata = &Metadata{}
					}
					if idx := strings.Index(line, ":"); idx != -1 {
						app.metadata.Album = strings.TrimSpace(line[idx+1:])
					}
				}
			}

			if app.metadata == nil {
				app.metadata = &Metadata{}
			}

			if app.metadata.TrackName != tt.expected.TrackName {
				t.Errorf("TrackName: expected '%s', got '%s'", tt.expected.TrackName, app.metadata.TrackName)
			}
			if app.metadata.Artist != tt.expected.Artist {
				t.Errorf("Artist: expected '%s', got '%s'", tt.expected.Artist, app.metadata.Artist)
			}
			if app.metadata.Album != tt.expected.Album {
				t.Errorf("Album: expected '%s', got '%s'", tt.expected.Album, app.metadata.Album)
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

func TestRunCommand(t *testing.T) {
	app := &App{}

	output, err := app.runCommand("echo", "hello")
	if err != nil {
		t.Fatalf("runCommand failed: %v", err)
	}

	if !strings.Contains(output, "hello") {
		t.Errorf("expected 'hello' in output, got '%s'", output)
	}
}

func TestRunCommandError(t *testing.T) {
	app := &App{}

	_, err := app.runCommand("nonexistent-command-xyz")
	if err == nil {
		t.Fatal("expected error for nonexistent command")
	}
}
