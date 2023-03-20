// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"

	"github.com/choria-io/appbuilder/builder"
	"github.com/choria-io/fisk"
	"github.com/kballard/go-shellquote"
)

//go:embed bash_helpers.sh
var bashHelper []byte

type Backoff struct {
	MaxAttempts uint   `json:"max_attempts"`
	PolicySteps uint   `json:"steps"`
	PolicyMin   string `json:"min_sleep"`
	PolicyMax   string `json:"max_sleep"`
}

type Command struct {
	Command     string             `json:"command"`
	Environment []string           `json:"environment"`
	Transform   *builder.Transform `json:"transform"`
	Script      string             `json:"script"`
	Shell       string             `json:"shell"`
	Backoff     *Backoff           `json:"backoff"`
	WorkingDir  string             `json:"dir"`
	NoHelper    bool               `json:"no_helper"`

	builder.GenericSubCommands
	builder.GenericCommand
}

type Exec struct {
	defnDir    string
	userDir    string
	arguments  map[string]any
	flags      map[string]any
	cmd        *fisk.CmdClause
	def        *Command
	ctx        context.Context
	log        builder.Logger
	bo         *policy
	helperPath string
	b          *builder.AppBuilder
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
	ErrorHelperFailed    = errors.New("saving helper script failed")
)

func NewExecCommand(b *builder.AppBuilder, j json.RawMessage, log builder.Logger) (builder.Command, error) {
	exec := &Exec{
		def:       &Command{},
		ctx:       b.Context(),
		defnDir:   b.DefinitionDirectory(),
		userDir:   b.UserWorkingDirectory(),
		b:         b,
		log:       log,
		arguments: map[string]any{},
		flags:     map[string]any{},
	}

	err := json.Unmarshal(j, exec.def)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", builder.ErrInvalidDefinition, err)
	}

	err = exec.configureBackoff()
	if err != nil {
		return nil, err
	}

	return exec, nil
}

func (r *Exec) configureBackoff() error {
	if r.def.Backoff == nil {
		return nil
	}

	if r.def.Backoff.PolicyMin == "" {
		r.def.Backoff.PolicyMin = "500ms"
	}
	if r.def.Backoff.PolicyMax == "" {
		r.def.Backoff.PolicyMax = "20s"
	}
	if r.def.Backoff.PolicySteps == 0 {
		r.def.Backoff.PolicySteps = r.def.Backoff.MaxAttempts
	}

	min, err := time.ParseDuration(r.def.Backoff.PolicyMin)
	if err != nil {
		return err
	}
	max, err := time.ParseDuration(r.def.Backoff.PolicyMax)
	if err != nil {
		return err
	}

	r.bo, err = newPolicy(r.def.Backoff.PolicySteps, min, max)
	if err != nil {
		return err
	}

	return nil
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

	if r.def.Backoff != nil {
		if r.def.Backoff.PolicySteps < 2 {
			errs = append(errs, fmt.Sprintf("invalid backoff policy steps: '%d'", r.def.Backoff.PolicySteps))
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
	r.cmd = builder.CreateGenericCommand(app, &r.def.GenericCommand, r.arguments, r.flags, r.b, r.runCommand)

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
	run.Stdout = r.b.Stdout()
	run.Stderr = r.b.Stderr()
	run.Dir = r.def.WorkingDir

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
	run.Stderr = r.b.Stderr()
	run.Dir = r.def.WorkingDir

	out, err := run.Output()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrorExecutionFailed, err)
	}

	tRes, err := r.def.Transform.TransformBytes(r.ctx, out, r.arguments, r.flags, r.b.Configuration())
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(r.b.Stdout(), string(tRes))
	return err
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

func (r *Exec) funcMap() template.FuncMap {
	return template.FuncMap{
		"UserWorkingDir": func() string {
			return r.userDir
		},
		"AppDir": func() string {
			return r.defnDir
		},
		"TaskDir": func() string {
			return r.defnDir
		},
		"BashHelperPath": func() string {
			return r.helperPath
		},
	}
}

func (r *Exec) runCommand(_ *fisk.ParseContext) error {
	var cmd string
	var err error
	var parts []string

	if !r.def.NoHelper {
		tf, err := os.CreateTemp(r.userDir, "appbuilder-*")
		if err != nil {
			return fmt.Errorf("%w: %v", ErrorHelperFailed, err)
		}
		defer os.Remove(tf.Name())

		_, err = tf.Write(bashHelper)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrorHelperFailed, err)
		}
		tf.Close()
		r.helperPath = tf.Name()
	}

	if r.def.Command != "" {
		cmd, err = builder.ParseStateTemplateWithFuncMap(r.def.Command, r.arguments, r.flags, r.b.Configuration(), r.funcMap())
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

		script, err := builder.ParseStateTemplateWithFuncMap(r.def.Script, r.arguments, r.flags, r.b.Configuration(), r.funcMap())
		if err != nil {
			return fmt.Errorf("%w: %v", ErrorTemplateFailed, err)
		}

		parts = append(shell, script)
	}

	if r.def.WorkingDir != "" {
		r.def.WorkingDir, err = builder.ParseStateTemplateWithFuncMap(r.def.WorkingDir, r.arguments, r.flags, r.b.Configuration(), r.funcMap())
		if err != nil {
			return err
		}
		r.log.Debugf("Running command in directory %s", r.def.WorkingDir)
	}

	if len(parts) == 0 {
		return ErrorInvalidCommand
	}

	var env []string
	for _, e := range r.def.Environment {
		v, err := builder.ParseStateTemplateWithFuncMap(e, r.arguments, r.flags, r.b.Configuration(), r.funcMap())
		if err != nil {
			return fmt.Errorf("%w: %v", ErrorTemplateFailed, err)
		}
		env = append(env, v)
	}

	try := 1
	for {
		if r.def.Transform == nil {
			err = r.runInTerminal(parts[0], parts[1:], append(env, fmt.Sprintf("BUILDER_TRY=%d", try)))
		} else {
			err = r.runWithTransform(parts[0], parts[1:], append(env, fmt.Sprintf("BUILDER_TRY=%d", try)))
		}

		// if it was good or we dont have backoff just return whatever is there
		if err == nil || r.def.Backoff == nil {
			return err
		}

		// we only retry on ExitError, others are returned
		if !errors.Is(err, ErrorExecutionFailed) {
			return err
		}

		d := r.bo.duration(try)
		r.log.Warnf("Execution failed on try %d / %d, retrying after %v based on backoff policy: %v", try, r.def.Backoff.MaxAttempts, d, err)

		if uint(try) >= r.def.Backoff.MaxAttempts {
			r.log.Errorf("Failing after %d tries", try)
			return err
		}

		err = r.bo.sleep(r.ctx, d)
		if err != nil {
			return err
		}
		try++
	}
}
