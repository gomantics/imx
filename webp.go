package imx

import (
	"fmt"
	"io"
)

// ExtractWebPMetadata extracts metadata from a WebP file.
func ExtractWebPMetadata(r io.ReadSeeker, md *ImageMetadata) error {
	// Reset to beginning
	_, err := r.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	// Read RIFF header (12 bytes)
	var header [12]byte
	if _, err = io.ReadFull(r, header[:]); err != nil {
		return fmt.Errorf("failed to read WebP header: %w", err)
	}

	// Verify RIFF signature
	if string(header[0:4]) != "RIFF" {
		return fmt.Errorf("invalid WebP file: missing RIFF signature")
	}

	// Verify WEBP signature
	if string(header[8:12]) != "WEBP" {
		return fmt.Errorf("invalid WebP file: missing WEBP signature")
	}

	// Read chunk type (4 bytes)
	var chunkType [4]byte
	if _, err = io.ReadFull(r, chunkType[:]); err != nil {
		return fmt.Errorf("failed to read WebP chunk type: %w", err)
	}
	chunkTypeStr := string(chunkType[:])

	hasAnimation := false
	hasAlpha := false

	// Handle different WebP formats
	switch chunkTypeStr {
	case "VP8 ":
		// Simple lossy format
		err = parseVP8(r, md)
		if err != nil {
			return err
		}

	case "VP8L":
		// Lossless format
		err = parseVP8L(r, md)
		if err != nil {
			return err
		}
		hasAlpha = true // VP8L supports alpha

	case "VP8X":
		// Extended format (supports animation, alpha, etc.)
		anim, alpha, parseErr := parseVP8X(r, md)
		if parseErr != nil {
			return parseErr
		}
		hasAnimation = anim
		hasAlpha = alpha
	default:
		return fmt.Errorf("unsupported WebP chunk type: %s", chunkTypeStr)
	}

	md.ColorSpace = "RGB"
	if hasAlpha {
		md.ColorSpace = "RGBA"
	}
	md.setAdditional("HasAnimation", hasAnimation)
	md.setAdditional("HasAlpha", hasAlpha)

	return nil
}

// parseVP8 parses a simple VP8 (lossy) WebP chunk
func parseVP8(r io.ReadSeeker, md *ImageMetadata) error {
	// Read chunk size (already read, but we need to skip it)
	// VP8 format: 3 bytes key frame header, then dimensions
	var keyFrame [10]byte
	if _, err := io.ReadFull(r, keyFrame[:]); err != nil {
		return fmt.Errorf("failed to read VP8 key frame: %w", err)
	}

	// Verify key frame signature
	if keyFrame[0] != 0x9D || keyFrame[1] != 0x01 || keyFrame[2] != 0x2A {
		return fmt.Errorf("invalid VP8 key frame")
	}

	// Dimensions are in the key frame (14-bit values, little-endian)
	width := int(keyFrame[6]) | (int(keyFrame[7]&0x3F) << 8)
	height := int(keyFrame[8]) | (int(keyFrame[9]&0x3F) << 8)

	md.Width = width + 1
	md.Height = height + 1
	md.ColorDepth = 24 // VP8 is always 24-bit RGB

	return nil
}

// parseVP8L parses a VP8L (lossless) WebP chunk
func parseVP8L(r io.ReadSeeker, md *ImageMetadata) error {
	// Read VP8L header (5 bytes)
	var header [5]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return fmt.Errorf("failed to read VP8L header: %w", err)
	}

	// Verify VP8L signature
	if header[0] != 0x2F {
		return fmt.Errorf("invalid VP8L signature")
	}

	// Dimensions are encoded in the header (14 bits each)
	width := int(header[1]) | (int(header[2]&0x3F) << 8)
	height := int(header[2]>>6) | (int(header[3]&0xF) << 2) | (int(header[4]&0x3F) << 10)

	md.Width = width + 1
	md.Height = height + 1
	md.ColorDepth = 32 // VP8L supports alpha, so 32-bit RGBA

	return nil
}

// parseVP8X parses a VP8X (extended) WebP chunk
func parseVP8X(r io.ReadSeeker, md *ImageMetadata) (bool, bool, error) {
	// Read VP8X header (10 bytes)
	var header [10]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return false, false, fmt.Errorf("failed to read VP8X header: %w", err)
	}

	// Verify VP8X signature
	if string(header[0:4]) != "VP8X" {
		return false, false, fmt.Errorf("invalid VP8X signature")
	}

	// Flags (1 byte)
	flags := header[4]

	// Dimensions (3 bytes width, 3 bytes height, stored as 24-bit little-endian)
	// Width: bytes 6-8 (24 bits)
	width := int(header[6]) | (int(header[7]) << 8) | (int(header[8]&0x0F) << 16)
	// Height: upper 4 bits of byte 8, bytes 9-10 (20 bits total, but we use 24)
	height := int(header[8]>>4) | (int(header[9]) << 4)

	md.Width = width + 1
	md.Height = height + 1
	md.ColorDepth = 24
	if (flags & 0x10) != 0 {
		md.ColorDepth = 32 // Has alpha
	}

	md.setAdditional("Reserved", (flags&0xE0)>>5)
	md.setAdditional("ICC", (flags&0x20) != 0)
	md.setAdditional("Alpha", (flags&0x10) != 0)
	md.setAdditional("EXIF", (flags&0x08) != 0)
	md.setAdditional("XMP", (flags&0x04) != 0)
	md.setAdditional("Animation", (flags&0x02) != 0)

	// Check for ICC profile
	if (flags & 0x20) != 0 {
		md.HasICCProfile = true
	}

	return (flags & 0x02) != 0, (flags & 0x10) != 0, nil
}
