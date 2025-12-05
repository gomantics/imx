package imx

// Import parsers to register them at init time
import (
	_ "github.com/gomantics/imx/internal/container/jpeg"
	_ "github.com/gomantics/imx/internal/meta/exif"
)
