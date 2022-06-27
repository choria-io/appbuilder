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

	"gopkg.in/alessio/shellescape.v1"
)

func dereferenceArgsOrFlags(input map[string]interface{}) map[string]interface{} {
	res := map[string]interface{}{}
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

// ParseStateTemplate parses body as a go text template with supplied values exposed to the user
func ParseStateTemplate(body string, args map[string]interface{}, flags map[string]interface{}, cfg interface{}) (string, error) {
	state := templateState{
		Arguments: dereferenceArgsOrFlags(args),
		Flags:     dereferenceArgsOrFlags(flags),
		Config:    cfg,
	}

	funcs := map[string]interface{}{
		"require": func(v interface{}, reason string) (interface{}, error) {
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
		},

		"escape": func(v string) string {
			return shellescape.Quote(v)
		},

		"read_file": func(v string) (string, error) {
			b, err := os.ReadFile(v)
			if err != nil {
				return "", err
			}

			return string(b), nil
		},

		"default": func(v interface{}, dflt string) string {
			switch c := v.(type) {
			case string:
				if c != "" {
					return c
				}
			}

			return dflt
		},
	}

	temp, err := template.New("choria").Funcs(funcs).Parse(body)
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
