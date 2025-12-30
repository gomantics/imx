package flac

// STREAMINFO field offsets
// Reference: FLAC specification, Section 4.2.1
const (
	streamInfoMinBlockSizeOffset = 0
	streamInfoMaxBlockSizeOffset = 2
	streamInfoMinFrameSizeOffset = 4
	streamInfoMaxFrameSizeOffset = 7
	streamInfoSampleRateOffset   = 10
	streamInfoChannelsOffset     = 12
	streamInfoBitsPerSampleStart = 12
	streamInfoBitsPerSampleEnd   = 13
	streamInfoTotalSamplesStart  = 13
	streamInfoTotalSamplesEnd    = 18
	streamInfoMD5Offset          = 18
	streamInfoMD5Size            = 16
	streamInfoMinSize            = 34
)

// Block size limits
const (
	// maxBlockSize is the maximum reasonable metadata block size (8MB)
	// This prevents excessive memory allocation from malformed files
	// Note: FLAC block length field is 24 bits, so absolute max is 16,777,215 bytes
	maxBlockSize = 8 * 1024 * 1024
)

// Seek table constants
const (
	seekPointSize = 18 // Each seek point is 18 bytes
)

// Application block constants
const (
	applicationIDSize = 4 // Application ID is always 4 bytes
)
