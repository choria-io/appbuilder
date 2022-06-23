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
	"strings"

	"github.com/choria-io/appbuilder/builder"
	"github.com/choria-io/fisk"
	"github.com/kballard/go-shellquote"
)

type Command struct {
	Command     string                    `json:"command"`
	Environment []string                  `json:"environment"`
	Transform   *builder.GenericTransform `json:"transform"`
	Script      string                    `json:"script"`
	Shell       string                    `json:"shell"`

	builder.GenericSubCommands
	builder.GenericCommand
}

type Exec struct {
	Arguments map[string]*string
	Flags     map[string]*string
	cmd       *fisk.CmdClause
	def       *Command
	ctx       context.Context
	log       builder.Logger
	b         *builder.AppBuilder
}

func Register() error {
	return builder.RegisterCommand("exec", NewExecCommand)
}

func MustRegister() {
	builder.MustRegisterCommand("exec", NewExecCommand)
}

var (
	ErrorInvalidCommand  = errors.New("invalid command")
	ErrorTemplateFailed  = errors.New("template error")
	ErrorExecutionFailed = errors.New("execution failed")
)

func NewExecCommand(b *builder.AppBuilder, j json.RawMessage, log builder.Logger) (builder.Command, error) {
	exec := &Exec{
		def:       &Command{},
		ctx:       b.Context(),
		b:         b,
		log:       log,
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

func (r *Exec) Validate(log builder.Logger) error {
	if r.def.Type != "exec" {
		return fmt.Errorf("not an exec command")
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

	if r.def.Command == "" && r.def.Script == "" {
		errs = append(errs, "a command or script is required")
	}

	if r.def.Command != "" && r.def.Script != "" {
		errs = append(errs, "only one of command or script is allowed")
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

func (r *Exec) SubCommands() []json.RawMessage {
	return r.def.Commands
}

func (r *Exec) CreateCommand(app builder.KingpinCommand) (*fisk.CmdClause, error) {
	r.cmd = builder.CreateGenericCommand(app, &r.def.GenericCommand, r.Arguments, r.Flags, r.b.Configuration(), r.runCommand)

	return r.cmd, nil
}

func (r *Exec) logCommand(cmd string, args []string, env []string) {
	r.log.Debugf("Executing command %q", cmd)

	for _, e := range env {
		r.log.Debugf("Environment: %s", e)
	}

	for i, a := range args {
		r.log.Debugf("Argument %d: %v", i, a)
	}
}

func (r *Exec) runInTerminal(cmd string, args []string, env []string) error {
	r.logCommand(cmd, args, env)

	if os.Getenv("BUILDER_DRY_RUN") != "" {
		return fmt.Errorf("%s: dry run mode", ErrorExecutionFailed)
	}

	run := exec.CommandContext(r.ctx, cmd, args...)
	run.Env = append(os.Environ(), env...)
	run.Stdin = os.Stdin
	run.Stdout = os.Stdout
	run.Stderr = os.Stderr

	err := run.Run()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrorExecutionFailed, err)
	}

	return nil
}

func (r *Exec) runWithTransform(cmd string, args []string, env []string) error {
	r.logCommand(cmd, args, env)

	if os.Getenv("BUILDER_DRY_RUN") == "1" {
		return fmt.Errorf("%s: dry run mode", ErrorExecutionFailed)
	}

	run := exec.CommandContext(r.ctx, cmd, args...)
	run.Env = append(os.Environ(), env...)

	run.Stdin = os.Stdin
	run.Stderr = os.Stderr

	out, err := run.Output()
	if err != nil {
		return err
	}

	return r.def.Transform.FTransformJSON(r.ctx, os.Stdout, out)
}

func (r *Exec) findShell() []string {
	if r.def.Shell != "" {
		parts, err := shellquote.Split(r.def.Shell)
		if err != nil {
			return nil
		}

		if len(parts) == 1 {
			parts = append(parts, "-c")
		}

		return parts
	}

	if shell := os.Getenv("SHELL"); shell != "" {
		return []string{shell, "-c"}
	}

	if _, err := os.Stat("/bin/bash"); !os.IsNotExist(err) {
		return []string{"/bin/bash", "-c"}
	}

	return []string{"/bin/sh", "-c"}
}

func (r *Exec) runCommand(_ *fisk.ParseContext) error {
	var cmd string
	var err error
	var parts []string

	if r.def.Command != "" {
		cmd, err = builder.ParseStateTemplate(r.def.Command, r.Arguments, r.Flags, r.b.Configuration())
		if err != nil {
			return fmt.Errorf("%w: %v", ErrorTemplateFailed, err)
		}

		parts, err = shellquote.Split(cmd)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrorInvalidCommand, err)
		}
	} else {
		shell := r.findShell()
		if len(shell) == 0 {
			return fmt.Errorf("cannot determine shell, set SHELL or shell property")
		}

		script, err := builder.ParseStateTemplate(r.def.Script, r.Arguments, r.Flags, r.b.Configuration())
		if err != nil {
			return fmt.Errorf("%w: %v", ErrorTemplateFailed, err)
		}

		parts = append(shell, script)
	}

	if len(parts) == 0 {
		return ErrorInvalidCommand
	}

	var env []string
	for _, e := range r.def.Environment {
		v, err := builder.ParseStateTemplate(e, r.Arguments, r.Flags, r.b.Configuration())
		if err != nil {
			return fmt.Errorf("%w: %v", ErrorTemplateFailed, err)
		}
		env = append(env, v)
	}

	if r.def.Transform == nil {
		return r.runInTerminal(parts[0], parts[1:], env)
	} else {
		return r.runWithTransform(parts[0], parts[1:], env)
	}
}
