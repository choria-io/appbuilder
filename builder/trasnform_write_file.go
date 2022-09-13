// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"text/template"

	"github.com/dustin/go-humanize"
)

type writeFileTransform struct {
	File    string `json:"file"`
	Message string `json:"message,omitempty"`
	Replace bool   `json:"replace,omitempty"`
}

func newWriteFileTransform(trans *Transform) (*writeFileTransform, error) {
	// copy it
	wf := *trans.WriteFile
	return &wf, nil
}

func (wf *writeFileTransform) Validate(_ Logger) error {
	if wf.File == "" {
		return fmt.Errorf("a file name is required")
	}

	return nil
}

func (wf *writeFileTransform) Transform(ctx context.Context, r io.Reader, args map[string]any, flags map[string]any, cfg any) (io.Reader, error) {
	input, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	target, err := ParseStateTemplate(wf.File, args, flags, cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid file name template: %v", err)
	}

	_, err = os.Stat(target)
	if !os.IsNotExist(err) && !wf.Replace {
		return nil, fmt.Errorf("%s already exist", target)
	}

	err = os.WriteFile(target, input, 0644)
	if err != nil {
		return nil, err
	}

	if wf.Message != "" {
		t, err := template.New("write_file").Parse(wf.Message + "\n")
		if err != nil {
			return nil, fmt.Errorf("invalid message template: %v", err)
		}
		out := bytes.NewBuffer([]byte{})
		t.Execute(out, map[string]any{
			"Target":   target,
			"Contents": input,
			"Bytes":    len(input),
			"IBytes":   humanize.IBytes(uint64(len(input))),
		})

		return out, nil
	}

	return bytes.NewReader(input), nil
}
