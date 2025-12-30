package xmp

import "testing"

func TestContextType_String(t *testing.T) {
	tests := []struct {
		ct   ContextType
		want string
	}{
		{CTX_ROOT, "ROOT"},
		{CTX_RDF, "RDF"},
		{CTX_DESCRIPTION, "DESCRIPTION"},
		{CTX_PROPERTY, "PROPERTY"},
		{CTX_ARRAY, "ARRAY"},
		{CTX_LI, "LI"},
		{CTX_STRUCT_FIELD, "STRUCT_FIELD"},
		{ContextType(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.ct.String(); got != tt.want {
				t.Errorf("ContextType.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPropKind_String(t *testing.T) {
	tests := []struct {
		pk   PropKind
		want string
	}{
		{KindSimple, "Simple"},
		{KindArray, "Array"},
		{KindStruct, "Struct"},
		{KindUnknown, "Unknown"},
		{PropKind(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.pk.String(); got != tt.want {
				t.Errorf("PropKind.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
