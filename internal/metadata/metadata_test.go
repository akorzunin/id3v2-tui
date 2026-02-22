package metadata

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/bogem/id3v2"

	"id3v2-tui/internal/commands"
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

	executor := commands.NewExecutor()
	meta, err := Read(executor, testFile)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if meta == nil {
		t.Fatal("metadata is nil")
	}

	if meta.TrackName != "Test Song" {
		t.Errorf("expected track name 'Test Song', got '%s'", meta.TrackName)
	}

	if meta.Artist != "Test Artist" {
		t.Errorf("expected artist 'Test Artist', got '%s'", meta.Artist)
	}

	if meta.Album != "Test Album" {
		t.Errorf("expected album 'Test Album', got '%s'", meta.Album)
	}

	clearMetadata(t)
}

func TestSaveMetadata(t *testing.T) {
	setupTestFile(t)
	clearMetadata(t)

	executor := commands.NewExecutor()
	meta := &Metadata{
		TrackName: "Saved Song",
		Artist:    "Saved Artist",
		Album:     "Saved Album",
	}

	err := Save(executor, testFile, meta)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	readMeta, err := Read(executor, testFile)
	if err != nil {
		t.Fatalf("Read after save failed: %v", err)
	}

	if readMeta.TrackName != "Saved Song" {
		t.Errorf("expected 'Saved Song', got '%s'", readMeta.TrackName)
	}

	if readMeta.Artist != "Saved Artist" {
		t.Errorf("expected 'Saved Artist', got '%s'", readMeta.Artist)
	}

	if readMeta.Album != "Saved Album" {
		t.Errorf("expected 'Saved Album', got '%s'", readMeta.Album)
	}

	clearMetadata(t)
}

func TestSaveEmptyMetadata(t *testing.T) {
	setupTestFile(t)
	clearMetadata(t)

	executor := commands.NewExecutor()
	meta := &Metadata{
		TrackName: "",
		Artist:    "",
		Album:     "",
	}

	err := Save(executor, testFile, meta)
	if err != nil {
		t.Fatalf("Save with empty fields failed: %v", err)
	}

	clearMetadata(t)
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
