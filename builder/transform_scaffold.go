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

	"github.com/choria-io/appbuilder/scaffold"
)

type scaffoldTransform struct {
	Target               string              `json:"target"`
	Post                 []map[string]string `json:"post"`
	SkipEmpty            bool                `json:"skip_empty"`
	CustomLeftDelimiter  string              `json:"left_delimiter"`
	CustomRightDelimiter string              `json:"right_delimiter"`
}

func newScaffoldTransform(trans *Transform) (*scaffoldTransform, error) {
	sc := *trans.Scaffold

	return &sc, nil
}

func (st *scaffoldTransform) Validate(_ Logger) error {
	if st.Target == "" {
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

	scfg := scaffold.Config{
		TargetDirectory:      st.Target,
		Source:               input,
		Post:                 st.Post,
		SkipEmpty:            st.SkipEmpty,
		CustomLeftDelimiter:  st.CustomLeftDelimiter,
		CustomRightDelimiter: st.CustomLeftDelimiter,
	}

	scfg.TargetDirectory, err = ParseStateTemplateWithFuncMap(st.Target, args, flags, b.Configuration(), b.TemplateFuncs(true))
	if err != nil {
		return nil, err
	}

	s, err := scaffold.New(scfg, b.TemplateFuncs(true))
	if err != nil {
		return nil, err
	}

	s.Logger(b.log)

	err = s.Render(NewTemplateState(args, flags, b.Configuration(), nil))
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(j), nil
}
