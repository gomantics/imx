package iptc

import "testing"

func TestGetDatasetInfo(t *testing.T) {
	tests := []struct {
		name       string
		record     Record
		datasetID  uint8
		wantName   string
		wantRepeat bool
	}{
		// Envelope record
		{"Envelope RecordVersion", RecordEnvelope, 0, "RecordVersion", false},
		{"Envelope Destination", RecordEnvelope, 5, "Destination", true},
		{"Envelope DateSent", RecordEnvelope, 70, "DateSent", false},
		{"Envelope Unknown", RecordEnvelope, 255, "", false},

		// Application record
		{"App RecordVersion", RecordApplication, 0, "RecordVersion", false},
		{"App ObjectName", RecordApplication, 5, "ObjectName", false},
		{"App Keywords", RecordApplication, 25, "Keywords", true},
		{"App Byline", RecordApplication, 80, "Byline", true},
		{"App City", RecordApplication, 90, "City", false},
		{"App Caption", RecordApplication, 120, "Caption-Abstract", false},
		{"App Prefs", RecordApplication, 221, "Prefs", false},
		{"App SceneCode", RecordApplication, 240, "SceneCode", true},
		{"App Unknown", RecordApplication, 255, "", false},

		// NewsPhoto record
		{"NewsPhoto RecordVersion", RecordNewsPhoto, 0, "RecordVersion", false},
		{"NewsPhoto PictureNumber", RecordNewsPhoto, 5, "PictureNumber", false},
		{"NewsPhoto PixelsPerLine", RecordNewsPhoto, 10, "PixelsPerLine", false},
		{"NewsPhoto Unknown", RecordNewsPhoto, 255, "", false},

		// PreObjectData record
		{"PreObject SizeMode", RecordPreObjectData, 10, "SizeMode", false},
		{"PreObject Unknown", RecordPreObjectData, 255, "", false},

		// ObjectData record
		{"ObjectData SubFile", RecordObjectData, 10, "SubFile", true},
		{"ObjectData Unknown", RecordObjectData, 255, "", false},

		// PostObjectData record
		{"PostObject ConfirmedSize", RecordPostObjectData, 10, "ConfirmedObjectDataSize", false},
		{"PostObject Unknown", RecordPostObjectData, 255, "", false},

		// Unknown record
		{"Unknown Record", Record(99), 0, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := getDatasetInfo(tt.record, tt.datasetID)
			if info.Name != tt.wantName {
				t.Errorf("getDatasetInfo().Name = %q, want %q", info.Name, tt.wantName)
			}
			if info.Repeatable != tt.wantRepeat {
				t.Errorf("getDatasetInfo().Repeatable = %v, want %v", info.Repeatable, tt.wantRepeat)
			}
		})
	}
}

func TestGetDatasetName(t *testing.T) {
	tests := []struct {
		record    Record
		datasetID uint8
		want      string
	}{
		{RecordApplication, 5, "ObjectName"},
		{RecordApplication, 80, "Byline"},
		{RecordApplication, 255, ""},
		{RecordEnvelope, 70, "DateSent"},
		{Record(99), 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
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
		{"Keywords", RecordApplication, 25, true},
		{"Byline", RecordApplication, 80, true},
		{"City", RecordApplication, 90, false},
		{"ObjectName", RecordApplication, 5, false},
		{"Envelope Destination", RecordEnvelope, 5, true},
		{"Unknown", RecordApplication, 255, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRepeatable(tt.record, tt.datasetID); got != tt.want {
				t.Errorf("isRepeatable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnvelopeDatasetsCoverage(t *testing.T) {
	// Ensure all envelope datasets are accessible
	expectedDatasets := []uint8{0, 5, 20, 22, 30, 40, 50, 60, 70, 80, 90, 100, 120, 122}
	for _, id := range expectedDatasets {
		info := getDatasetInfo(RecordEnvelope, id)
		if info.Name == "" {
			t.Errorf("Envelope dataset %d should have a name", id)
		}
	}
}

func TestApplicationDatasetsCoverage(t *testing.T) {
	// Ensure common application datasets are accessible
	commonDatasets := []uint8{0, 5, 10, 25, 55, 60, 80, 90, 95, 100, 101, 105, 110, 115, 116, 120}
	for _, id := range commonDatasets {
		info := getDatasetInfo(RecordApplication, id)
		if info.Name == "" {
			t.Errorf("Application dataset %d should have a name", id)
		}
	}
}
