// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
)

// Transform is a generic transformation definition
type Transform struct {
	// Query is a JQ query to process, deprecated for backwards compatibility only
	Query string `json:"query"`

	// JQ is a JQ query to process
	JQ *jqTransform `json:"jq,omitempty"`

	// LineGraph is an ascii line graph from a single json array of float64 or float64 per line
	LineGraph *lineGraphTransform `json:"line_graph,omitempty"`

	// BarGraph is a ascii bar graph from a json map[string]float64
	BarGraph *barGraphTransform `json:"bar_graph,omitempty"`

	// Pipeline is a series of transforms to pass the data through
	Pipeline []Transform `json:"pipeline,omitempty"`

	// Template parses input through Go templates
	Template *templateTransform `json:"template,omitempty"`

	// Report turns row orientated data into a paged report
	Report *reportTransform `json:"report,omitempty"`

	// WriteFile writes data to a file
	WriteFile *writeFileTransform `json:"write_file,omitempty"`

	// ToJSON converts from YAML or JSON into JSON
	ToJSON *toJSONTransform `json:"to_json,omitempty"`

	// ToYAML converts from JSON to YAML
	ToYAML *toYAMLTransform `json:"to_yaml,omitempty"`

	// Scaffold renders complex multi file output from an input data structure
	Scaffold *scaffoldTransform `json:"scaffold,omitempty"`

	// CCMManifest executes a CCM manifest using the input data as manifest data
	CCMManifest *ccmManifestTransform `json:"ccm_manifest,omitempty"`
}

type transformer interface {
	Transform(ctx context.Context, r io.Reader, args map[string]any, flags map[string]any, b *AppBuilder) (io.Reader, error)
	Validate(Logger) error
}

var ErrInvalidTransform = errors.New("invalid transform")

func (t *Transform) transformerForQuery() (transformer, error) {
	switch {
	case t.Query != "" || t.JQ != nil:
		return newJQTransform(t)

	case t.LineGraph != nil:
		return newLineGraphTransform(t)

	case t.BarGraph != nil:
		return newBarGraphTransform(t)

	case t.Template != nil:
		return newTemplateTransform(t)

	case t.Report != nil:
		return newReportTransform(t)

	case t.WriteFile != nil:
		return newWriteFileTransform(t)

	case t.CCMManifest != nil:
		return newCCMManifestTransform(t)

	case len(t.Pipeline) > 0:
		return newPipelineTransform(t)

	case t.ToJSON != nil:
		return newToJSONTransform(t)

	case t.ToYAML != nil:
		return newToYAMLTransform(t)

	case t.Scaffold != nil:
		return newScaffoldTransform(t)

	default:
		return nil, fmt.Errorf("%w: no transform", ErrInvalidTransform)
	}
}

func (t *Transform) Validate(log Logger) error {
	trans, err := t.transformerForQuery()
	if err != nil {
		return err
	}

	return trans.Validate(log)
}

func (t *Transform) TransformBytes(ctx context.Context, r []byte, args map[string]any, flags map[string]any, b *AppBuilder) ([]byte, error) {
	res, err := t.Transform(ctx, bytes.NewReader(r), args, flags, b)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(res)
}

func (t *Transform) Transform(ctx context.Context, r io.Reader, args map[string]any, flags map[string]any, b *AppBuilder) (io.Reader, error) {
	trans, err := t.transformerForQuery()
	if err != nil {
		return nil, err
	}

	return trans.Transform(ctx, r, args, flags, b)

}
