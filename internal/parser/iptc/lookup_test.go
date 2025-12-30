package iptc

import (
	"encoding/binary"
	"testing"
)

func TestGetDatasetInfo(t *testing.T) {
	tests := []struct {
		name      string
		record    Record
		datasetID uint8
		wantName  string
		wantRep   bool
	}{
		// Envelope record
		{"Envelope RecordVersion", RecordEnvelope, 0, "RecordVersion", false},
		{"Envelope Destination", RecordEnvelope, 5, "Destination", true},
		{"Envelope DateSent", RecordEnvelope, 70, "DateSent", false},

		// Application record
		{"App RecordVersion", RecordApplication, 0, "RecordVersion", false},
		{"App ObjectName", RecordApplication, 5, "ObjectName", false},
		{"App Keywords", RecordApplication, 25, "Keywords", true},
		{"App Byline", RecordApplication, 80, "Byline", true},
		{"App City", RecordApplication, 90, "City", false},

		// NewsPhoto record
		{"NewsPhoto RecordVersion", RecordNewsPhoto, 0, "RecordVersion", false},
		{"NewsPhoto PictureNumber", RecordNewsPhoto, 5, "PictureNumber", false},

		// PreObjectData record
		{"PreObjectData SizeMode", RecordPreObjectData, 10, "SizeMode", false},

		// ObjectData record
		{"ObjectData SubFile", RecordObjectData, 10, "SubFile", true},

		// PostObjectData record
		{"PostObjectData Confirmed", RecordPostObjectData, 10, "ConfirmedObjectDataSize", false},

		// Unknown record
		{"Unknown record", Record(99), 5, "", false},

		// Unknown dataset ID
		{"Unknown dataset in known record", RecordApplication, 255, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getDatasetInfo(tt.record, tt.datasetID)
			if got.Name != tt.wantName {
				t.Errorf("getDatasetInfo() Name = %q, want %q", got.Name, tt.wantName)
			}
			if got.Repeatable != tt.wantRep {
				t.Errorf("getDatasetInfo() Repeatable = %v, want %v", got.Repeatable, tt.wantRep)
			}
		})
	}
}

func TestGetDatasetName(t *testing.T) {
	tests := []struct {
		name      string
		record    Record
		datasetID uint8
		want      string
	}{
		{"Known dataset", RecordApplication, 25, "Keywords"},
		{"Unknown dataset", RecordApplication, 255, "Dataset2:255"},
		{"Unknown record", Record(99), 5, "Dataset99:5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getDatasetName(tt.record, tt.datasetID); got != tt.want {
				t.Errorf("getDatasetName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsRepeatable(t *testing.T) {
	tests := []struct {
		name      string
		record    Record
		datasetID uint8
		want      bool
	}{
		{"Keywords - repeatable", RecordApplication, 25, true},
		{"ObjectName - not repeatable", RecordApplication, 5, false},
		{"Unknown - not repeatable", RecordApplication, 255, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRepeatable(tt.record, tt.datasetID); got != tt.want {
				t.Errorf("isRepeatable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseDatasetValue(t *testing.T) {
	t.Run("Application RecordVersion", func(t *testing.T) {
		data := make([]byte, 2)
		binary.BigEndian.PutUint16(data, 4)
		got := parseDatasetValue(RecordApplication, 0, data)
		if v, ok := got.(int); !ok || v != 4 {
			t.Errorf("parseDatasetValue() = %v, want int(4)", got)
		}
	})

	t.Run("Application RecordVersion short data", func(t *testing.T) {
		data := []byte{0}
		got := parseDatasetValue(RecordApplication, 0, data)
		if _, ok := got.(string); !ok {
			t.Errorf("parseDatasetValue() = %T, want string", got)
		}
	})

	t.Run("Application Urgency", func(t *testing.T) {
		data := []byte{'5'}
		got := parseDatasetValue(RecordApplication, 10, data)
		if v, ok := got.(int); !ok || v != 5 {
			t.Errorf("parseDatasetValue() = %v, want int(5)", got)
		}
	})

	t.Run("Application Urgency invalid", func(t *testing.T) {
		data := []byte{'x'}
		got := parseDatasetValue(RecordApplication, 10, data)
		if _, ok := got.(string); !ok {
			t.Errorf("parseDatasetValue() = %T, want string", got)
		}
	})

	t.Run("Application DateCreated", func(t *testing.T) {
		data := []byte("20240115")
		got := parseDatasetValue(RecordApplication, 55, data)
		if s, ok := got.(string); !ok || s != "2024-01-15" {
			t.Errorf("parseDatasetValue() = %v, want '2024-01-15'", got)
		}
	})

	t.Run("Application DigitalCreationDate", func(t *testing.T) {
		data := []byte("20240115")
		got := parseDatasetValue(RecordApplication, 62, data)
		if s, ok := got.(string); !ok || s != "2024-01-15" {
			t.Errorf("parseDatasetValue() = %v, want '2024-01-15'", got)
		}
	})

	t.Run("Application TimeCreated", func(t *testing.T) {
		data := []byte("143025+0500")
		got := parseDatasetValue(RecordApplication, 60, data)
		if s, ok := got.(string); !ok || s != "14:30:25+05:00" {
			t.Errorf("parseDatasetValue() = %v, want '14:30:25+05:00'", got)
		}
	})

	t.Run("Application DigitalCreationTime", func(t *testing.T) {
		data := []byte("143025")
		got := parseDatasetValue(RecordApplication, 63, data)
		if s, ok := got.(string); !ok || s != "14:30:25" {
			t.Errorf("parseDatasetValue() = %v, want '14:30:25'", got)
		}
	})

	t.Run("Application ReleaseDate", func(t *testing.T) {
		data := []byte("20240115")
		got := parseDatasetValue(RecordApplication, 30, data)
		if s, ok := got.(string); !ok || s != "2024-01-15" {
			t.Errorf("parseDatasetValue() = %v, want '2024-01-15'", got)
		}
	})

	t.Run("Application ExpirationDate", func(t *testing.T) {
		data := []byte("20240115")
		got := parseDatasetValue(RecordApplication, 37, data)
		if s, ok := got.(string); !ok || s != "2024-01-15" {
			t.Errorf("parseDatasetValue() = %v, want '2024-01-15'", got)
		}
	})

	t.Run("Application ReleaseTime", func(t *testing.T) {
		data := []byte("143025")
		got := parseDatasetValue(RecordApplication, 35, data)
		if s, ok := got.(string); !ok || s != "14:30:25" {
			t.Errorf("parseDatasetValue() = %v, want '14:30:25'", got)
		}
	})

	t.Run("Application ExpirationTime", func(t *testing.T) {
		data := []byte("143025")
		got := parseDatasetValue(RecordApplication, 38, data)
		if s, ok := got.(string); !ok || s != "14:30:25" {
			t.Errorf("parseDatasetValue() = %v, want '14:30:25'", got)
		}
	})

	t.Run("Application Prefs", func(t *testing.T) {
		data := []byte("1:2:3:4")
		got := parseDatasetValue(RecordApplication, 221, data)
		if s, ok := got.(string); !ok || s == "" {
			t.Errorf("parseDatasetValue() = %v, want non-empty string", got)
		}
	})

	t.Run("Envelope RecordVersion", func(t *testing.T) {
		data := make([]byte, 2)
		binary.BigEndian.PutUint16(data, 2)
		got := parseDatasetValue(RecordEnvelope, 0, data)
		if v, ok := got.(int); !ok || v != 2 {
			t.Errorf("parseDatasetValue() = %v, want int(2)", got)
		}
	})

	t.Run("Envelope DateSent", func(t *testing.T) {
		data := []byte("20240115")
		got := parseDatasetValue(RecordEnvelope, 70, data)
		if s, ok := got.(string); !ok || s != "2024-01-15" {
			t.Errorf("parseDatasetValue() = %v, want '2024-01-15'", got)
		}
	})

	t.Run("Envelope TimeSent", func(t *testing.T) {
		data := []byte("143025")
		got := parseDatasetValue(RecordEnvelope, 80, data)
		if s, ok := got.(string); !ok || s != "14:30:25" {
			t.Errorf("parseDatasetValue() = %v, want '14:30:25'", got)
		}
	})

	t.Run("Default string value", func(t *testing.T) {
		data := []byte("test value\x00")
		got := parseDatasetValue(RecordApplication, 120, data)
		if s, ok := got.(string); !ok || s != "test value" {
			t.Errorf("parseDatasetValue() = %v, want 'test value'", got)
		}
	})
}

func TestParseDateString(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{"Valid 8-digit date", []byte("20240115"), "2024-01-15"},
		{"Short date", []byte("2024"), "2024"},
		{"Empty", []byte(""), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseDateString(tt.data); got != tt.want {
				t.Errorf("parseDateString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseTimeString(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{"Time without timezone", []byte("143025"), "14:30:25"},
		{"Time with timezone", []byte("143025+0500"), "14:30:25+05:00"},
		{"Time with negative timezone", []byte("143025-0800"), "14:30:25-08:00"},
		{"Short time", []byte("14"), "14"},
		{"Empty", []byte(""), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseTimeString(tt.data); got != tt.want {
				t.Errorf("parseTimeString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParsePrefs(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			"Valid prefs",
			[]byte("1:2:3:4"),
			"Tagged:1, ColorClass:2, Rating:3, FrameNum:4",
		},
		{
			"Prefs with null bytes",
			[]byte("1:2:3:4\x00\x00"),
			"Tagged:1, ColorClass:2, Rating:3, FrameNum:4",
		},
		{
			"Insufficient parts",
			[]byte("1:2"),
			"1:2",
		},
		{
			"Empty",
			[]byte(""),
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parsePrefs(tt.data); got != tt.want {
				t.Errorf("parsePrefs() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTrimNullBytes(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{"No null bytes", []byte("test"), "test"},
		{"Trailing null bytes", []byte("test\x00\x00"), "test"},
		{"Leading and trailing spaces", []byte("  test  \x00"), "test"},
		{"Only null bytes", []byte("\x00\x00\x00"), ""},
		{"Empty", []byte(""), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := trimNullBytes(tt.data); got != tt.want {
				t.Errorf("trimNullBytes() = %q, want %q", got, tt.want)
			}
		})
	}
}
