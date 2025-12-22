// Copyright (c) 2022-2025, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package parent

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/choria-io/appbuilder/builder"
	"github.com/choria-io/fisk"
	"github.com/goccy/go-yaml"
)

type Command struct {
	IncludeFile string `json:"include_file"`
	builder.GenericSubCommands
	builder.GenericCommand
}

type Parent struct {
	cmd *fisk.CmdClause
	def *Command
}

func Register() {
	builder.RegisterCommand("parent", NewParentCommand)
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
		return nil, fmt.Errorf("%w: %v", builder.ErrInvalidDefinition, err)
	}

	if parent.def.IncludeFile != "" {
		err = parent.includeCommands()
		if err != nil {
			return nil, err
		}
		parent.def.IncludeFile = ""
	}

	return parent, nil
}

func (p *Parent) includeCommands() error {
	b, err := os.ReadFile(p.def.IncludeFile)
	if err != nil {
		return err
	}

	name := p.def.Name

	cfgj, err := yaml.YAMLToJSON(b)
	if err != nil {
		return err
	}

	err = json.Unmarshal(cfgj, p.def)
	if err != nil {
		return err
	}

	p.def.Name = name

	return nil
}

func (p *Parent) String() string { return fmt.Sprintf("%s (parent)", p.def.Name) }

func (p *Parent) Validate(log builder.Logger) error {
	if p.def.Type != "parent" {
		return fmt.Errorf("not a parent command")
	}

	var errs []string

	err := p.def.GenericCommand.Validate(log)
	if err != nil {
		errs = append(errs, err.Error())
	}

	err = p.def.GenericSubCommands.Validate(log)
	if err != nil {
		errs = append(errs, err.Error())
	}

	if len(p.def.Flags) > 0 {
		errs = append(errs, "parent commands can not have flags")
	}

	if len(p.def.Arguments) > 0 {
		errs = append(errs, "parent commands can not have arguments")
	}

	if len(p.def.Commands) == 0 {
		errs = append(errs, "parent requires sub commands")
	}

	if p.def.ConfirmPrompt != "" {
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

func (p *Parent) CreateCommand(app builder.KingpinCommand) (*fisk.CmdClause, error) {
	p.cmd = app.Command(p.def.Name, p.def.Description)
	for _, a := range p.def.Aliases {
		p.cmd.Alias(a)
	}

	return p.cmd, nil
}
