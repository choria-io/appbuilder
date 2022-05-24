// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package parent

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/choria-io/appbuilder/builder"
	"gopkg.in/alecthomas/kingpin.v2"
)

type Command struct {
	builder.GenericSubCommands
	builder.GenericCommand
}

type Parent struct {
	cmd *kingpin.CmdClause
	def *Command
}

func MustRegister() {
	builder.MustRegisterCommand("parent", NewParentCommand)
}

func NewParentCommand(_ *builder.AppBuilder, j json.RawMessage, _ builder.Logger) (builder.Command, error) {
	parent := &Parent{
		def: &Command{},
	}

	err := json.Unmarshal(j, parent.def)
	if err != nil {
		return nil, err
	}

	return parent, nil
}

func (r *Parent) String() string { return fmt.Sprintf("%s (parent)", r.def.Name) }

func (r *Parent) Validate(log builder.Logger) error {
	if r.def.Type != "parent" {
		return fmt.Errorf("not a parent command")
	}

	var errs []string

	err := r.def.GenericCommand.Validate(log)
	if err != nil {
		errs = append(errs, err.Error())
	}

	err = r.def.GenericSubCommands.Validate(log)
	if err != nil {
		errs = append(errs, err.Error())
	}

	if len(r.def.Flags) > 0 {
		errs = append(errs, "parent commands can not have flags")
	}

	if len(r.def.Arguments) > 0 {
		errs = append(errs, "parent commands can not have arguments")
	}

	if len(r.def.Commands) == 0 {
		errs = append(errs, "parent requires sub commands")
	}

	if r.def.ConfirmPrompt != "" {
		errs = append(errs, "parents do not accept confirm prompts")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, ", "))
	}

	return nil
}

func (p *Parent) SubCommands() []json.RawMessage {
	return p.def.Commands
}

func (p *Parent) CreateCommand(app builder.KingpinCommand) (*kingpin.CmdClause, error) {
	p.cmd = app.Command(p.def.Name, p.def.Description)
	for _, a := range p.def.Aliases {
		p.cmd.Alias(a)
	}

	return p.cmd, nil
}