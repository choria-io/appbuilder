// Copyright (c) 2023, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package forms

import (
	"fmt"
)

// see comments in graph.go

func newArrayEntry(v []any) entry {
	e := &arrayEntry{}
	e.set(v)

	return e
}

type arrayEntry struct {
	graph
	isSet bool
	val   []any
}

func (s *arrayEntry) isEmptyValue() bool {
	return len(s.val) == 0
}

func (a *arrayEntry) value() (nilValue bool, value any) {
	return a.isSet, a.val
}

func (a *arrayEntry) combinedValue() (nilValue bool, value any) {
	res := make([]any, len(a.val))
	copy(res, a.val)

	a.eachChild(func(e entry) {
		_, val := e.combinedValue()
		res = append(res, val)
	})

	return len(res) > 0, res
}

func (a *arrayEntry) set(v any) error {
	sv, ok := v.([]any)
	if !ok {
		return fmt.Errorf("incompatible value")
	}

	a.val = sv
	a.isSet = true

	return nil
}
