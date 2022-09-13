// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"bytes"
	"errors"
	"os"
	"reflect"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"gopkg.in/alessio/shellescape.v1"
)

func dereferenceArgsOrFlags(input map[string]any) map[string]any {
	res := map[string]any{}
	for k, v := range input {
		e := reflect.ValueOf(v).Elem()

		// the only kinds of values we support
		if e.Kind() == reflect.Bool {
			res[k] = e.Bool()
		} else {
			res[k] = e.String()
		}
	}

	return res
}

func templateFuncs(all bool) template.FuncMap {
	funcs := map[string]any{}
	if all {
		funcs = sprig.TxtFuncMap()
	}

	funcs["require"] = func(v any, reason string) (any, error) {
		err := errors.New("value required")
		if reason != "" {
			err = errors.New(reason)
		}

		switch val := v.(type) {
		case string:
			if val == "" {
				return "", err
			}
		default:
			if v == nil {
				return "", err
			}
		}

		return v, nil
	}

	funcs["escape"] = func(v string) string {
		return shellescape.Quote(v)
	}

	funcs["read_file"] = func(v string) (string, error) {
		b, err := os.ReadFile(v)
		if err != nil {
			return "", err
		}

		return string(b), nil
	}

	funcs["default"] = func(v any, dflt string) string {
		switch c := v.(type) {
		case string:
			if c != "" {
				return c
			}
		}

		return dflt
	}

	return funcs
}

// ParseStateTemplate parses body as a go text template with supplied values exposed to the user
func ParseStateTemplate(body string, args map[string]any, flags map[string]any, cfg any) (string, error) {
	state := templateState{
		Arguments: dereferenceArgsOrFlags(args),
		Flags:     dereferenceArgsOrFlags(flags),
		Config:    cfg,
	}

	temp, err := template.New("choria").Funcs(templateFuncs(false)).Parse(body)
	if err != nil {
		return "", err
	}

	var b bytes.Buffer
	err = temp.Execute(&b, state)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}
