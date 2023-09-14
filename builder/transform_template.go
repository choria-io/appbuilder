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
	"text/template"
)

type templateTransform struct {
	Contents string `json:"contents"`
	Source   string `json:"source"`
}

func newTemplateTransform(trans *Transform) (*templateTransform, error) {
	// copy it
	tmpl := *trans.Template
	return &tmpl, nil
}

func (tt *templateTransform) Validate(_ Logger) error {
	if tt.Source == "" && tt.Contents == "" {
		return fmt.Errorf("contents or source is required")
	}

	if tt.Source != "" && tt.Contents != "" {
		return fmt.Errorf("contents and source cannot be used together")
	}

	return nil
}

func (tt *templateTransform) Transform(ctx context.Context, r io.Reader, args map[string]any, flags map[string]any, b *AppBuilder) (io.Reader, error) {
	var input any

	j, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(j, &input)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON input for template: %v", err)
	}

	templ := template.New("builder")
	templ.Funcs(TemplateFuncs(true))

	switch {
	case tt.Source != "":
		var source string
		source, err = ParseStateTemplate(tt.Source, args, flags, b.Configuration())
		if err != nil {
			return nil, fmt.Errorf("invalid source template: %v", err)
		}

		_, err = templ.ParseFiles(source)
	case tt.Contents != "":
		_, err = templ.Parse(tt.Contents)
	}
	if err != nil {
		return nil, fmt.Errorf("could not parse template: %v", err)
	}

	out := bytes.NewBuffer([]byte{})
	state := NewTemplateState(args, flags, b.Configuration(), input)

	err = templ.Execute(out, state)
	if err != nil {
		return nil, fmt.Errorf("could not execute template: %v", err)
	}

	return out, nil
}
