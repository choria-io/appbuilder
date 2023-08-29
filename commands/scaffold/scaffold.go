// Copyright (c) 2023, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package scaffold

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/choria-io/appbuilder/builder"
	"github.com/choria-io/appbuilder/scaffold"
	"github.com/choria-io/fisk"
)

type Command struct {
	Target          string              `json:"target"`
	SourceDirectory string              `json:"source_directory"`
	Source          map[string]any      `json:"source"`
	Post            []map[string]string `json:"post"`

	builder.GenericSubCommands
	builder.GenericCommand
}

type Scaffold struct {
	defnDir   string
	userDir   string
	arguments map[string]any
	flags     map[string]any
	cmd       *fisk.CmdClause
	def       *Command
	ctx       context.Context
	log       builder.Logger
	b         *builder.AppBuilder
}

var (
	ErrorInvalidConfiguration = errors.New("invalid configuration")
	ErrRenderFailed           = errors.New("render failed")
)

func Register() error {
	return builder.RegisterCommand("scaffold", NewScaffoldCommand)
}

func MustRegister() {
	builder.MustRegisterCommand("scaffold", NewScaffoldCommand)
}

func NewScaffoldCommand(b *builder.AppBuilder, j json.RawMessage, log builder.Logger) (builder.Command, error) {
	s := &Scaffold{
		def:       &Command{},
		ctx:       b.Context(),
		defnDir:   b.DefinitionDirectory(),
		userDir:   b.UserWorkingDirectory(),
		b:         b,
		log:       log,
		arguments: map[string]any{},
		flags:     map[string]any{},
	}

	err := json.Unmarshal(j, s.def)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", builder.ErrInvalidDefinition, err)
	}

	return s, nil
}

func (r *Scaffold) String() string { return fmt.Sprintf("%s (scaffold)", r.def.Name) }

func (r *Scaffold) Validate(log builder.Logger) error {
	if r.def.Type != "scaffold" {
		return fmt.Errorf("not an scaffold command")
	}

	var errs []string

	if r.def.Target == "" {
		errs = append(errs, "target is required")
	}

	if len(r.def.Source) == 0 && r.def.SourceDirectory == "" {
		errs = append(errs, "no sources provided")
	}

	if r.def.SourceDirectory != "" {
		_, err := os.Stat(r.def.SourceDirectory)
		if err != nil {
			errs = append(errs, fmt.Sprintf("cannot read source directory: %v", err))
		}
	}

	if _, err := os.Stat(r.def.Target); !os.IsNotExist(err) {
		errs = append(errs, "target directory exist")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, ", "))
	}

	return nil
}

func (r *Scaffold) SubCommands() []json.RawMessage {
	return r.def.Commands
}

func (r *Scaffold) CreateCommand(app builder.KingpinCommand) (*fisk.CmdClause, error) {
	r.cmd = builder.CreateGenericCommand(app, &r.def.GenericCommand, r.arguments, r.flags, r.b, r.runCommand)

	return r.cmd, nil
}

func (r *Scaffold) runCommand(_ *fisk.ParseContext) error {
	cfg := scaffold.Config{
		Source: r.def.Source,
		Post:   r.def.Post,
	}

	var err error

	if cfg.SourceDirectory != "" {
		cfg.SourceDirectory, err = builder.ParseStateTemplate(r.def.SourceDirectory, r.arguments, r.flags, r.b.Configuration())
		if err != nil {
			return err
		}
	}

	cfg.TargetDirectory, err = builder.ParseStateTemplate(r.def.Target, r.arguments, r.flags, r.b.Configuration())
	if err != nil {
		return err
	}

	s, err := scaffold.New(cfg, builder.TemplateFuncs(true))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrorInvalidConfiguration, err)
	}

	s.Logger(builder.NewDefaultLogger())

	err = s.Render(builder.NewTemplateState(r.arguments, r.flags, r.b.Configuration(), nil))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRenderFailed, err)
	}

	return nil
}
