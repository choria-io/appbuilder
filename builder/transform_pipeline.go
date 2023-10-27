// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"context"
	"io"
)

type pipelineTransform struct {
	pipe *Transform
}

func newPipelineTransform(trans *Transform) (*pipelineTransform, error) {
	return &pipelineTransform{pipe: trans}, nil
}

func (pt *pipelineTransform) Validate(log Logger) error {
	for _, t := range pt.pipe.Pipeline {
		transformer, err := t.transformerForQuery()
		if err != nil {
			return err
		}

		err = transformer.Validate(log)
		if err != nil {
			return err
		}
	}

	return nil
}

func (pt *pipelineTransform) Transform(ctx context.Context, r io.Reader, args map[string]any, flags map[string]any, b *AppBuilder) (io.Reader, error) {
	out := r

	for _, t := range pt.pipe.Pipeline {
		trans, err := t.transformerForQuery()
		if err != nil {
			return nil, err
		}

		out, err = trans.Transform(ctx, out, args, flags, b)
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}
