package xmp

import (
	"fmt"

	"github.com/gomantics/imx/internal/common"
)

// flattenNodeMap converts the nested NodeMap into a flat Directory structure
// suitable for the public API, while preserving hierarchical data in nested maps/slices.

func flattenNodeMap(nodeMap NodeMap, namespaces map[string]string) common.Directory {
	dir := common.Directory{
		Spec: common.SpecXMP,
		Name: directoryName,
		Tags: make(map[common.TagID]common.Tag),
	}

	for key, values := range nodeMap {
		// Resolve prefix: first from runtime namespaces, then well-known, finally fallback
		prefix, ok := namespaces[key.URI]
		if !ok {
			if wellKnown, found := wellKnownPrefixes[key.URI]; found {
				prefix = wellKnown
			} else {
				prefix = defaultPrefix // Fallback for unknown namespaces
			}
		}

		tagID := common.TagID(fmt.Sprintf(tagIDFormat, prefix, key.Local))

		var finalVal any
		var dataType string

		if len(values) == 1 {
			finalVal, dataType = flattenVal(values[0])
		} else {
			var list []any
			for _, v := range values {
				val, _ := flattenVal(v)
				list = append(list, val)
			}
			finalVal = list
			dataType = "array"
		}

		dir.Tags[tagID] = common.Tag{
			Spec:     common.SpecXMP,
			ID:       tagID,
			Name:     key.Local,
			DataType: dataType,
			Value:    finalVal,
		}
	}
	return dir
}

// flattenVal recursively converts a PropertyValue into a Go value suitable for the public API.
// Simple values are type-inferred (bool, int, float, or string).
// Array values become []any with recursive flattening of items.
// Struct values become map[string]any with field keys as "prefix:name".
// Returns the flattened value and a string describing its data type.
func flattenVal(v PropertyValue) (any, string) {
	switch v.Kind {
	case KindSimple:
		return inferType(v.Scalar)
	case KindArray:
		var list []any
		for _, item := range v.Items {
			val, _ := flattenVal(item)
			list = append(list, val)
		}
		return list, "array"
	case KindStruct:
		m := make(map[string]any)
		for _, f := range v.Fields {
			val, _ := flattenVal(f.Value)
			// Struct field key: prefix:name
			k := fmt.Sprintf("%s:%s", f.Prefix, f.Name)
			m[k] = val
		}
		return m, "struct"
	}
	return nil, unknownDataType
}
