package xmp

import (
	"fmt"

	"github.com/gomantics/imx/internal/meta"
)

// flattenNodeMap converts the nested NodeMap into a flat Directory structure
// suitable for the public API, while preserving hierarchical data in nested maps/slices.

func flattenNodeMap(nodeMap NodeMap, namespaces map[string]string) meta.Directory {
	dir := meta.Directory{
		Spec: meta.SpecXMP,
		Name: "XMP",
		Tags: make(map[meta.TagID]meta.Tag),
	}

	for key, values := range nodeMap {
		prefix := namespaces[key.URI]
		if prefix == "" {
			// Fallback if not captured? should be captured.
			prefix = wellKnownPrefixes[key.URI]
			if prefix == "" {
				prefix = "ns"
			}
		}

		tagID := meta.TagID(fmt.Sprintf("XMP-%s:%s", prefix, key.Local))

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

		dir.Tags[tagID] = meta.Tag{
			Spec:     meta.SpecXMP,
			ID:       tagID,
			Name:     key.Local,
			DataType: dataType,
			Value:    finalVal,
		}
	}
	return dir
}

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
	return nil, "unknown"
}
