// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"fmt"
	"regexp"
	"strings"
)

type AppCheat struct {
	Enabled bool     `json:"enabled,omitempty"`
	Tags    []string `json:"tags,omitempty"`

	GenericCommandCheat
}

// Definition defines the entire application, it's the root of the app with all possible sub commands below it
type Definition struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Version     string    `json:"version"`
	Author      string    `json:"author"`
	Cheats      *AppCheat `json:"cheat"`

	GenericSubCommands

	commands []Command
}

const (
	semverVerifier = `^(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`
)

// Validate validates the top of the definition for validity
func (d *Definition) Validate(log Logger) error {
	errs := []string{}

	if d.Name == "" {
		errs = append(errs, "name is required")
	}
	if d.Description == "" {
		errs = append(errs, "description is required")
	}
	if matched, err := regexp.MatchString(semverVerifier, d.Version); !matched || err != nil {
		errs = append(errs, "version is required to be a valid semver")
	}
	if d.Author == "" {
		errs = append(errs, "author is required")
	}
	if len(d.Commands) == 0 {
		errs = append(errs, "no commands defined")
	}

	if len(errs) > 0 {
		return fmt.Errorf("%w: application: %s", ErrInvalidDefinition, strings.Join(errs, ", "))
	}

	return nil
}
