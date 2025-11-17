package imx

var (
	pngSignature  = [...]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	riffSignature = [...]byte{0x52, 0x49, 0x46, 0x46}
	webpSignature = [...]byte{0x57, 0x45, 0x42, 0x50}
)

// detectFormat identifies the image format by examining the magic bytes.
// It returns the format name as a string, or an empty string if the format is not recognized.
func detectFormat(magicBytes []byte) string {
	if len(magicBytes) < 2 {
		return ""
	}

	// JPEG: FF D8 FF
	if len(magicBytes) >= 3 && magicBytes[0] == 0xFF && magicBytes[1] == 0xD8 && magicBytes[2] == 0xFF {
		return "JPEG"
	}

	// PNG: 89 50 4E 47 0D 0A 1A 0A
	if hasPrefix(magicBytes, pngSignature[:]) {
		return "PNG"
	}

	// GIF: 47 49 46 38 37 61 (GIF87a) or 47 49 46 38 39 61 (GIF89a)
	if len(magicBytes) >= 6 {
		if magicBytes[0] == 0x47 && magicBytes[1] == 0x49 && magicBytes[2] == 0x46 &&
			magicBytes[3] == 0x38 && (magicBytes[4] == 0x37 || magicBytes[4] == 0x39) &&
			magicBytes[5] == 0x61 {
			return "GIF"
		}
	}

	// WebP: RIFF (52 49 46 46) ... WEBP (57 45 42 50)
	if len(magicBytes) >= 12 && hasPrefix(magicBytes, riffSignature[:]) &&
		magicBytes[8] == webpSignature[0] && magicBytes[9] == webpSignature[1] &&
		magicBytes[10] == webpSignature[2] && magicBytes[11] == webpSignature[3] {
		return "WebP"
	}

	// BMP: 42 4D (BM)
	if len(magicBytes) >= 2 && magicBytes[0] == 0x42 && magicBytes[1] == 0x4D {
		return "BMP"
	}

	return ""
}

func hasPrefix(buf, prefix []byte) bool {
	if len(buf) < len(prefix) {
		return false
	}
	for i, b := range prefix {
		if buf[i] != b {
			return false
		}
	}
	return true
}
