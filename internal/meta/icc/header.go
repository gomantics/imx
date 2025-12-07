package icc

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/gomantics/imx/internal/common"
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
	profileSize, _ := common.ReadUint32(data, 0, binary.BigEndian)
	h.ProfileSize = profileSize

	// Bytes 4-7: Preferred CMM type (4-char signature)
	cmmSlice, _ := common.SafeSlice(data, 4, 4)
	h.PreferredCMM = string(cmmSlice)

	// Bytes 8-11: Profile version
	// Major version in byte 8, minor/bugfix in high/low nibbles of byte 9
	h.Version = Version{
		Major:  data[8],
		Minor:  data[9] >> 4,
		BugFix: data[9] & 0x0F,
	}

	// Bytes 12-15: Profile/Device class
	profileClass, _ := common.ReadUint32(data, 12, binary.BigEndian)
	h.ProfileClass = ProfileClass(profileClass)

	// Bytes 16-19: Data color space
	dataColorSpace, _ := common.ReadUint32(data, 16, binary.BigEndian)
	h.DataColorSpace = ColorSpace(dataColorSpace)

	// Bytes 20-23: Profile Connection Space (PCS)
	pcs, _ := common.ReadUint32(data, 20, binary.BigEndian)
	h.PCS = ColorSpace(pcs)

	// Bytes 24-35: Creation date/time (dateTimeNumber)
	dateSlice, _ := common.SafeSlice(data, 24, 12)
	h.Created = parseDateTimeNumber(dateSlice)

	// Bytes 36-39: Profile signature (should be 'acsp')
	sig, _ := common.ReadUint32(data, 36, binary.BigEndian)
	h.Signature = signatureToString(sig)
	if sig != ICCSignature {
		return nil, fmt.Errorf("invalid ICC signature: expected 'acsp', got '%s'", h.Signature)
	}

	// Bytes 40-43: Primary platform
	platform, _ := common.ReadUint32(data, 40, binary.BigEndian)
	h.Platform = Platform(platform)

	// Bytes 44-47: Profile flags
	flags, _ := common.ReadUint32(data, 44, binary.BigEndian)
	h.Flags = ProfileFlags(flags)

	// Bytes 48-51: Device manufacturer
	manufSlice, _ := common.SafeSlice(data, 48, 4)
	h.DeviceManufacturer = string(manufSlice)

	// Bytes 52-55: Device model
	modelSlice, _ := common.SafeSlice(data, 52, 4)
	h.DeviceModel = string(modelSlice)

	// Bytes 56-63: Device attributes
	devAttr, _ := common.ReadUint64(data, 56, binary.BigEndian)
	h.DeviceAttributes = DeviceAttributes(devAttr)

	// Bytes 64-67: Rendering intent
	renderIntent, _ := common.ReadUint32(data, 64, binary.BigEndian)
	h.RenderingIntent = RenderingIntent(renderIntent)

	// Bytes 68-79: PCS illuminant (XYZ, s15Fixed16Number format)
	xyzSlice, _ := common.SafeSlice(data, 68, 12)
	h.PCSIlluminant = parseXYZNumber(xyzSlice)

	// Bytes 80-83: Profile creator
	creatorSlice, _ := common.SafeSlice(data, 80, 4)
	h.Creator = string(creatorSlice)

	// Bytes 84-99: Profile ID (MD5 checksum, version 4+)
	profileIDSlice, _ := common.SafeSlice(data, 84, 16)
	h.ProfileID = make([]byte, 16)
	copy(h.ProfileID, profileIDSlice)

	// Bytes 100-127: Reserved (should be zeros)

	return h, nil
}

// parseDateTimeNumber parses a 12-byte dateTimeNumber
func parseDateTimeNumber(data []byte) time.Time {
	if len(data) < 12 {
		return time.Time{}
	}

	year16, _ := common.ReadUint16(data, 0, binary.BigEndian)
	month16, _ := common.ReadUint16(data, 2, binary.BigEndian)
	day16, _ := common.ReadUint16(data, 4, binary.BigEndian)
	hour16, _ := common.ReadUint16(data, 6, binary.BigEndian)
	minute16, _ := common.ReadUint16(data, 8, binary.BigEndian)
	second16, _ := common.ReadUint16(data, 10, binary.BigEndian)

	year := int(year16)
	month := int(month16)
	day := int(day16)
	hour := int(hour16)
	minute := int(minute16)
	second := int(second16)

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

	xSlice, _ := common.SafeSlice(data, 0, 4)
	ySlice, _ := common.SafeSlice(data, 4, 4)
	zSlice, _ := common.SafeSlice(data, 8, 4)

	x, _ := common.ParseS15Fixed16(xSlice)
	y, _ := common.ParseS15Fixed16(ySlice)
	z, _ := common.ParseS15Fixed16(zSlice)

	return XYZNumber{
		X: x,
		Y: y,
		Z: z,
	}
}
