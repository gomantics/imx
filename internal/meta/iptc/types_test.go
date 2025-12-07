package iptc

import "testing"

func TestRecord_String(t *testing.T) {
	tests := []struct {
		record Record
		want   string
	}{
		{RecordEnvelope, "Envelope"},
		{RecordApplication, "Application"},
		{RecordNewsPhoto, "NewsPhoto"},
		{RecordPreObjectData, "PreObjectData"},
		{RecordObjectData, "ObjectData"},
		{RecordPostObjectData, "PostObjectData"},
		{Record(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.record.String(); got != tt.want {
				t.Errorf("Record.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResourceConstants(t *testing.T) {
	// Verify resource ID constants
	if ResourceIPTC != 0x0404 {
		t.Errorf("ResourceIPTC = 0x%04X, want 0x0404", ResourceIPTC)
	}
	if ResourceXMP != 0x0424 {
		t.Errorf("ResourceXMP = 0x%04X, want 0x0424", ResourceXMP)
	}
	if ResourceICCProfile != 0x040F {
		t.Errorf("ResourceICCProfile = 0x%04X, want 0x040F", ResourceICCProfile)
	}
}
