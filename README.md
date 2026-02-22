# id3v2-tui

A terminal user interface for editing ID3v2 metadata tags on MP3 files.

<img width="1296" height="657" alt="Image" src="https://github.com/user-attachments/assets/817007e5-ee9f-4d9a-a029-cc2b5ea5d95a" />

## Features

- Browse directories and select MP3 files
- Edit track name, artist, album, and cover image
- Direct file editing mode via command line argument
- Keyboard-driven navigation

## Requirements

- Go 1.25+
- `ffmpeg` (includes ffprobe, for reading metadata and embedding cover art)

### Installing Dependencies

```bash
# archlinux
pacman -S ffmpeg

# On Debian/Ubuntu
sudo apt install ffmpeg

# On macOS
brew install ffmpeg
```

## Build

```bash
go build -o ./build/id3v2-tui
```

## Usage

```bash
# Interactive mode - browse files
./id3v2-tui

# Direct edit mode - edit specific file
./id3v2-tui /path/to/song.mp3
```

### Keybindings

| Key             | Action                       |
| --------------- | ---------------------------- |
| `↑/↓`           | Navigate file list           |
| `Enter`         | Open directory / select file |
| `Tab/Shift+Tab` | Cycle focus between panels   |
| `Esc`           | Clear form fields            |
| `q`             | Quit                         |

## Testing

```bash
go test .
```

Tests require a `test/test.mp3` file for metadata operations.

## Run pre-commit hooks

```bash
pre-commit run --all-files
```

## Similar Projects

- [clid](https://github.com/sudormrfbin/clid/tree/rewrite) - python based TUI
- [id3tui](https://github.com/cozyGalvinism/id3tui) - rust based TUI
- [kid3](https://github.com/KDE/kid3) - KDE's GUI for editing ID3 tags

## License

MIT
