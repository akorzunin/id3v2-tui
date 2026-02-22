package metadata

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bogem/id3v2"

	"id3v2-tui/internal/commands"
)

type Metadata struct {
	TrackName string
	Artist    string
	Album     string
	CoverPath string
}

type ffprobeFormat struct {
	Tags struct {
		Title  string `json:"title"`
		Artist string `json:"artist"`
		Album  string `json:"album"`
	} `json:"tags"`
}

type ffprobeOutput struct {
	Format ffprobeFormat `json:"format"`
}

func Read(executor commands.Executor, filePath string) (*Metadata, error) {
	output, err := executor.Run("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", filePath)
	if err != nil {
		return &Metadata{}, nil
	}

	var probe ffprobeOutput
	if err := json.Unmarshal([]byte(output), &probe); err != nil {
		return &Metadata{}, nil
	}

	return &Metadata{
		TrackName: probe.Format.Tags.Title,
		Artist:    probe.Format.Tags.Artist,
		Album:     probe.Format.Tags.Album,
	}, nil
}

func Save(executor commands.Executor, filePath string, meta *Metadata) error {
	if meta.CoverPath != "" {
		_, err := executor.Run("ffmpeg", "-i", filePath, "-i", meta.CoverPath,
			"-map", "0:0", "-map", "1:0", "-c:v", "copy", "-id3v2_version", "3",
			"-metadata:s:v", "title=Album cover", "-metadata:s:v", "comment=Cover (front)",
			"-metadata", "title="+meta.TrackName,
			"-metadata", "artist="+meta.Artist,
			"-metadata", "album="+meta.Album,
			"-c:a", "copy", "-y", filePath+".tmp.mp3")
		if err != nil {
			return fmt.Errorf("failed to set cover: %w", err)
		}
		os.Rename(filePath+".tmp.mp3", filePath)
		return nil
	}

	tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer tag.Close()

	if meta.TrackName != "" {
		tag.SetTitle(meta.TrackName)
	}
	if meta.Artist != "" {
		tag.SetArtist(meta.Artist)
	}
	if meta.Album != "" {
		tag.SetAlbum(meta.Album)
	}

	if err := tag.Save(); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	return nil
}
