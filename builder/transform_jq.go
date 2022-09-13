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

	"github.com/itchyny/gojq"
)

type jqTransform struct {
	Query string `json:"query"`

	def *Transform
}

func newJQTransform(trans *Transform) (*jqTransform, error) {
	t := &jqTransform{
		def: trans,
	}

	// keeps backwards compat
	if t.def.JQ != nil {
		t.Query = t.def.JQ.Query
	}

	if t.Query == "" {
		t.Query = t.def.Query
	}

	return t, nil
}

func (t *jqTransform) Validate(_ Logger) error {
	if t.Query == "" {
		return fmt.Errorf("%w: no JQ query defined", ErrInvalidTransform)
	}

	return nil
}

func (t *jqTransform) Transform(ctx context.Context, r io.Reader, args map[string]any, flags map[string]any, cfg any) (io.Reader, error) {
	j, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	out := bytes.NewBuffer([]byte{})
	var input any

	err = json.Unmarshal(j, &input)
	if err != nil {
		return nil, fmt.Errorf("json output parse error: %v", err)
	}

	query, err := ParseStateTemplate(t.Query, args, flags, cfg)
	if err != nil {
		return nil, err
	}

	q, err := gojq.Parse(query)
	if err != nil {
		return nil, err
	}

	iter := q.RunWithContext(ctx, input)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}

		switch val := v.(type) {
		case error:
			return nil, val
		case string:
			fmt.Fprintln(out, val)
		default:
			j, err := json.MarshalIndent(val, "", "  ")
			if err != nil {
				return nil, err
			}
			fmt.Fprintln(out, string(j))
		}
	}

	return out, nil
}
