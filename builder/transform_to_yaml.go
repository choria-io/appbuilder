// Copyright (c) 2023-2025, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/goccy/go-yaml"
)

type toYAMLTransform struct{}

func newToYAMLTransform(trans *Transform) (*toYAMLTransform, error) {
	yt := *trans.ToYAML
	return &yt, nil
}

func (t *toYAMLTransform) Validate(_ Logger) error {
	return nil
}

func (t *toYAMLTransform) Transform(ctx context.Context, r io.Reader, args map[string]any, flags map[string]any, b *AppBuilder) (io.Reader, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var input any

	if !isJSON(data) {
		return nil, fmt.Errorf("unsupported input format")
	}

	err = json.Unmarshal(data, &input)
	if err != nil {
		return nil, err
	}

	out, err := yaml.Marshal(input)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(out), err
}
