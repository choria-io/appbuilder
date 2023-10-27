// Copyright (c) 2023, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package forms

import (
	"fmt"
)

// see comments in graph.go

func newStringEntry(v string) entry {
	e := &stringEntry{}
	e.set(v)

	return e
}

type stringEntry struct {
	graph
	isSet bool
	val   string
}

func (s *stringEntry) addChild(e entry) (entry, error) {
	if _, ok := e.(*objEntry); !ok {
		return nil, fmt.Errorf("incompatible type, only object child values are supported")
	}

	return s.graph.addChild(e)
}

func (s *stringEntry) set(v any) error {
	sv, ok := v.(string)
	if !ok {
		return fmt.Errorf("incompatible value")
	}

	s.val = sv
	s.isSet = true

	return nil
}

func (s *stringEntry) value() (nilValue bool, value any) {
	return s.isSet, s.val
}

func (s *stringEntry) combinedValue() (nilValue bool, value any) {
	if !s.isSet {
		return false, nil
	}

	if len(s.children) == 1 {
		_, cv := s.children[0].combinedValue()
		return true, map[string]any{s.val: cv}
	}

	resMap := map[string]any{}
	res := map[string]any{
		s.val: resMap,
	}

	s.eachChild(func(e entry) {
		if isSet, val := e.combinedValue(); isSet {
			mv := val.(map[string]any)
			for k := range mv {
				resMap[k] = mv[k]
			}
		}
	})

	return true, res
}

func (s *stringEntry) isEmptyValue() bool {
	return s.val == ""
}
