// Copyright (c) 2023, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/ghodss/yaml"
)

type toJSONTransform struct {
	Prefix string `json:"prefix"`
	Indent string `json:"indent"`
}

func newToJSONTransform(trans *Transform) (*toJSONTransform, error) {
	jt := *trans.ToJSON
	return &jt, nil
}

func (t *toJSONTransform) Validate(_ Logger) error {
	return nil
}

func (t *toJSONTransform) Transform(ctx context.Context, r io.Reader, args map[string]any, flags map[string]any, b *AppBuilder) (io.Reader, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var input any
	var out []byte

	if isJSON(data) {
		err = json.Unmarshal(data, &input)
	} else {
		err = yaml.Unmarshal(data, &input)
	}
	if err != nil {
		return nil, err
	}

	if t.Prefix == "" && t.Indent == "" {
		out, err = json.Marshal(input)
	} else {
		out, err = json.MarshalIndent(input, t.Prefix, t.Indent)
	}
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(out), nil
}
