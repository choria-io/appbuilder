// Copyright (c) 2023, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package forms

import (
	"os"

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
