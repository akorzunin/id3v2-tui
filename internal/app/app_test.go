package app

import (
	"testing"
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
	if app.executor == nil {
		t.Error("app.executor is nil")
	}
}
