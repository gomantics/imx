package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/gomantics/imx"
	"github.com/gomantics/imx/cmd/imx/ui"
)

// SummaryFormatter outputs a curated summary of key metadata
type SummaryFormatter struct {
	config *Config
}

// Format writes summary output
func (f *SummaryFormatter) Format(w io.Writer, results []*Result) error {
	for i, result := range results {
		if i > 0 {
			fmt.Fprintln(w)
		}

		if err := f.formatSingle(w, result); err != nil {
			return err
		}
	}
	return nil
}

func (f *SummaryFormatter) formatSingle(w io.Writer, result *Result) error {
	// Print file name
	if !f.config.Quiet {
		if f.config.NoColor {
			fmt.Fprintf(w, "%s\n", result.File)
		} else {
			ui.Bold.Fprintf(w, "%s\n", result.File)
		}
	}

	// Handle errors
	if result.Error != nil {
		if f.config.NoColor {
			fmt.Fprintf(w, "Error: %v\n", result.Error)
		} else {
			ui.Red.Fprintf(w, "Error: %v\n", result.Error)
		}
		return nil
	}

	if result.Meta == nil {
		return nil
	}

	// Camera info
	cameraMake := f.getTagValue(result.Meta, "EXIF", "Make")
	cameraModel := f.getTagValue(result.Meta, "EXIF", "Model")
	if cameraMake != "" || cameraModel != "" {
		f.printField(w, "Camera", fmt.Sprintf("%s %s", cameraMake, cameraModel))
	}

	// Date
	date := f.getTagValue(result.Meta, "EXIF", "DateTimeOriginal")
	if date == "" {
		date = f.getTagValue(result.Meta, "EXIF", "DateTime")
	}
	if date != "" {
		// Format date if time format is specified
		if f.config.TimeFormat != "" && f.config.TimeFormat != "iso" {
			date = ui.FormatTime(date, f.config.TimeFormat)
		}
		f.printField(w, "Date", date)
	}

	// Dimensions
	width := f.getTagValue(result.Meta, "EXIF", "ImageWidth")
	height := f.getTagValue(result.Meta, "EXIF", "ImageHeight")
	if width == "" {
		width = f.getTagValue(result.Meta, "EXIF", "PixelXDimension")
	}
	if height == "" {
		height = f.getTagValue(result.Meta, "EXIF", "PixelYDimension")
	}
	if width != "" && height != "" {
		f.printField(w, "Dimensions", fmt.Sprintf("%s × %s", width, height))
	}

	// GPS
	lat := f.getRawTagValue(result.Meta, "EXIF", "GPSLatitude")
	lon := f.getRawTagValue(result.Meta, "EXIF", "GPSLongitude")
	latRef := f.getTagValue(result.Meta, "EXIF", "GPSLatitudeRef")
	lonRef := f.getTagValue(result.Meta, "EXIF", "GPSLongitudeRef")
	if lat != nil && lon != nil {
		gpsFormat := f.config.GPSFormat
		if gpsFormat == "" {
			gpsFormat = "dms"
		}
		gpsStr := ui.FormatGPS(lat, lon, latRef, lonRef, gpsFormat)
		f.printField(w, "GPS", gpsStr)
	}

	// Exposure
	exposure := f.getTagValue(result.Meta, "EXIF", "ExposureTime")
	fNumber := f.getTagValue(result.Meta, "EXIF", "FNumber")
	iso := f.getTagValue(result.Meta, "EXIF", "ISOSpeedRatings")
	if exposure != "" || fNumber != "" || iso != "" {
		var parts []string
		if exposure != "" {
			parts = append(parts, exposure+"s")
		}
		if fNumber != "" {
			parts = append(parts, "f/"+fNumber)
		}
		if iso != "" {
			parts = append(parts, "ISO "+iso)
		}
		f.printField(w, "Exposure", strings.Join(parts, "  "))
	}

	// Lens
	lens := f.getTagValue(result.Meta, "EXIF", "LensModel")
	if lens == "" {
		lens = f.getTagValue(result.Meta, "EXIF", "Lens")
	}
	if lens != "" {
		f.printField(w, "Lens", lens)
	}

	// Copyright
	copyright := f.getTagValue(result.Meta, "EXIF", "Copyright")
	if copyright == "" {
		copyright = f.getTagValue(result.Meta, "IPTC", "CopyrightNotice")
	}
	if copyright != "" {
		f.printField(w, "Copyright", copyright)
	}

	// Tag counts by directory
	counts := make(map[string]int)
	result.Meta.Each(func(dir imx.Directory, tag imx.Tag) bool {
		counts[strings.ToLower(dir.Name)]++
		return true
	})

	var dirParts []string
	for _, dirType := range []string{"exif", "iptc", "xmp", "icc"} {
		if c := counts[dirType]; c > 0 {
			dirPart := fmt.Sprintf("%s:%d", dirType, c)
			if !f.config.NoColor {
				dirColor := ui.SpecColor(dirType)
				dirPart = dirColor.Sprint(dirType) + fmt.Sprintf(":%d", c)
			}
			dirParts = append(dirParts, dirPart)
		}
	}
	if len(dirParts) > 0 {
		f.printField(w, "Tags", strings.Join(dirParts, "  "))
	}

	return nil
}

func (f *SummaryFormatter) printField(w io.Writer, label, value string) {
	if f.config.NoColor {
		fmt.Fprintf(w, "  %-12s %s\n", label+":", value)
	} else {
		ui.Dim.Fprintf(w, "  %-12s", label+":")
		fmt.Fprintf(w, " %s\n", value)
	}
}

func (f *SummaryFormatter) getTagValue(meta *imx.Metadata, dirName string, name string) string {
	var result string
	meta.Each(func(dir imx.Directory, tag imx.Tag) bool {
		if strings.EqualFold(dir.Name, dirName) && tag.Name == name {
			result = ui.FormatValue(tag.Value, true)
			return false
		}
		return true
	})
	return result
}

func (f *SummaryFormatter) getRawTagValue(meta *imx.Metadata, dirName string, name string) any {
	var result any
	meta.Each(func(dir imx.Directory, tag imx.Tag) bool {
		if strings.EqualFold(dir.Name, dirName) && tag.Name == name {
			result = tag.Value
			return false
		}
		return true
	})
	return result
}
