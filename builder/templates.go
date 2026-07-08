// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"reflect"
	"text/template"

	"al.essio.dev/pkg/shellescape"

	"github.com/choria-io/appbuilder/internal/sprig"
)

func dereferenceArgsOrFlags(input map[string]any) map[string]any {
	res := map[string]any{}
	for k, v := range input {
		e := reflect.ValueOf(v).Elem()

		// the only kinds of values we support
		switch e.Kind() {
		case reflect.Bool:
			res[k] = e.Bool()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			res[k] = e.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			res[k] = e.Uint()
		case reflect.Float32, reflect.Float64:
			res[k] = e.Float()
		case reflect.String:
			res[k] = e.String()
		default:
			res[k] = fmt.Sprintf("%v", e)
		}
	}

	return res
}

func TemplateFuncs(all bool) template.FuncMap {
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

	funcs["env"] = func(v string) string {
		return os.Getenv(v)
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

// TemplateOption configures template rendering and state construction
type TemplateOption func(*templateOpts)

type templateOpts struct {
	sprig bool
	funcs template.FuncMap
	input any
}

func newTemplateOpts(opts ...TemplateOption) *templateOpts {
	o := &templateOpts{}
	for _, opt := range opts {
		opt(o)
	}

	return o
}

// WithSprig enables the sprig template function library, which is off by default
func WithSprig() TemplateOption {
	return func(o *templateOpts) {
		o.sprig = true
	}
}

// WithFuncs adds caller specific template functions on top of the standard set
func WithFuncs(funcs template.FuncMap) TemplateOption {
	return func(o *templateOpts) {
		if o.funcs == nil {
			o.funcs = template.FuncMap{}
		}
		for n, f := range funcs {
			o.funcs[n] = f
		}
	}
}

// WithInput sets the .Input value exposed to the template
func WithInput(input any) TemplateOption {
	return func(o *templateOpts) {
		o.input = input
	}
}

// NewTemplateState creates the state exposed to templates with Config and Secrets filled from the builder
func (b *AppBuilder) NewTemplateState(args map[string]any, flags map[string]any, opts ...TemplateOption) *TemplateState {
	o := newTemplateOpts(opts...)

	return &TemplateState{
		Arguments: dereferenceArgsOrFlags(args),
		Flags:     dereferenceArgsOrFlags(flags),
		Config:    b.cfg,
		Secrets:   b.secrets,
		Input:     o.input,
	}
}

// RenderTemplate parses and executes body as a Go text template. Config and Secrets are filled from
// the builder, the standard functions and the builder directory functions are always available and
// sprig functions are opt-in via WithSprig.
func (b *AppBuilder) RenderTemplate(body string, args map[string]any, flags map[string]any, opts ...TemplateOption) (string, error) {
	o := newTemplateOpts(opts...)

	funcs := b.TemplateFuncs(o.sprig)
	for n, f := range o.funcs {
		funcs[n] = f
	}

	temp, err := template.New("choria").Funcs(funcs).Parse(body)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = temp.Execute(&buf, b.NewTemplateState(args, flags, opts...))
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
