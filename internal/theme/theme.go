package theme

import "github.com/gdamore/tcell/v2"

const (
	HexPrimary   = "#1C4D8D"
	HexSecondary = "#0F2854"
	HexText      = "#BDE8F5"
	HexTextDim   = "#5A8FBF"
)

var (
	Background = tcell.ColorBlack
	Primary    = tcell.NewHexColor(0x1C4D8D)
	Secondary  = tcell.NewHexColor(0x0F2854)
	Text       = tcell.NewHexColor(0xBDE8F5)
	TextDim    = tcell.NewHexColor(0x5A8FBF)
	Error      = tcell.NewHexColor(0xE57373)
)
