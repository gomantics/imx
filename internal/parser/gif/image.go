package gif

import (
	"io"
)

// skipImage skips over a GIF image data block
func skipImage(r io.ReaderAt, pos int64, buf *[11]byte) (int64, bool) {
	// Read Image Descriptor (9 bytes)
	_, err := r.ReadAt(buf[:imageDescriptorSize], pos)
	if err != nil {
		return pos, false
	}

	pos += imageDescriptorSize

	// Check for Local Color Table
	packed := buf[8]
	hasLCT := (packed & maskGlobalColorTable) != 0
	if hasLCT {
		lctSize := 1 << ((packed & maskColorTableSize) + 1)
		pos += int64(lctSize * colorTableEntrySize)
	}

	// Skip LZW minimum code size
	pos++

	// Skip image data sub-blocks
	pos = skipDataSubBlocks(r, pos, buf)

	return pos, true
}

// countFrames counts the number of image frames in a GIF
// Also extracts animation loop count from NETSCAPE2.0 extension if present
func countFrames(r io.ReaderAt, startPos int64, buf *[11]byte) (int, int) {
	pos := startPos
	frameCount := 0
	loopCount := 0 // 0 means loop forever

	for {
		_, err := r.ReadAt(buf[:1], pos)
		if err != nil {
			break
		}

		separator := buf[0]
		pos++

		switch separator {
		case separatorExtension:
			// Read extension label
			_, err := r.ReadAt(buf[:1], pos)
			if err != nil {
				return frameCount, loopCount
			}

			label := buf[0]
			pos++

			if label == labelApplicationExt {
				// Read block size
				_, err := r.ReadAt(buf[:1], pos)
				if err != nil {
					return frameCount, loopCount
				}
				blockSize := buf[0]
				pos++

				if blockSize == applicationExtBlockSize {
					// Read application identifier
					_, err := r.ReadAt(buf[:applicationExtBlockSize], pos)
					if err != nil {
						return frameCount, loopCount
					}
					pos += applicationExtBlockSize

					appID := string(buf[0:applicationIDLength])
					authCode := string(buf[applicationIDLength:applicationExtBlockSize])

					// Check for NETSCAPE2.0 extension (animation loop count)
					if appID == netscapeApplicationID && authCode == netscapeAuthCode {
						// Read sub-block
						_, err := r.ReadAt(buf[:netscapeSubBlockSize], pos)
						if err == nil && buf[0] == netscapeSubBlockSize {
							// buf[1] should be 1 (sub-block ID)
							// buf[2] and next byte are loop count (little-endian uint16)
							var loopBuf [2]byte
							r.ReadAt(loopBuf[:], pos+netscapeLoopCountOffset)
							loopCount = int(loopBuf[0]) | (int(loopBuf[1]) << 8)
							pos = skipDataSubBlocks(r, pos, buf)
						} else {
							pos = skipDataSubBlocks(r, pos, buf)
						}
					} else {
						pos = skipDataSubBlocks(r, pos, buf)
					}
				} else {
					pos = skipDataSubBlocks(r, pos, buf)
				}
			} else {
				// Other extensions, skip them
				pos = skipDataSubBlocks(r, pos, buf)
			}

		case separatorImageDescriptor:
			frameCount++
			var ok bool
			pos, ok = skipImage(r, pos, buf)
			if !ok {
				return frameCount, loopCount
			}

		case separatorTrailer:
			return frameCount, loopCount

		case separatorBlockTerminator:
			// Continue

		default:
			// Unknown separator, stop
			return frameCount, loopCount
		}
	}

	return frameCount, loopCount
}
