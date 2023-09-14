// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/guptarohit/asciigraph"
)

type lineGraphTransform struct {
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Caption   string `json:"caption"`
	Precision uint   `json:"precision"`
	JSONInput bool   `json:"json"`
}

func newLineGraphTransform(trans *Transform) (*lineGraphTransform, error) {
	// copy it
	lg := *trans.LineGraph
	return &lg, nil
}

func (lg *lineGraphTransform) Validate(_ Logger) error { return nil }

func (t *lineGraphTransform) Transform(ctx context.Context, r io.Reader, args map[string]any, flags map[string]any, b *AppBuilder) (io.Reader, error) {
	out := bytes.NewBuffer([]byte{})
	var input []float64

	if t.JSONInput {
		j, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(j, &input)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON input for line graph: %v", err)
		}
	} else {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()
			f, err := strconv.ParseFloat(line, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid number in input: %v", err)
			}
			input = append(input, f)
		}
	}

	var opts []asciigraph.Option

	if t.Height > 0 {
		opts = append(opts, asciigraph.Height(t.Height))
	} else {
		opts = append(opts, asciigraph.Height(20))
	}
	if t.Width > 0 {
		opts = append(opts, asciigraph.Width(t.Width))
	} else {
		opts = append(opts, asciigraph.Width(40))
	}
	if t.Precision > 0 {
		opts = append(opts, asciigraph.Precision(t.Precision))
	}
	if t.Caption != "" {
		caption, err := ParseStateTemplate(t.Caption, args, flags, b.Configuration())
		if err != nil {
			return nil, fmt.Errorf("invalid line graph caption: %v", err)
		}
		opts = append(opts, asciigraph.Caption(caption))
	}

	fmt.Fprintln(out, asciigraph.Plot(input, opts...))

	return out, nil
}
