// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package file

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/choria-io/appbuilder/builder"
	"github.com/choria-io/fisk"
)

type Command struct {
	File   string `json:"file"`
	Chroot string `json:"chroot"`
	Action string `json:"action"`

	builder.GenericSubCommands
	builder.GenericCommand
}

type File struct {
	arguments map[string]any
	flags     map[string]any
	cmd       *fisk.CmdClause
	def       *Command
	ctx       context.Context
	log       builder.Logger
	b         *builder.AppBuilder
}

var ErrorInvalidConfiguration = errors.New("invalid configuration")

func Register() error {
	return builder.RegisterCommand("file", NewFileCommand)
}

func MustRegister() {
	builder.MustRegisterCommand("file", NewFileCommand)
}

func NewFileCommand(b *builder.AppBuilder, j json.RawMessage, log builder.Logger) (builder.Command, error) {
	f := &File{
		def:       &Command{},
		ctx:       b.Context(),
		b:         b,
		log:       log,
		arguments: map[string]any{},
		flags:     map[string]any{},
	}

	err := json.Unmarshal(j, f.def)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", builder.ErrInvalidDefinition, err)
	}

	return f, nil
}

func (r *File) String() string { return fmt.Sprintf("%s (file)", r.def.Name) }

func (r *File) Validate(log builder.Logger) error {
	if r.def.Type != "file" {
		return fmt.Errorf("not a file command")
	}

	var errs []string

	err := r.def.GenericCommand.Validate(log)
	if err != nil {
		errs = append(errs, err.Error())
	}

	if r.def.File == "" {
		errs = append(errs, "file is required")
	}

	switch r.def.Action {
	case "":
		r.def.Action = "read"
	case "read":
	default:
		errs = append(errs, fmt.Sprintf("invalid action %q, only read is supported", r.def.Action))
	}

	if len(errs) > 0 {
		return fmt.Errorf("%w: %s", ErrorInvalidConfiguration, strings.Join(errs, ", "))
	}

	return nil
}

func (r *File) SubCommands() []json.RawMessage {
	return r.def.Commands
}

func (r *File) CreateCommand(app builder.KingpinCommand) (*fisk.CmdClause, error) {
	r.cmd = builder.CreateGenericCommand(app, &r.def.GenericCommand, r.arguments, r.flags, r.b, r.runCommand)

	return r.cmd, nil
}

func (r *File) runCommand(_ *fisk.ParseContext) error {
	name, err := r.b.RenderTemplate(r.def.File, r.arguments, r.flags, builder.WithSprig())
	if err != nil {
		return fmt.Errorf("invalid file template: %w", err)
	}

	var chroot string
	if r.def.Chroot != "" {
		chroot, err = r.b.RenderTemplate(r.def.Chroot, r.arguments, r.flags, builder.WithSprig())
		if err != nil {
			return fmt.Errorf("invalid chroot template: %w", err)
		}
	}

	var f *os.File
	if chroot != "" {
		f, err = os.OpenInRoot(chroot, name)
	} else {
		f, err = os.Open(name)
	}
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(r.b.Stdout(), f)

	return err
}
