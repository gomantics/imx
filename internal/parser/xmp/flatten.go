package xmp

import (
	"strings"

	"github.com/gomantics/imx/internal/parser"
)

func flattenNodeMap(nodeMap NodeMap, namespaces map[string]string) []parser.Directory {
	tagsByPrefix := make(map[string][]parser.Tag)

	for key, values := range nodeMap {
		prefix, ok := namespaces[key.URI]
		if !ok {
			if wellKnown, found := wellKnownPrefixes[key.URI]; found {
				prefix = wellKnown
			} else {
				prefix = defaultPrefix
			}
		}

		tagID := parser.TagID("XMP-" + prefix + ":" + key.Local)

		var finalVal any
		var dataType string

		if len(values) == 1 {
			finalVal, dataType = flattenVal(values[0])
		} else {
			list := make([]any, 0, len(values))
			for _, v := range values {
				val, _ := flattenVal(v)
				list = append(list, val)
			}
			finalVal = list
			dataType = "array"
		}

		tag := parser.Tag{
			ID:       tagID,
			Name:     key.Local,
			Value:    finalVal,
			DataType: dataType,
		}

		tagsByPrefix[prefix] = append(tagsByPrefix[prefix], tag)
	}

	var dirs []parser.Directory
	for prefix, tags := range tagsByPrefix {
		dir := parser.Directory{
			Name: "XMP-" + prefix,
			Tags: tags,
		}
		dirs = append(dirs, dir)
	}

	return dirs
}

func flattenVal(v PropertyValue) (any, string) {
	switch v.Kind {
	case KindSimple:
		return inferType(v.Scalar)
	case KindArray:
		list := make([]any, 0, len(v.Items))
		for _, item := range v.Items {
			val, _ := flattenVal(item)
			list = append(list, val)
		}
		return list, "array"
	case KindStruct:
		m := make(map[string]any, len(v.Fields))
		for _, f := range v.Fields {
			val, _ := flattenVal(f.Value)
			k := f.Prefix + ":" + f.Name
			m[k] = val
		}
		return m, "struct"
	}
	return nil, unknownDataType
}

func finalizeValue(ctx *ContextFrame) PropertyValue {
	if ctx.propKind == KindArray || len(ctx.items) > 0 {
		return PropertyValue{Kind: KindArray, Items: ctx.items}
	}
	if ctx.propKind == KindStruct || len(ctx.fields) > 0 {
		return PropertyValue{Kind: KindStruct, Fields: ctx.fields}
	}

	txt := strings.TrimSpace(ctx.text.String())
	return PropertyValue{Kind: KindSimple, Scalar: txt}
}
