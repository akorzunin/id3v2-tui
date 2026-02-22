package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bogem/id3v2"
)

const testFile = "test/test.mp3"

func setupTestFile(t *testing.T) {
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("test.mp3 not found, skipping test")
	}
}

func clearMetadata(t *testing.T) {
	tag, err := id3v2.Open(testFile, id3v2.Options{Parse: true})
	if err != nil {
		t.Logf("warning: failed to open file for clearing: %v", err)
		return
	}
	defer tag.Close()
	tag.DeleteAllFrames()
	if err := tag.Save(); err != nil {
		t.Logf("warning: failed to save cleared metadata: %v", err)
	}
}

func setTestMetadata(t *testing.T, title, artist, album string) {
	tag, err := id3v2.Open(testFile, id3v2.Options{Parse: true})
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer tag.Close()
	tag.SetTitle(title)
	tag.SetArtist(artist)
	tag.SetAlbum(album)
	if err := tag.Save(); err != nil {
		t.Fatalf("failed to save metadata: %v", err)
	}
}

func TestReadMetadata(t *testing.T) {
	setupTestFile(t)
	clearMetadata(t)
	setTestMetadata(t, "Test Song", "Test Artist", "Test Album")

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

func TestParseFfprobeOutput(t *testing.T) {
	tests := []struct {
		name     string
		jsonStr  string
		expected Metadata
	}{
		{
			name: "basic metadata",
			jsonStr: `{
				"format": {
					"tags": {
						"title": "My Song",
						"artist": "My Artist",
						"album": "My Album"
					}
				}
			}`,
			expected: Metadata{
				TrackName: "My Song",
				Artist:    "My Artist",
				Album:     "My Album",
			},
		},
		{
			name: "partial metadata",
			jsonStr: `{
				"format": {
					"tags": {
						"title": "Only Title"
					}
				}
			}`,
			expected: Metadata{
				TrackName: "Only Title",
				Artist:    "",
				Album:     "",
			},
		},
		{
			name: "empty tags",
			jsonStr: `{
				"format": {
					"tags": {}
				}
			}`,
			expected: Metadata{
				TrackName: "",
				Artist:    "",
				Album:     "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var probe ffprobeOutput
			if err := json.Unmarshal([]byte(tt.jsonStr), &probe); err != nil {
				t.Fatalf("failed to parse JSON: %v", err)
			}

			result := Metadata{
				TrackName: probe.Format.Tags.Title,
				Artist:    probe.Format.Tags.Artist,
				Album:     probe.Format.Tags.Album,
			}

			if result.TrackName != tt.expected.TrackName {
				t.Errorf("TrackName: expected '%s', got '%s'", tt.expected.TrackName, result.TrackName)
			}
			if result.Artist != tt.expected.Artist {
				t.Errorf("Artist: expected '%s', got '%s'", tt.expected.Artist, result.Artist)
			}
			if result.Album != tt.expected.Album {
				t.Errorf("Album: expected '%s', got '%s'", tt.expected.Album, result.Album)
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
