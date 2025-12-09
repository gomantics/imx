package iptc

import (
	"fmt"

	"github.com/gomantics/imx/internal/common"
)

// Parser implements meta.Parser for IPTC-IIM metadata
type Parser struct{}

// New creates a new IPTC parser
func New() *Parser {
	return &Parser{}
}

// Spec returns the metadata spec this parser handles
func (p *Parser) Spec() common.Spec {
	return common.SpecIPTC
}

// Parse extracts IPTC metadata from raw blocks
func (p *Parser) Parse(blocks []common.RawBlock) ([]common.Directory, error) {
	if len(blocks) == 0 {
		return nil, nil
	}

	var allDatasets []Dataset

	for _, block := range blocks {
		if block.Spec != common.SpecIPTC {
			continue
		}

		// Parse Photoshop IRB to extract IPTC data
		iptcData, err := parsePhotoshopIRB(block.Payload)
		if err != nil {
			continue
		}

		if len(iptcData) == 0 {
			// Try parsing as raw IPTC-IIM (some formats embed it directly)
			iptcData = block.Payload
		}

		// Parse IPTC-IIM data
		datasets, _ := parseIPTCIIM(iptcData)
		allDatasets = append(allDatasets, datasets...)
	}

	if len(allDatasets) == 0 {
		return nil, nil
	}

	// Build directories from datasets
	dirs := p.buildDirectories(allDatasets)
	return dirs, nil
}

// buildDirectories creates common.Directory structures from parsed datasets
func (p *Parser) buildDirectories(datasets []Dataset) []common.Directory {
	// Group datasets by record
	byRecord := make(map[Record][]Dataset)
	for _, ds := range datasets {
		byRecord[ds.Record] = append(byRecord[ds.Record], ds)
	}

	var dirs []common.Directory

	// Process each record
	for record, recordDatasets := range byRecord {
		dir := common.Directory{
			Spec: common.SpecIPTC,
			Name: fmt.Sprintf("IPTC-%s", record.String()),
			Tags: make(map[common.TagID]common.Tag),
		}

		// Track values for repeatable fields
		repeatableValues := make(map[string][]any)
		repeatableRaws := make(map[string][][]byte)

		for _, ds := range recordDatasets {
			tagID := common.TagID("IPTC:" + ds.Name)

			if isRepeatable(ds.Record, ds.DatasetID) {
				// Aggregate repeatable field values into arrays
				repeatableValues[ds.Name] = append(repeatableValues[ds.Name], ds.Value)
				repeatableRaws[ds.Name] = append(repeatableRaws[ds.Name], ds.Raw)
			} else {
				// Non-repeatable field - create tag directly
				dataType := "string"
				switch ds.Value.(type) {
				case int:
					dataType = "int"
				}

				dir.Tags[tagID] = common.Tag{
					Spec:     common.SpecIPTC,
					ID:       tagID,
					Name:     ds.Name,
					DataType: dataType,
					Value:    ds.Value,
					Raw:      ds.Raw,
				}
			}
		}

		// Create tags for repeatable fields with aggregated values
		for name, values := range repeatableValues {
			tagID := common.TagID("IPTC:" + name)
			var value any
			if len(values) == 1 {
				value = values[0]
			} else {
				value = values
			}

			dir.Tags[tagID] = common.Tag{
				Spec:     common.SpecIPTC,
				ID:       tagID,
				Name:     name,
				DataType: "array",
				Value:    value,
				Raw:      repeatableRaws[name][0], // Use first raw value
			}
		}

		if len(dir.Tags) > 0 {
			dirs = append(dirs, dir)
		}
	}

	return dirs
}
