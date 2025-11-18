package imx

// Format represents a supported image format.
type Format string

const (
	FormatUnknown Format = ""
	FormatJPEG    Format = "JPEG"
	FormatPNG     Format = "PNG"
	FormatGIF     Format = "GIF"
	FormatWebP    Format = "WebP"
	FormatBMP     Format = "BMP"
)

// ColorSpace captures the color representation used by an image.
type ColorSpace string

const (
	ColorSpaceUnknown        ColorSpace = "Unknown"
	ColorSpaceRGB            ColorSpace = "RGB"
	ColorSpaceRGBA           ColorSpace = "RGBA"
	ColorSpaceCMYK           ColorSpace = "CMYK"
	ColorSpaceGrayscale      ColorSpace = "Grayscale"
	ColorSpaceGrayscaleAlpha ColorSpace = "GrayscaleAlpha"
	ColorSpaceIndexed        ColorSpace = "Indexed"
)

// ImageMetadata contains comprehensive metadata extracted from an image file.
type ImageMetadata struct {
	Format        Format                 `json:"format"`
	Width         int                    `json:"width"`
	Height        int                    `json:"height"`
	FileSize      int64                  `json:"fileSize"`
	ColorDepth    int                    `json:"colorDepth"`
	ColorSpace    ColorSpace             `json:"colorSpace"`
	HasICCProfile bool                   `json:"hasICCProfile"`
	EXIF          map[string]interface{} `json:"exif,omitempty"`
	Additional    map[string]interface{} `json:"additional,omitempty"`
}
