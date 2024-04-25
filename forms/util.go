// Copyright (c) 2023, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package forms

import (
	"bytes"
	"github.com/choria-io/appbuilder/internal/sprig"
	"os"
	"text/template"

	"github.com/AlecAivazis/survey/v2"
	terminal "golang.org/x/term"
)

func propertyEmptyVal(p Property) any {
	switch p.IfEmpty {
	case ArrayIfEmpty:
		return map[string]any{p.Name: []any{}}
	case ObjectIfEmpty:
		return map[string]any{p.Name: map[string]any{}}
	default:
		return map[string]any{}
	}
}
func askConfirmation(prompt string, dflt bool) (bool, error) {
	ans := dflt

	err := survey.AskOne(&survey.Confirm{
		Message: prompt,
		Default: dflt,
	}, &ans)

	return ans, err
}

func isTerminal() bool {
	return terminal.IsTerminal(int(os.Stdin.Fd())) && terminal.IsTerminal(int(os.Stdout.Fd()))
}

func isOneOf(val string, valid ...string) bool {
	for _, v := range valid {
		if val == v {
			return true
		}
	}
	return false
}

func renderTemplate(tmpl string, env map[string]any) (string, error) {
	t, err := template.New("form").Funcs(sprig.TxtFuncMap()).Parse(tmpl)
	if err != nil {
		return "", err
	}

	out := bytes.NewBuffer([]byte{})

	err = t.Execute(out, env)
	if err != nil {
		return "", err
	}

	return out.String(), nil
}
