// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/choria-io/goform"
)

type reportTransform struct {
	Name         string `json:"name,omitempty"`
	BodyLayout   string `json:"body,omitempty"`
	HeaderLayout string `json:"header,omitempty"`
	FooterLayout string `json:"footer,omitempty"`
	RowsPerPage  int    `json:"rows_per_page,omitempty"`
	InitialQuery string `json:"initial_query,omitempty"`
	SourceFile   string `json:"source_file,omitempty"`
}

func newReportTransform(trans *Transform) (*reportTransform, error) {
	report := *trans.Report
	return &report, nil
}

func (rt *reportTransform) Validate(_ Logger) error {
	if rt.BodyLayout == "" && rt.SourceFile == "" {
		return fmt.Errorf("body layout or source file is required")
	}

	return nil
}

func (rt *reportTransform) Transform(ctx context.Context, r io.Reader, args map[string]any, flags map[string]any, cfg any) (io.Reader, error) {
	j, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	name, err := ParseStateTemplate(rt.Name, args, flags, cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid report name template: %v", err)
	}

	var rprt *goform.Report
	if rt.SourceFile != "" {
		var source string
		source, err = ParseStateTemplate(rt.SourceFile, args, flags, cfg)
		if err != nil {
			return nil, fmt.Errorf("invalid report source template: %v", err)
		}
		rprt, err = goform.NewFromFile(source, name)
	} else {
		rprt, err = goform.New(name, rt.HeaderLayout, rt.BodyLayout, rt.FooterLayout, rt.RowsPerPage)
	}
	if err != nil {
		return nil, fmt.Errorf("could not initialize report: %v", err)
	}

	out := bytes.NewBuffer([]byte{})
	if rt.InitialQuery != "" {
		var input any
		err = json.Unmarshal(j, &input)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON input for report: %v", err)
		}

		err = rprt.WriteReportContainedRows(out, input, rt.InitialQuery)
		if err != nil {
			return nil, fmt.Errorf("could not generated nested data report: %v", err)
		}
	} else {
		var input []any
		err = json.Unmarshal(j, &input)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON input for report: %v", err)
		}

		err = rprt.WriteReport(out, input)
		if err != nil {
			return nil, fmt.Errorf("could not generated report: %v", err)
		}
	}

	return out, nil
}
