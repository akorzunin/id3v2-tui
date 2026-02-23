package app

import (
	"io"
	"os"
	"testing"

	"id3v2-tui/internal/metadata"
)

func TestIsMP3File(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"song.mp3", true},
		{"song.MP3", true},
		{"song.Mp3", true},
		{"song.wav", false},
		{"song.flac", false},
		{"mp3.txt", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := IsMP3File(tt.filename)
			if result != tt.expected {
				t.Errorf("IsMP3File(%q) = %v, expected %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestNewApp(t *testing.T) {
	app := NewApp()
	if app == nil {
		t.Fatal("NewApp returned nil")
	}
	if app.meta == nil {
		t.Error("app.meta is nil")
	}
}

func TestGettersSetters(t *testing.T) {
	app := NewApp()

	if app.getRoot() != nil {
		t.Error("expected root to be nil initially")
	}

	if app.getForm() != nil {
		t.Error("expected form to be nil initially")
	}

	if app.getFileList() != nil {
		t.Error("expected fileList to be nil initially")
	}

	if app.getCurrentDir() != "" {
		t.Error("expected currentDir to be empty initially")
	}

	app.setCurrentDir("/home/user/music")
	if app.getCurrentDir() != "/home/user/music" {
		t.Errorf("expected '/home/user/music', got '%s'", app.getCurrentDir())
	}

	if app.getFocusIndex() != 0 {
		t.Error("expected focusIndex to be 0 initially")
	}

	app.setFocusIndex(5)
	if app.getFocusIndex() != 5 {
		t.Errorf("expected 5, got %d", app.getFocusIndex())
	}
}

func TestGetMetadata(t *testing.T) {
	app := NewApp()

	meta := app.GetMetadata()
	if meta == nil {
		t.Error("GetMetadata returned nil")
	}

	app.meta = &metadata.Metadata{
		TrackName: "Test Song",
		Artist:    "Test Artist",
		Album:     "Test Album",
	}

	meta = app.GetMetadata()
	if meta.TrackName != "Test Song" {
		t.Errorf("expected 'Test Song', got '%s'", meta.TrackName)
	}
	if meta.Artist != "Test Artist" {
		t.Errorf("expected 'Test Artist', got '%s'", meta.Artist)
	}
	if meta.Album != "Test Album" {
		t.Errorf("expected 'Test Album', got '%s'", meta.Album)
	}
}

func TestSetMetadata(t *testing.T) {
	app := NewApp()

	meta := &metadata.Metadata{
		TrackName: "New Song",
		Artist:    "New Artist",
		Album:     "New Album",
	}

	app.SetMetadata(meta)

	if app.meta.TrackName != "New Song" {
		t.Errorf("expected 'New Song', got '%s'", app.meta.TrackName)
	}
}

func TestRunDirectEditWithDirectory(t *testing.T) {
	app := NewApp()

	err := app.Run("/home")
	if err == nil {
		t.Error("expected error for directory path")
	}
	if err != nil && err.Error() != "path is a directory, not a file" {
		t.Errorf("expected 'path is a directory, not a file', got '%s'", err.Error())
	}
}

func TestRunDirectEditWithInvalidPath(t *testing.T) {
	app := NewApp()

	err := app.Run("/nonexistent/path/to/file.mp3")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestReadMetadata(t *testing.T) {
	testFile := "../../test/test-w-metadata.mp3"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("test-w-metadata.mp3 not found")
	}

	app := NewApp()
	err := app.readMetadata(testFile)
	if err != nil {
		t.Errorf("readMetadata failed: %v", err)
	}

	if app.meta == nil {
		t.Fatal("metadata is nil after readMetadata")
	}

	if app.meta.TrackName == "" && app.meta.Artist == "" && app.meta.Album == "" {
		t.Log("Note: test file has no metadata")
	}
}

func TestReadMetadataNonExistent(t *testing.T) {
	app := NewApp()
	err := app.readMetadata("/nonexistent/file.mp3")
	if err != nil {
		t.Errorf("readMetadata should not fail for non-existent file, got: %v", err)
	}
	if app.meta == nil {
		t.Error("metadata should be initialized even for non-existent file")
	}
}

func TestSaveMetadataAndRestore(t *testing.T) {
	testFile := "../../test/test.mp3"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("test.mp3 not found")
	}

	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/test.mp3"

	src, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer src.Close()

	dst, err := os.Create(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		t.Fatalf("Failed to copy test file: %v", err)
	}
	dst.Close()

	originalMeta, err := metadata.Read(tmpFile)
	if err != nil {
		os.Remove(tmpFile)
		t.Fatalf("Failed to read original metadata: %v", err)
	}

	app := NewApp()
	app.originalMeta = originalMeta

	diff, err := app.saveMetadata(tmpFile, "Test Track", "Test Artist", "Test Album", "")
	if err != nil {
		os.Remove(tmpFile)
		t.Fatalf("saveMetadata failed: %v", err)
	}
	_ = diff

	verifyMeta, err := metadata.Read(tmpFile)
	if err != nil {
		os.Remove(tmpFile)
		t.Fatalf("Failed to read saved metadata: %v", err)
	}

	if verifyMeta.TrackName != "Test Track" {
		t.Errorf("expected track 'Test Track', got '%s'", verifyMeta.TrackName)
	}
	if verifyMeta.Artist != "Test Artist" {
		t.Errorf("expected artist 'Test Artist', got '%s'", verifyMeta.Artist)
	}
	if verifyMeta.Album != "Test Album" {
		t.Errorf("expected album 'Test Album', got '%s'", verifyMeta.Album)
	}

	err = metadata.Save(tmpFile, originalMeta)
	if err != nil {
		os.Remove(tmpFile)
		t.Fatalf("Failed to restore original metadata: %v", err)
	}
	os.Remove(tmpFile)
}

func TestSaveMetadataWithCoverAndRestore(t *testing.T) {
	testFile := "../../test/test.mp3"
	coverPath := "../../test/test-cover.png"

	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("test.mp3 not found")
	}
	if _, err := os.Stat(coverPath); os.IsNotExist(err) {
		t.Skip("test-cover.png not found")
	}

	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/test.mp3"

	src, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer src.Close()

	dst, err := os.Create(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		t.Fatalf("Failed to copy test file: %v", err)
	}
	dst.Close()

	originalMeta, err := metadata.Read(tmpFile)
	if err != nil {
		os.Remove(tmpFile)
		t.Fatalf("Failed to read original metadata: %v", err)
	}

	app := NewApp()
	app.originalMeta = originalMeta

	_, err = app.saveMetadata(tmpFile, "Cover Test", "Cover Artist", "Cover Album", coverPath)
	if err != nil {
		os.Remove(tmpFile)
		t.Fatalf("saveMetadata with cover failed: %v", err)
	}

	verifyMeta, err := metadata.Read(tmpFile)
	if err != nil {
		os.Remove(tmpFile)
		t.Fatalf("Failed to read saved metadata: %v", err)
	}

	if verifyMeta.TrackName != "Cover Test" {
		t.Errorf("expected track 'Cover Test', got '%s'", verifyMeta.TrackName)
	}

	restoredMeta := &metadata.Metadata{
		TrackName: originalMeta.TrackName,
		Artist:    originalMeta.Artist,
		Album:     originalMeta.Album,
		CoverPath: "",
	}
	err = metadata.Save(tmpFile, restoredMeta)
	if err != nil {
		os.Remove(tmpFile)
		t.Fatalf("Failed to restore original metadata: %v", err)
	}
	os.Remove(tmpFile)
}
