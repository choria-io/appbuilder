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
	"math"
	"sort"
	"strings"

	"github.com/dustin/go-humanize"
)

type barGraphTransform struct {
	Width   int    `json:"width"`
	Caption string `json:"caption"`
	Bytes   bool   `json:"bytes"`
}

func newBarGraphTransform(trans *Transform) (*barGraphTransform, error) {
	// copy it
	bg := *trans.BarGraph

	if bg.Width == 0 {
		bg.Width = 40
	}

	return &bg, nil
}

func (bg *barGraphTransform) Validate(_ Logger) error { return nil }

func (bg *barGraphTransform) Transform(ctx context.Context, r io.Reader, args map[string]any, flags map[string]any, cfg any) (io.Reader, error) {
	out := bytes.NewBuffer([]byte{})
	var input map[string]float64
	var caption string
	var err error

	if bg.Caption != "" {
		caption, err = ParseStateTemplate(bg.Caption, args, flags, cfg)
		if err != nil {
			return nil, fmt.Errorf("invalid bar graph caption: %v", err)
		}
	}

	j, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(j, &input)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON input for bar graph: %v", err)
	}

	longest := 0
	min := math.MaxFloat64
	max := -math.MaxFloat64
	keys := []string{}
	for k, v := range input {
		keys = append(keys, k)
		if len(k) > longest {
			longest = len(k)
		}

		if v < min {
			min = v
		}

		if v > max {
			max = v
		}
	}

	sort.Slice(keys, func(i, j int) bool {
		return input[keys[i]] < input[keys[j]]
	})

	if caption != "" {
		fmt.Fprintln(out, caption)
		fmt.Fprintln(out)
	}

	var steps float64
	if max == min {
		steps = max / float64(bg.Width)
	} else {
		steps = (max - min) / float64(bg.Width)
	}

	longestLine := 0
	for _, k := range keys {
		v := input[k]

		var blocks int
		if v-min == 0 {
			blocks = bg.Width
		} else {
			blocks = int((v - min) / steps)
		}

		var h string
		if bg.Bytes {
			h = humanize.IBytes(uint64(v))
		} else {
			h = humanize.Commaf(v)
		}

		bar := strings.Repeat("█", blocks)
		if blocks == 0 {
			bar = "▏"
		}

		line := fmt.Sprintf("%s%s: %s (%s)", strings.Repeat(" ", longest-len(k)+2), k, bar, h)
		if len(line) > longestLine {
			longestLine = len(line)
		}

		fmt.Fprintln(out, line)
	}

	return out, nil
}
