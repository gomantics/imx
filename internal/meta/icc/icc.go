package icc

import (
	"encoding/binary"
	"fmt"

	"github.com/gomantics/imx/internal/common"
)

// Parser implements meta.Parser for ICC color profiles
type Parser struct{}

// New creates a new ICC profile parser
func New() *Parser {
	return &Parser{}
}

// Spec returns the metadata spec this parser handles
func (p *Parser) Spec() common.Spec {
	return common.SpecICC
}

// Parse extracts ICC profile metadata from raw blocks
func (p *Parser) Parse(blocks []common.RawBlock) ([]common.Directory, error) {
	if len(blocks) == 0 {
		return nil, nil
	}

	// Reassemble ICC profile from potentially multiple segments
	profileData, _ := p.reassembleSegments(blocks)

	if len(profileData) == 0 {
		return nil, nil
	}

	var dirs []common.Directory

	for i, data := range profileData {
		profile, err := p.parseProfile(data)
		if err != nil {
			// Skip malformed profiles, continue with others
			continue
		}

		dir := p.buildDirectory(profile, i)
		dirs = append(dirs, dir)
	}

	return dirs, nil
}

// reassembleSegments reassembles ICC profile data from multiple APP2 segments
// JPEG splits large ICC profiles across multiple APP2 markers
func (p *Parser) reassembleSegments(blocks []common.RawBlock) ([][]byte, error) {
	type segmentInfo struct {
		segmentNum    int
		totalSegments int
		data          []byte
	}

	var segments []segmentInfo

	for _, block := range blocks {
		if block.Spec != common.SpecICC {
			continue
		}

		if len(block.Payload) < 2 {
			// Too short to have segment header, skip
			continue
		}

		// JPEG ICC segment header: segmentNum (1 byte) + totalSegments (1 byte)
		segmentNum := int(block.Payload[0])
		totalSegments := int(block.Payload[1])
		profileData := block.Payload[2:]

		// Validate segment numbers
		if segmentNum == 0 || totalSegments == 0 || segmentNum > totalSegments {
			// Invalid segmentation, try as complete profile
			if len(block.Payload) >= MinProfileSize && p.looksLikeICCHeader(block.Payload) {
				segments = append(segments, segmentInfo{
					segmentNum:    1,
					totalSegments: 1,
					data:          block.Payload,
				})
			}
			continue
		}

		segments = append(segments, segmentInfo{
			segmentNum:    segmentNum,
			totalSegments: totalSegments,
			data:          profileData,
		})
	}

	if len(segments) == 0 {
		return nil, nil
	}

	// Group segments by total count (segments with same totalSegments belong together)
	groups := make(map[int][]segmentInfo)
	for _, seg := range segments {
		groups[seg.totalSegments] = append(groups[seg.totalSegments], seg)
	}

	var profiles [][]byte

	for totalSegments, segs := range groups {
		if totalSegments == 1 && len(segs) >= 1 {
			// Single-segment profile(s)
			for _, seg := range segs {
				if len(seg.data) >= MinProfileSize {
					profiles = append(profiles, seg.data)
				}
			}
			continue
		}

		// Multi-segment profile - reassemble in order
		if len(segs) != totalSegments {
			// Incomplete, skip
			continue
		}

		// Sort by segment number and concatenate
		assembled := make([]byte, 0)
		complete := true
		for i := 1; i <= totalSegments; i++ {
			found := false
			for _, seg := range segs {
				if seg.segmentNum == i {
					assembled = append(assembled, seg.data...)
					found = true
					break
				}
			}
			if !found {
				complete = false
				break
			}
		}

		if complete && len(assembled) >= MinProfileSize {
			profiles = append(profiles, assembled)
		}
	}

	return profiles, nil
}

// looksLikeICCHeader checks if data starts with a valid ICC header
func (p *Parser) looksLikeICCHeader(data []byte) bool {
	if len(data) < 40 {
		return false
	}

	// Check for 'acsp' signature at offset 36
	sig := binary.BigEndian.Uint32(data[36:40])
	return sig == ICCSignature
}

// parseProfile parses a complete ICC profile
func (p *Parser) parseProfile(data []byte) (*Profile, error) {
	if len(data) < MinProfileSize {
		return nil, fmt.Errorf("profile too small: %d bytes", len(data))
	}

	header, err := parseHeader(data)
	if err != nil {
		return nil, fmt.Errorf("parse header: %w", err)
	}

	tags, err := parseTagTable(data)
	if err != nil {
		return nil, fmt.Errorf("parse tag table: %w", err)
	}

	return &Profile{
		Header: *header,
		Tags:   tags,
		Data:   data,
	}, nil
}

// buildDirectory converts a parsed profile into a common.Directory
func (p *Parser) buildDirectory(profile *Profile, index int) common.Directory {
	dir := common.Directory{
		Spec: common.SpecICC,
		Name: fmt.Sprintf("ICC Profile %d", index+1),
		Tags: make(map[common.TagID]common.Tag),
	}

	h := &profile.Header

	// Header fields
	p.addTag(&dir, "ProfileSize", "uint32", int(h.ProfileSize))
	p.addTag(&dir, "PreferredCMM", "string", trimNull(h.PreferredCMM))
	p.addTag(&dir, "Version", "string", h.Version.String())
	p.addTag(&dir, "ProfileClass", "string", h.ProfileClass.String())
	p.addTag(&dir, "ColorSpace", "string", h.DataColorSpace.String())
	p.addTag(&dir, "PCS", "string", h.PCS.String())

	if !h.Created.IsZero() {
		p.addTag(&dir, "CreateDate", "time", h.Created)
	}

	p.addTag(&dir, "Platform", "string", h.Platform.String())
	p.addTag(&dir, "RenderingIntent", "string", h.RenderingIntent.String())

	if trimNull(h.DeviceManufacturer) != "" && trimNull(h.DeviceManufacturer) != "\x00\x00\x00\x00" {
		p.addTag(&dir, "DeviceManufacturer", "string", trimNull(h.DeviceManufacturer))
	}
	if trimNull(h.DeviceModel) != "" && trimNull(h.DeviceModel) != "\x00\x00\x00\x00" {
		p.addTag(&dir, "DeviceModel", "string", trimNull(h.DeviceModel))
	}
	if trimNull(h.Creator) != "" && trimNull(h.Creator) != "\x00\x00\x00\x00" {
		p.addTag(&dir, "Creator", "string", trimNull(h.Creator))
	}

	// PCS Illuminant
	p.addTag(&dir, "PCSIlluminant", "xyz", []float64{
		h.PCSIlluminant.X,
		h.PCSIlluminant.Y,
		h.PCSIlluminant.Z,
	})

	// Flags - human readable
	p.addTag(&dir, "ProfileFlags", "string", formatFlags(h.Flags))

	// Device Attributes - human readable
	p.addTag(&dir, "DeviceAttributes", "string", formatDeviceAttributes(h.DeviceAttributes))

	// Profile ID (MD5 hash, v4+) - only if non-zero
	if !isZeroBytes(h.ProfileID) {
		p.addTag(&dir, "ProfileID", "hex", fmt.Sprintf("%x", h.ProfileID))
	}

	// Parse tag values
	for _, entry := range profile.Tags {
		parsed := parseTagValue(nil, entry, profile.Data)
		if parsed.Value == nil {
			continue
		}

		tagID := common.TagID("ICC:" + parsed.Name)

		// Skip if we already have this tag (from header)
		if _, exists := dir.Tags[tagID]; exists {
			continue
		}

		dir.Tags[tagID] = common.Tag{
			Spec:     common.SpecICC,
			ID:       tagID,
			Name:     parsed.Name,
			DataType: parsed.TypeSig,
			Value:    parsed.Value,
			Raw:      parsed.Raw,
		}
	}

	return dir
}

// addTag adds a tag to the directory
func (p *Parser) addTag(dir *common.Directory, name string, dataType string, value any) {
	id := common.TagID("ICC:" + name)
	dir.Tags[id] = common.Tag{
		Spec:     common.SpecICC,
		ID:       id,
		Name:     name,
		DataType: dataType,
		Value:    value,
	}
}

// trimNull removes null bytes from a string
func trimNull(s string) string {
	for i, c := range s {
		if c == 0 {
			return s[:i]
		}
	}
	return s
}

// isZeroBytes checks if all bytes are zero
func isZeroBytes(data []byte) bool {
	for _, b := range data {
		if b != 0 {
			return false
		}
	}
	return true
}

// formatFlags returns a human-readable string for profile flags
func formatFlags(f ProfileFlags) string {
	embedded := "Not Embedded"
	if f.IsEmbedded() {
		embedded = "Embedded"
	}
	independent := "Independent"
	if !f.IsIndependent() {
		independent = "Not Independent"
	}
	return embedded + ", " + independent
}

// formatDeviceAttributes returns a human-readable string for device attributes
func formatDeviceAttributes(a DeviceAttributes) string {
	var parts []string
	if a.IsReflective() {
		parts = append(parts, "Reflective")
	} else {
		parts = append(parts, "Transparency")
	}
	if a.IsGlossy() {
		parts = append(parts, "Glossy")
	} else {
		parts = append(parts, "Matte")
	}
	if a.IsPositive() {
		parts = append(parts, "Positive")
	} else {
		parts = append(parts, "Negative")
	}
	if a.IsColor() {
		parts = append(parts, "Color")
	} else {
		parts = append(parts, "Black & White")
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += ", " + parts[i]
	}
	return result
}
