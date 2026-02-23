package metadata

import (
	"os"
	"strings"
	"testing"

	"github.com/bogem/id3v2"
)

const testFile = "./../../test/test.mp3"

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

func TestDiffNoChanges(t *testing.T) {
	m1 := &Metadata{
		TrackName: "Song",
		Artist:    "Artist",
		Album:     "Album",
	}
	m2 := &Metadata{
		TrackName: "Song",
		Artist:    "Artist",
		Album:     "Album",
	}

	diff := m1.Diff(m2)
	if diff != "" {
		t.Errorf("expected no changes, got '%s'", diff)
	}
}

func TestDiffOneChange(t *testing.T) {
	m1 := &Metadata{
		TrackName: "Song",
		Artist:    "Artist",
		Album:     "Album",
	}
	m2 := &Metadata{
		TrackName: "New Song",
		Artist:    "Artist",
		Album:     "Album",
	}

	diff := m1.Diff(m2)
	if diff == "" {
		t.Error("expected changes, got empty string")
	}
	if !contains(diff, "Song") {
		t.Errorf("expected diff to contain 'Song', got '%s'", diff)
	}
}

func TestDiffTwoChanges(t *testing.T) {
	m1 := &Metadata{
		TrackName: "Song",
		Artist:    "Artist",
		Album:     "Album",
	}
	m2 := &Metadata{
		TrackName: "New Song",
		Artist:    "New Artist",
		Album:     "Album",
	}

	diff := m1.Diff(m2)
	if diff == "" {
		t.Error("expected changes, got empty string")
	}
	if !contains(diff, "Song") || !contains(diff, "Artist") {
		t.Errorf("expected diff to contain 'Song' and 'Artist', got '%s'", diff)
	}
}

func TestDiffThreeChanges(t *testing.T) {
	m1 := &Metadata{
		TrackName: "Song",
		Artist:    "Artist",
		Album:     "Album",
	}
	m2 := &Metadata{
		TrackName: "New Song",
		Artist:    "New Artist",
		Album:     "New Album",
	}

	diff := m1.Diff(m2)
	if diff == "" {
		t.Error("expected changes, got empty string")
	}
	if !contains(diff, "Song") || !contains(diff, "Artist") || !contains(diff, "Album") {
		t.Errorf("expected diff to contain all changes, got '%s'", diff)
	}
}

func TestDiffEmptyToNonEmpty(t *testing.T) {
	m1 := &Metadata{
		TrackName: "",
		Artist:    "",
		Album:     "",
	}
	m2 := &Metadata{
		TrackName: "Song",
		Artist:    "Artist",
		Album:     "Album",
	}

	diff := m1.Diff(m2)
	if diff == "" {
		t.Error("expected changes, got empty string")
	}
	if !contains(diff, "(empty)") {
		t.Errorf("expected diff to contain '(empty)' for original values, got '%s'", diff)
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && strings.Contains(s, substr)
}

func TestSaveWithJPGCover(t *testing.T) {
	setupTestFile(t)
	clearMetadata(t)

	coverPath := "./../../test/test-cover.png"
	if _, err := os.Stat(coverPath); os.IsNotExist(err) {
		t.Skip("test-cover.png not found")
	}

	meta := &Metadata{
		TrackName: "Song with Cover",
		Artist:    "Artist",
		Album:     "Album",
		CoverPath: coverPath,
	}

	err := Save(testFile, meta)
	if err != nil {
		t.Fatalf("Save with cover failed: %v", err)
	}

	readMeta, err := Read(testFile)
	if err != nil {
		t.Fatalf("Read after save failed: %v", err)
	}

	if readMeta.TrackName != "Song with Cover" {
		t.Errorf("expected 'Song with Cover', got '%s'", readMeta.TrackName)
	}

	clearMetadata(t)
}

func TestReadNonExistentFile(t *testing.T) {
	meta, err := Read("./nonexistent-file.mp3")
	if err != nil {
		t.Errorf("expected no error for non-existent file, got: %v", err)
	}
	if meta == nil {
		t.Error("expected metadata to be returned even on error")
	}
	if meta.TrackName != "" || meta.Artist != "" || meta.Album != "" {
		t.Error("expected empty metadata for non-existent file")
	}
}

func TestReadFromFileWithMetadata(t *testing.T) {
	testFileWithMeta := "./../../test/test-w-metadata.mp3"
	if _, err := os.Stat(testFileWithMeta); os.IsNotExist(err) {
		t.Skip("test-w-metadata.mp3 not found")
	}

	meta, err := Read(testFileWithMeta)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if meta == nil {
		t.Fatal("metadata is nil")
	}

	if meta.TrackName == "" && meta.Artist == "" && meta.Album == "" {
		t.Log("Warning: file has no metadata, test may not be useful")
	}
}
