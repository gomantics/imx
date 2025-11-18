package formats

// Detect identifies the image format by examining the magic bytes.
// It returns the format name as a string, or an empty string if the format is not recognized.
func Detect(magicBytes []byte) string {
	if len(magicBytes) < 2 {
		return ""
	}

	// JPEG: FF D8 FF
	if len(magicBytes) >= 3 && magicBytes[0] == 0xFF && magicBytes[1] == 0xD8 && magicBytes[2] == 0xFF {
		return "JPEG"
	}

	// PNG: 89 50 4E 47 0D 0A 1A 0A
	if len(magicBytes) >= 8 {
		pngSig := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
		match := true
		for i := 0; i < 8; i++ {
			if magicBytes[i] != pngSig[i] {
				match = false
				break
			}
		}
		if match {
			return "PNG"
		}
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
	if len(magicBytes) >= 12 {
		if magicBytes[0] == 0x52 && magicBytes[1] == 0x49 && magicBytes[2] == 0x46 && magicBytes[3] == 0x46 {
			// Check for WEBP at offset 8
			if len(magicBytes) >= 12 && magicBytes[8] == 0x57 && magicBytes[9] == 0x45 &&
				magicBytes[10] == 0x42 && magicBytes[11] == 0x50 {
				return "WebP"
			}
		}
	}

	// BMP: 42 4D (BM)
	if len(magicBytes) >= 2 && magicBytes[0] == 0x42 && magicBytes[1] == 0x4D {
		return "BMP"
	}

	return ""
}
