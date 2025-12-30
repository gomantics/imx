package ui

import (
	"github.com/fatih/color"
)

var (
	// Text styles
	Bold = color.New(color.Bold)
	Dim  = color.New(color.Faint)

	// Colors
	Red     = color.New(color.FgRed)
	Green   = color.New(color.FgGreen)
	Yellow  = color.New(color.FgYellow)
	Blue    = color.New(color.FgBlue)
	Cyan    = color.New(color.FgCyan)
	Magenta = color.New(color.FgMagenta)
	White   = color.New(color.FgWhite)

	// Combined styles
	BoldRed     = color.New(color.Bold, color.FgRed)
	BoldGreen   = color.New(color.Bold, color.FgGreen)
	BoldYellow  = color.New(color.Bold, color.FgYellow)
	BoldCyan    = color.New(color.Bold, color.FgCyan)
	BoldMagenta = color.New(color.Bold, color.FgMagenta)
)

// DisableColors disables all color output
func DisableColors() {
	color.NoColor = true
}

// EnableColors enables color output
func EnableColors() {
	color.NoColor = false
}

// SpecColor returns a color for the given spec type
func SpecColor(dirName string) *color.Color {
	switch dirName {
	case "EXIF", "exif", "TIFF IFD0", "TIFF IFD1", "TIFF SubIFD":
		return Green
	case "IPTC", "iptc":
		return Blue
	case "XMP", "xmp":
		return Cyan
	case "ICC", "icc":
		return Yellow
	case "ID3", "id3", "ID3v2.2", "ID3v2.3", "ID3v2.4":
		return Magenta
	case "FLAC", "flac", "FLAC-StreamInfo", "FLAC-VorbisComment", "FLAC-Picture", "FLAC-Application", "FLAC-SeekTable", "FLAC-CueSheet":
		return Green
	case "MP4", "mp4", "MP4-ftyp", "MP4-moov", "MP4-ilst":
		return Cyan
	default:
		return White
	}
}

// BoldSpecColor returns a bold color for the given spec type
func BoldSpecColor(dirName string) *color.Color {
	switch dirName {
	case "EXIF", "exif", "TIFF IFD0", "TIFF IFD1", "TIFF SubIFD":
		return BoldGreen
	case "IPTC", "iptc":
		return color.New(color.Bold, color.FgBlue)
	case "XMP", "xmp":
		return BoldCyan
	case "ICC", "icc":
		return BoldYellow
	case "ID3", "id3", "ID3v2.2", "ID3v2.3", "ID3v2.4":
		return BoldMagenta
	case "FLAC", "flac", "FLAC-StreamInfo", "FLAC-VorbisComment", "FLAC-Picture", "FLAC-Application", "FLAC-SeekTable", "FLAC-CueSheet":
		return BoldGreen
	case "MP4", "mp4", "MP4-ftyp", "MP4-moov", "MP4-ilst":
		return BoldCyan
	default:
		return Bold
	}
}
