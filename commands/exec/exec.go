// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/choria-io/appbuilder/builder"
	"github.com/kballard/go-shellquote"
	"gopkg.in/alecthomas/kingpin.v2"
)

type Command struct {
	Command string `json:"command"`

	builder.GenericSubCommands
	builder.GenericCommand
}

type Exec struct {
	Arguments map[string]*string
	Flags     map[string]*string
	cmd       *kingpin.CmdClause
	def       *Command
	ctx       context.Context
	b         *builder.CLIBuilder
}

func MustRegister() {
	builder.MustRegisterCommand("exec", NewExecCommand)
}

var (
	ErrorInvalidCommand  = errors.New("invalid command")
	ErrorTemplateFailed  = errors.New("template error")
	ErrorExecutionFailed = errors.New("execution failed")
)

func NewExecCommand(b *builder.CLIBuilder, j json.RawMessage, _ builder.Logger) (builder.Command, error) {
	exec := &Exec{
		def:       &Command{},
		ctx:       b.Context(),
		b:         b,
		Arguments: map[string]*string{},
		Flags:     map[string]*string{},
	}

	err := json.Unmarshal(j, exec.def)
	if err != nil {
		return nil, err
	}

	return exec, nil
}

func (r *Exec) String() string { return fmt.Sprintf("%s (exec)", r.def.Name) }

func (r *Exec) Validate(log builder.Logger) error { return nil }

func (r *Exec) SubCommands() []json.RawMessage {
	return r.def.Commands
}

func (r *Exec) CreateCommand(app builder.KingpinCommand) (*kingpin.CmdClause, error) {
	r.cmd = builder.CreateGenericCommand(app, &r.def.GenericCommand, r.Arguments, r.Flags, r.runCommand)

	return r.cmd, nil
}

func (r *Exec) runCommand(_ *kingpin.ParseContext) error {
	cmd, err := builder.ParseStateTemplate(r.def.Command, r.Arguments, r.Flags, r.b.Configuration())
	if err != nil {
		return fmt.Errorf("%w: %v", ErrorTemplateFailed, err)
	}

	parts, err := shellquote.Split(cmd)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrorInvalidCommand, err)
	}
	if len(parts) == 0 {
		return ErrorInvalidCommand
	}

	run := exec.CommandContext(r.ctx, parts[0], parts[1:]...)
	run.Env = os.Environ()
	run.Stdin = os.Stdin
	run.Stdout = os.Stdout
	run.Stderr = os.Stderr

	err = run.Run()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrorExecutionFailed, err)
	}

	return nil
}
