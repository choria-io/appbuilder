// Copyright (c) 2023, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/choria-io/scaffold"
)

type scaffoldTransform struct {
	// TargetDirectory is where to place the resulting rendered files, must not exist
	TargetDirectory string `json:"target"`
	// SourceDirectory reads templates from a directory, mutually exclusive with Source
	SourceDirectory string `json:"source_directory"`
	// Source reads templates from in-process memory
	Source map[string]any `json:"source"`
	// Post configures post-processing of files using filepath globs
	Post []map[string]string `json:"post"`
	// SkipEmpty skips files that are 0 bytes after rendering
	SkipEmpty bool `json:"skip_empty"`
	// Sets a custom template delimiter, useful for generating templates from templates
	CustomLeftDelimiter string `json:"left_delimiter"`
	// Sets a custom template delimiter, useful for generating templates from templates
	CustomRightDelimiter string `json:"right_delimiter"`
}

func newScaffoldTransform(trans *Transform) (*scaffoldTransform, error) {
	sc := *trans.Scaffold

	return &sc, nil
}

func (st *scaffoldTransform) Validate(_ Logger) error {
	if st.TargetDirectory == "" {
		return fmt.Errorf("target location is required")
	}

	return nil
}

func (st *scaffoldTransform) Transform(ctx context.Context, r io.Reader, args map[string]any, flags map[string]any, b *AppBuilder) (io.Reader, error) {
	j, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var input map[string]any
	json.Unmarshal(j, &input)

	st.SourceDirectory, err = ParseStateTemplateWithFuncMap(st.SourceDirectory, args, flags, b.Configuration(), b.TemplateFuncs(true))
	if err != nil {
		return nil, err
	}

	st.TargetDirectory, err = ParseStateTemplateWithFuncMap(st.TargetDirectory, args, flags, b.Configuration(), b.TemplateFuncs(true))
	if err != nil {
		return nil, err
	}

	cfg := scaffold.Config{
		TargetDirectory:      st.TargetDirectory,
		SourceDirectory:      st.SourceDirectory,
		Source:               st.Source,
		Post:                 st.Post,
		SkipEmpty:            st.SkipEmpty,
		CustomLeftDelimiter:  st.CustomLeftDelimiter,
		CustomRightDelimiter: st.CustomRightDelimiter,
	}

	s, err := scaffold.New(cfg, b.TemplateFuncs(true))
	if err != nil {
		return nil, err
	}

	s.Logger(b.log)

	err = s.Render(NewTemplateState(args, flags, b.Configuration(), input))
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(j), nil
}
