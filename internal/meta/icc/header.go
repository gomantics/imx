package icc

import (
	"encoding/binary"
	"fmt"
	"time"
)

const (
	// HeaderSize is the fixed size of an ICC profile header
	HeaderSize = 128

	// MinProfileSize is the minimum valid profile size (header only)
	MinProfileSize = HeaderSize

	// ICCSignature is the required signature at offset 36 ('acsp')
	ICCSignature = 0x61637370
)

// parseHeader parses the 128-byte ICC profile header
func parseHeader(data []byte) (*Header, error) {
	if len(data) < HeaderSize {
		return nil, fmt.Errorf("data too short for ICC header: %d bytes (need %d)", len(data), HeaderSize)
	}

	h := &Header{}

	// Bytes 0-3: Profile size
	h.ProfileSize = binary.BigEndian.Uint32(data[0:4])

	// Bytes 4-7: Preferred CMM type (4-char signature)
	h.PreferredCMM = string(data[4:8])

	// Bytes 8-11: Profile version
	// Major version in byte 8, minor/bugfix in high/low nibbles of byte 9
	h.Version = Version{
		Major:  data[8],
		Minor:  data[9] >> 4,
		BugFix: data[9] & 0x0F,
	}

	// Bytes 12-15: Profile/Device class
	h.ProfileClass = ProfileClass(binary.BigEndian.Uint32(data[12:16]))

	// Bytes 16-19: Data color space
	h.DataColorSpace = ColorSpace(binary.BigEndian.Uint32(data[16:20]))

	// Bytes 20-23: Profile Connection Space (PCS)
	h.PCS = ColorSpace(binary.BigEndian.Uint32(data[20:24]))

	// Bytes 24-35: Creation date/time (dateTimeNumber)
	h.Created = parseDateTimeNumber(data[24:36])

	// Bytes 36-39: Profile signature (should be 'acsp')
	sig := binary.BigEndian.Uint32(data[36:40])
	h.Signature = signatureToString(sig)
	if sig != ICCSignature {
		return nil, fmt.Errorf("invalid ICC signature: expected 'acsp', got '%s'", h.Signature)
	}

	// Bytes 40-43: Primary platform
	h.Platform = Platform(binary.BigEndian.Uint32(data[40:44]))

	// Bytes 44-47: Profile flags
	h.Flags = ProfileFlags(binary.BigEndian.Uint32(data[44:48]))

	// Bytes 48-51: Device manufacturer
	h.DeviceManufacturer = string(data[48:52])

	// Bytes 52-55: Device model
	h.DeviceModel = string(data[52:56])

	// Bytes 56-63: Device attributes
	h.DeviceAttributes = DeviceAttributes(binary.BigEndian.Uint64(data[56:64]))

	// Bytes 64-67: Rendering intent
	h.RenderingIntent = RenderingIntent(binary.BigEndian.Uint32(data[64:68]))

	// Bytes 68-79: PCS illuminant (XYZ, s15Fixed16Number format)
	h.PCSIlluminant = parseXYZNumber(data[68:80])

	// Bytes 80-83: Profile creator
	h.Creator = string(data[80:84])

	// Bytes 84-99: Profile ID (MD5 checksum, version 4+)
	h.ProfileID = make([]byte, 16)
	copy(h.ProfileID, data[84:100])

	// Bytes 100-127: Reserved (should be zeros)

	return h, nil
}

// parseDateTimeNumber parses a 12-byte dateTimeNumber
func parseDateTimeNumber(data []byte) time.Time {
	if len(data) < 12 {
		return time.Time{}
	}

	year := int(binary.BigEndian.Uint16(data[0:2]))
	month := int(binary.BigEndian.Uint16(data[2:4]))
	day := int(binary.BigEndian.Uint16(data[4:6]))
	hour := int(binary.BigEndian.Uint16(data[6:8]))
	minute := int(binary.BigEndian.Uint16(data[8:10]))
	second := int(binary.BigEndian.Uint16(data[10:12]))

	// Validate ranges
	if year == 0 || month < 1 || month > 12 || day < 1 || day > 31 {
		return time.Time{}
	}

	return time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)
}

// parseXYZNumber parses a 12-byte XYZ value (3 x s15Fixed16Number)
func parseXYZNumber(data []byte) XYZNumber {
	if len(data) < 12 {
		return XYZNumber{}
	}

	return XYZNumber{
		X: parseS15Fixed16(data[0:4]),
		Y: parseS15Fixed16(data[4:8]),
		Z: parseS15Fixed16(data[8:12]),
	}
}

// parseS15Fixed16 parses a signed 15.16 fixed-point number
func parseS15Fixed16(data []byte) float64 {
	val := int32(binary.BigEndian.Uint32(data))
	return float64(val) / 65536.0
}

// parseU16Fixed16 parses an unsigned 16.16 fixed-point number
func parseU16Fixed16(data []byte) float64 {
	val := binary.BigEndian.Uint32(data)
	return float64(val) / 65536.0
}

// parseU8Fixed8 parses an unsigned 8.8 fixed-point number
func parseU8Fixed8(data []byte) float64 {
	val := binary.BigEndian.Uint16(data)
	return float64(val) / 256.0
}
