package formats

import (
	"fmt"
	"io"
)

// ExtractWebP extracts metadata from a WebP file.
func ExtractWebP(r io.ReadSeeker) (*Result, error) {
	// Reset to beginning
	_, err := r.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	// Read RIFF header (12 bytes)
	header := make([]byte, 12)
	_, err = r.Read(header)
	if err != nil {
		return nil, fmt.Errorf("failed to read WebP header: %w", err)
	}

	// Verify RIFF signature
	if string(header[0:4]) != "RIFF" {
		return nil, fmt.Errorf("%w: missing RIFF signature", ErrInvalidData)
	}

	// Verify WEBP signature
	if string(header[8:12]) != "WEBP" {
		return nil, fmt.Errorf("%w: missing WEBP signature", ErrInvalidData)
	}

	// Read chunk type (4 bytes)
	chunkType := make([]byte, 4)
	_, err = r.Read(chunkType)
	if err != nil {
		return nil, fmt.Errorf("failed to read WebP chunk type: %w", err)
	}
	chunkTypeStr := string(chunkType)

	hasAnimation := false
	hasAlpha := false
	result := newResult()

	// Handle different WebP formats
	switch chunkTypeStr {
	case "VP8 ":
		// Simple lossy format
		err = parseVP8(r, result)
		if err != nil {
			return nil, err
		}

	case "VP8L":
		// Lossless format
		err = parseVP8L(r, result)
		if err != nil {
			return nil, err
		}
		hasAlpha = true // VP8L supports alpha

	case "VP8X":
		// Extended format (supports animation, alpha, etc.)
		err = parseVP8X(r, result)
		if err != nil {
			return nil, err
		}
		// Extract animation and alpha from additional metadata
		if anim, ok := result.Additional["Animation"].(bool); ok {
			hasAnimation = anim
		}
		if alpha, ok := result.Additional["Alpha"].(bool); ok {
			hasAlpha = alpha
		}

	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedFormat, chunkTypeStr)
	}

	result.ColorSpace = "RGB"
	if hasAlpha {
		result.ColorSpace = "RGBA"
	}
	result.Additional["HasAnimation"] = hasAnimation
	result.Additional["HasAlpha"] = hasAlpha

	return result, nil
}

// parseVP8 parses a simple VP8 (lossy) WebP chunk
func parseVP8(r io.ReadSeeker, res *Result) error {
	// Read chunk size (already read, but we need to skip it)
	// VP8 format: 3 bytes key frame header, then dimensions
	keyFrame := make([]byte, 10)
	_, err := r.Read(keyFrame)
	if err != nil {
		return fmt.Errorf("failed to read VP8 key frame: %w", err)
	}

	// Verify key frame signature
	if keyFrame[0] != 0x9D || keyFrame[1] != 0x01 || keyFrame[2] != 0x2A {
		return fmt.Errorf("%w: invalid VP8 key frame", ErrInvalidData)
	}

	// Dimensions are in the key frame (14-bit values, little-endian)
	width := int(keyFrame[6]) | (int(keyFrame[7]&0x3F) << 8)
	height := int(keyFrame[8]) | (int(keyFrame[9]&0x3F) << 8)

	res.Width = width + 1
	res.Height = height + 1
	res.ColorDepth = 24 // VP8 is always 24-bit RGB

	return nil
}

// parseVP8L parses a VP8L (lossless) WebP chunk
func parseVP8L(r io.ReadSeeker, res *Result) error {
	// Read VP8L header (5 bytes)
	header := make([]byte, 5)
	_, err := r.Read(header)
	if err != nil {
		return fmt.Errorf("failed to read VP8L header: %w", err)
	}

	// Verify VP8L signature
	if header[0] != 0x2F {
		return fmt.Errorf("%w: invalid VP8L signature", ErrInvalidData)
	}

	// Dimensions are encoded in the header (14 bits each)
	width := int(header[1]) | (int(header[2]&0x3F) << 8)
	height := int(header[2]>>6) | (int(header[3]&0xF) << 2) | (int(header[4]&0x3F) << 10)

	res.Width = width + 1
	res.Height = height + 1
	res.ColorDepth = 32 // VP8L supports alpha, so 32-bit RGBA

	return nil
}

// parseVP8X parses a VP8X (extended) WebP chunk
func parseVP8X(r io.ReadSeeker, res *Result) error {
	// Read VP8X header (10 bytes)
	header := make([]byte, 10)
	_, err := r.Read(header)
	if err != nil {
		return fmt.Errorf("failed to read VP8X header: %w", err)
	}

	// Verify VP8X signature
	if string(header[0:4]) != "VP8X" {
		return fmt.Errorf("%w: invalid VP8X signature", ErrInvalidData)
	}

	// Flags (1 byte)
	flags := header[4]

	// Dimensions (3 bytes width, 3 bytes height, stored as 24-bit little-endian)
	// Width: bytes 6-8 (24 bits)
	width := int(header[6]) | (int(header[7]) << 8) | (int(header[8]&0x0F) << 16)
	// Height: upper 4 bits of byte 8, bytes 9-10 (20 bits total, but we use 24)
	height := int(header[8]>>4) | (int(header[9]) << 4)

	res.Width = width + 1
	res.Height = height + 1
	res.ColorDepth = 24
	if (flags & 0x10) != 0 {
		res.ColorDepth = 32 // Has alpha
	}

	res.Additional["Reserved"] = (flags & 0xE0) >> 5
	res.Additional["ICC"] = (flags & 0x20) != 0
	res.Additional["Alpha"] = (flags & 0x10) != 0
	res.Additional["EXIF"] = (flags & 0x08) != 0
	res.Additional["XMP"] = (flags & 0x04) != 0
	res.Additional["Animation"] = (flags & 0x02) != 0

	// Check for ICC profile
	if (flags & 0x20) != 0 {
		res.HasICCProfile = true
	}

	return nil
}
