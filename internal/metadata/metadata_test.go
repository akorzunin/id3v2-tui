package metadata

import (
	"os"
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

	meta, err := Read(testFile)
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

	meta := &Metadata{
		TrackName: "Saved Song",
		Artist:    "Saved Artist",
		Album:     "Saved Album",
	}

	err := Save(testFile, meta)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	readMeta, err := Read(testFile)
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

	meta := &Metadata{
		TrackName: "",
		Artist:    "",
		Album:     "",
	}

	err := Save(testFile, meta)
	if err != nil {
		t.Fatalf("Save with empty fields failed: %v", err)
	}

	clearMetadata(t)
}
