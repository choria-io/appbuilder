// Copyright (c) 2023, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package forms

import (
	"fmt"
)

// see comments in graph.go

func newObjectEntry(v map[string]any) entry {
	e := &objEntry{}
	e.set(v)

	return e
}

type objEntry struct {
	graph
	isSet     bool
	val       map[string]any
	arrayMode bool
}

func (s *objEntry) isEmptyValue() bool {
	return len(s.val) == 0
}

func (o *objEntry) addChild(e entry) (entry, error) {
	switch e.(type) {
	case *objEntry, *stringEntry:
		return o.graph.addChild(e)

	case *arrayEntry:
		if o.hasChildren() {
			return nil, fmt.Errorf("only one array child is supported")
		}

		o.arrayMode = true

		return o.graph.addChild(e)

	default:
		return e, fmt.Errorf("incompatible child type")
	}
}

func (o *objEntry) value() (nilValue bool, value any) {
	return o.isSet, o.val
}

func (o *objEntry) arrayModeCombined(tk string) (nilValue bool, value any) {
	var isSet bool
	res := map[string]any{}

	isSet, res[tk] = o.children[0].combinedValue()

	return isSet, res
}

func (o *objEntry) objModeCombined(tk string) (nilValue bool, value any) {
	result := map[string]any{}
	resultMap := map[string]any{}
	if tk == "" {
		result = resultMap
	} else {
		result[tk] = resultMap
	}

	cvlist := []map[string]any{}

	o.eachChild(func(e entry) {
		_, cval := e.combinedValue()
		cvlist = append(cvlist, cval.(map[string]any))
	})

	for _, e := range cvlist {
		for k, v := range e {
			resultMap[k] = v
		}
	}

	return true, result
}

func (o *objEntry) combinedValue() (nilValue bool, value any) {
	if !o.hasChildren() {
		return o.isSet, o.val
	}

	if !o.isSet {
		return false, o.val
	}

	tk := ""
	for k := range o.val {
		tk = k
		break
	}

	if o.arrayMode {
		return o.arrayModeCombined(tk)
	} else {
		return o.objModeCombined(tk)
	}
}

func (o *objEntry) set(v any) error {
	sv, ok := v.(map[string]any)
	if !ok {
		return fmt.Errorf("incompatible value")
	}

	o.val = sv
	o.isSet = true

	return nil
}
