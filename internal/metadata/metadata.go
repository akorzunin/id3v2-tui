package metadata

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bogem/id3v2"
)

type Metadata struct {
	TrackName string
	Artist    string
	Album     string
	CoverPath string
}

func Read(filePath string) (*Metadata, error) {
	tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
	if err != nil {
		return &Metadata{}, nil
	}
	defer tag.Close()

	return &Metadata{
		TrackName: tag.Title(),
		Artist:    tag.Artist(),
		Album:     tag.Album(),
	}, nil
}

func getMimeType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	default:
		return "image/jpeg"
	}
}

func Save(filePath string, meta *Metadata) error {
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

	if meta.CoverPath != "" {
		artwork, err := os.ReadFile(meta.CoverPath)
		if err != nil {
			return fmt.Errorf("failed to read cover file: %w", err)
		}

		pic := id3v2.PictureFrame{
			Encoding:    id3v2.EncodingUTF8,
			MimeType:    getMimeType(meta.CoverPath),
			PictureType: id3v2.PTFrontCover,
			Description: "Front cover",
			Picture:     artwork,
		}
		tag.AddAttachedPicture(pic)
	}

	if err := tag.Save(); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	return nil
}
