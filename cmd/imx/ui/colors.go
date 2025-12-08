package ui

import (
	"github.com/fatih/color"
	"github.com/gomantics/imx"
)

var (
	// Text styles
	Bold = color.New(color.Bold)
	Dim  = color.New(color.Faint)

	// Colors
	Red    = color.New(color.FgRed)
	Green  = color.New(color.FgGreen)
	Yellow = color.New(color.FgYellow)
	Blue   = color.New(color.FgBlue)
	Cyan   = color.New(color.FgCyan)
	White  = color.New(color.FgWhite)

	// Combined styles
	BoldRed    = color.New(color.Bold, color.FgRed)
	BoldGreen  = color.New(color.Bold, color.FgGreen)
	BoldYellow = color.New(color.Bold, color.FgYellow)
	BoldCyan   = color.New(color.Bold, color.FgCyan)
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
func SpecColor(spec imx.Spec) *color.Color {
	switch spec.String() {
	case "exif":
		return Green
	case "iptc":
		return Blue
	case "xmp":
		return Cyan
	case "icc":
		return Yellow
	default:
		return White
	}
}

// BoldSpecColor returns a bold color for the given spec type
func BoldSpecColor(spec imx.Spec) *color.Color {
	switch spec.String() {
	case "exif":
		return BoldGreen
	case "iptc":
		return color.New(color.Bold, color.FgBlue)
	case "xmp":
		return BoldCyan
	case "icc":
		return BoldYellow
	default:
		return Bold
	}
}
