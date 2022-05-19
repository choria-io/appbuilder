// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"encoding/json"
	"fmt"
	"sync"
)

type plugin struct {
	cons CommandConstructor
}

var (
	commandPlugins map[string]*plugin
	pmu            sync.Mutex
)

// CommandConstructor should exist in any package that is used as a plugin
type CommandConstructor func(*CLIBuilder, json.RawMessage, Logger) (Command, error)

// RegisterCommand adds a new kind of command
func RegisterCommand(kind string, constructor CommandConstructor) error {
	pmu.Lock()
	defer pmu.Unlock()

	if commandPlugins == nil {
		commandPlugins = make(map[string]*plugin)
	}

	_, ok := commandPlugins[kind]
	if ok {
		return fmt.Errorf("%s: %s", ErrorDuplicatePlugin, kind)
	}

	commandPlugins[kind] = &plugin{constructor}

	return nil
}

// MustRegisterCommand registers a command and panics if it cannot
func MustRegisterCommand(kind string, constructor CommandConstructor) {
	err := RegisterCommand(kind, constructor)
	if err != nil {
		panic(err)
	}
}

func commandByKind(kind string) (CommandConstructor, error) {
	pmu.Lock()
	defer pmu.Unlock()

	cmd, ok := commandPlugins[kind]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrorUnknownPlugin, kind)
	}

	return cmd.cons, nil
}
