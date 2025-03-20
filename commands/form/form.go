// Copyright (c) 2023, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package form

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/choria-io/appbuilder/builder"
	"github.com/choria-io/fisk"
	"github.com/choria-io/scaffold/forms"
)

type Command struct {
	Properties []forms.Property `json:"properties"`

	Transform *builder.Transform `json:"transform"`

	builder.GenericCommand
	builder.GenericSubCommands
}

type Form struct {
	arguments map[string]any
	flags     map[string]any
	defnDir   string
	userDir   string
	cmd       *fisk.CmdClause
	def       *Command
	defBytes  []byte

	ctx context.Context
	log builder.Logger
	b   *builder.AppBuilder
}

func Register() error {
	return builder.RegisterCommand("form", NewFormCommand)
}

func MustRegister() {
	builder.MustRegisterCommand("form", NewFormCommand)
}

func NewFormCommand(b *builder.AppBuilder, j json.RawMessage, log builder.Logger) (builder.Command, error) {
	form := &Form{
		arguments: map[string]any{},
		flags:     map[string]any{},
		defnDir:   b.DefinitionDirectory(),
		userDir:   b.UserWorkingDirectory(),
		def:       &Command{},
		defBytes:  j,
		ctx:       b.Context(),
		log:       log,
		b:         b,
	}

	err := json.Unmarshal(j, form.def)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", builder.ErrInvalidDefinition, err)
	}

	return form, nil
}

func (r *Form) String() string { return fmt.Sprintf("%s (form)", r.def.Name) }

func (r *Form) Validate(log builder.Logger) error {
	if r.def.Type != "form" {
		return fmt.Errorf("not a form command")
	}

	var errs []string

	if len(r.def.Properties) == 0 {
		errs = append(errs, "no form properties")
	}

	err := r.def.GenericCommand.Validate(log)
	if err != nil {
		errs = append(errs, err.Error())
	}

	err = r.def.GenericSubCommands.Validate(log)
	if err != nil {
		errs = append(errs, err.Error())
	}

	if r.def.Transform != nil {
		err := r.def.Transform.Validate(log)
		if err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, ", "))
	}

	return nil
}

func (r *Form) SubCommands() []json.RawMessage {
	return r.def.Commands
}

func (r *Form) CreateCommand(app builder.KingpinCommand) (*fisk.CmdClause, error) {
	r.cmd = builder.CreateGenericCommand(app, &r.def.GenericCommand, r.arguments, r.flags, r.b, r.runCommand)

	return r.cmd, nil
}

func (r *Form) runCommand(_ *fisk.ParseContext) error {
	state := builder.NewTemplateState(r.arguments, r.flags, r.b.Configuration(), nil)

	defBytes, err := builder.ParseStateTemplateWithFuncMap(string(r.defBytes), r.arguments, r.flags, r.b.Configuration(), r.b.TemplateFuncs(true))
	if err != nil {
		return fmt.Errorf("%w: %v", builder.ErrInvalidDefinition, err)
	}

	err = json.Unmarshal([]byte(defBytes), r.def)
	if err != nil {
		return fmt.Errorf("%w: %v", builder.ErrInvalidDefinition, err)
	}

	form := forms.Form{
		Name:        r.def.Name,
		Description: r.def.Description,
		Properties:  r.def.Properties,
	}

	result, err := forms.ProcessForm(form, map[string]any{
		"Arguments": state.Arguments,
		"Flags":     state.Flags,
		"Config":    state.Config,
	})
	if err != nil {
		return err
	}

	rj, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	tRes := rj

	if r.def.Transform != nil {
		tRes, err = r.def.Transform.TransformBytes(r.ctx, rj, r.arguments, r.flags, r.b)
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprint(r.b.Stdout(), string(tRes))
	return err
}
