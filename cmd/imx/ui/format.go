package ui

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// FormatValue formats any value for display
func FormatValue(v any, full bool) string {
	switch val := v.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(val)
	case []byte:
		if len(val) > 20 && !full {
			return fmt.Sprintf("(binary, %d bytes)", len(val))
		}
		return fmt.Sprintf("%x", val)
	case []any:
		parts := make([]string, len(val))
		for i, item := range val {
			parts[i] = FormatValue(item, full)
		}
		return strings.Join(parts, ", ")
	case []float64:
		parts := make([]string, len(val))
		for i, f := range val {
			parts[i] = formatFloat(f)
		}
		return strings.Join(parts, ", ")
	case []uint16:
		parts := make([]string, len(val))
		for i, n := range val {
			parts[i] = fmt.Sprintf("%d", n)
		}
		return strings.Join(parts, ", ")
	case []int:
		parts := make([]string, len(val))
		for i, n := range val {
			parts[i] = fmt.Sprintf("%d", n)
		}
		return strings.Join(parts, ", ")
	case float64:
		return formatFloat(val)
	case float32:
		return formatFloat(float64(val))
	default:
		return fmt.Sprintf("%v", v)
	}
}

// formatFloat formats a float without unnecessary trailing zeros
func formatFloat(f float64) string {
	// Check if it's effectively an integer
	if f == math.Floor(f) && !math.IsInf(f, 0) && !math.IsNaN(f) {
		return fmt.Sprintf("%.0f", f)
	}

	s := fmt.Sprintf("%.6f", f)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}

// FormatGPS formats GPS coordinates in the specified format
func FormatGPS(lat, lon any, latRef, lonRef, format string) string {
	latDec := toDecimalDegrees(lat, latRef)
	lonDec := toDecimalDegrees(lon, lonRef)

	if latDec == 0 && lonDec == 0 {
		return fmt.Sprintf("%v, %v", lat, lon)
	}

	switch format {
	case "url":
		return fmt.Sprintf("https://maps.google.com/maps?q=%f,%f", latDec, lonDec)
	case "decimal":
		return fmt.Sprintf("%.6f, %.6f", latDec, lonDec)
	default: // dms
		return fmt.Sprintf("%s, %s",
			toDMS(latDec, latRef == "S" || latDec < 0, true),
			toDMS(lonDec, lonRef == "W" || lonDec < 0, false))
	}
}

// toDecimalDegrees converts GPS coordinates to decimal degrees
func toDecimalDegrees(coord any, ref string) float64 {
	switch v := coord.(type) {
	case []float64:
		if len(v) >= 3 {
			dec := v[0] + v[1]/60 + v[2]/3600
			if ref == "S" || ref == "W" {
				dec = -dec
			}
			return dec
		}
	case float64:
		return v
	}
	return 0
}

// toDMS converts decimal degrees to degrees/minutes/seconds format
func toDMS(decimal float64, isNegative, isLat bool) string {
	if decimal < 0 {
		decimal = -decimal
	}

	d := int(decimal)
	m := int((decimal - float64(d)) * 60)
	s := (decimal - float64(d) - float64(m)/60) * 3600

	dir := ""
	if isLat {
		if isNegative {
			dir = "S"
		} else {
			dir = "N"
		}
	} else {
		if isNegative {
			dir = "W"
		} else {
			dir = "E"
		}
	}

	return fmt.Sprintf("%d°%d'%.2f\"%s", d, m, s, dir)
}

// FormatTime formats time values based on the specified format
func FormatTime(t any, format string) string {
	var timeVal time.Time

	switch v := t.(type) {
	case time.Time:
		timeVal = v
	case string:
		// Try parsing common EXIF time format: "2006:01:02 15:04:05"
		parsed, err := time.Parse("2006:01:02 15:04:05", v)
		if err != nil {
			// Try ISO format
			parsed, err = time.Parse(time.RFC3339, v)
			if err != nil {
				// Return original string if can't parse
				return v
			}
		}
		timeVal = parsed
	default:
		return fmt.Sprintf("%v", t)
	}

	switch format {
	case "iso", "rfc3339":
		return timeVal.Format(time.RFC3339)
	case "unix":
		return fmt.Sprintf("%d", timeVal.Unix())
	case "human":
		return timeVal.Format("Jan 2, 2006 3:04 PM")
	default:
		// Custom Go layout
		return timeVal.Format(format)
	}
}
