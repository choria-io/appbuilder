// Copyright (c) 2023, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package forms

import (
	"fmt"
)

// This is a data structure that can be built using recursive techniques but that at any time can be interrogated
// for the value built so far.  Generally recursively built data structures are only assembled fully at the time that
// the last recursion finishes.
//
// For forms though we want to be able to say only ask about an item if a previous item meets some condition, this means
// we both need to build recursively but at any time be able to look back.
//
// This code, really  awful code, allows that by adding objects, strings and arrays into a dag that can be walked at
// any time to build the result accumulated thus far.  Even while building a map or array one can reference back and
// would include the currently-being-built map as it stands up to the point.
//
// This is an entirely internal detail, so while I am not happy with the code as it is its been a struggle to get working
// at all, so, I am shipping it as is for now.

type entry interface {
	addChild(entry) (entry, error)
	setParent(entry) error
	value() (nilValue bool, value any)
	combinedValue() (nilValue bool, value any)
	set(any) error
	isEmptyValue() bool
}

type graph struct {
	children []entry
	parent   entry
}

func (g *graph) addChild(e entry) (entry, error) {
	err := e.setParent(e)
	if err != nil {
		return nil, err
	}

	g.children = append(g.children, e)

	return e, nil
}

func (g *graph) setParent(e entry) error {
	if g.parent != nil {
		return fmt.Errorf("parent already set")
	}

	g.parent = e

	return nil
}

func (g *graph) hasChildren() bool {
	return len(g.children) > 0
}

func (g *graph) eachChild(cb func(entry)) {
	for i := 0; i < len(g.children); i++ {
		cb(g.children[i])
	}
}
