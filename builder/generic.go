// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/choria-io/fisk"
)

// GenericSubCommands is the typical sub commands most commands support, custom plugins can choose to use this if they support sub commands
type GenericSubCommands struct {
	Commands []json.RawMessage `json:"commands"`
}

// Validate is a noop here
func (c *GenericSubCommands) Validate(logger Logger) error {
	return nil
}

type GenericCommandCheat struct {
	Label string `json:"label,omitempty"`
	Cheat string `json:"cheat"`
}

// GenericCommand is a typical command with the minimal options all supported
type GenericCommand struct {
	Name          string               `json:"name"`
	Description   string               `json:"description"`
	Aliases       []string             `json:"aliases"`
	Type          string               `json:"type"`
	Arguments     []GenericArgument    `json:"arguments,omitempty"`
	Flags         []GenericFlag        `json:"flags,omitempty"`
	ConfirmPrompt string               `json:"confirm_prompt"`
	Banner        string               `json:"banner"`
	Cheat         *GenericCommandCheat `json:"cheat,omitempty"`
}

// Validate ensures the command is well-formed
func (c *GenericCommand) Validate(logger Logger) error {
	var errs []string
	if c.Name == "" {
		errs = append(errs, "name is required")
	}

	if requireDescription && c.Description == "" {
		errs = append(errs, "description is required")
	}
	if c.Type == "" {
		errs = append(errs, "command type is required")
	}

	if c.Cheat != nil && c.Cheat.Cheat == "" {
		errs = append(errs, "cheats require a body")
	}

	for _, f := range c.Flags {
		if len(f.Short) > 1 {
			errs = append(errs, fmt.Sprintf("short flag for %s must be 1 character", f.Name))
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, ", "))
	}

	return nil
}

// GenericArgument is a standard command line argument
type GenericArgument struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Required    bool     `json:"required"`
	Enum        []string `json:"enum"`
	Default     string   `json:"default"`
}

// AddToFiskCommand adds an argument to a command
func (a *GenericArgument) AddToFiskCommand(cmd *fisk.CmdClause, arguments map[string]any) {
	arg := cmd.Arg(a.Name, a.Description)
	if a.Required {
		arg.Required()
	}

	if a.Default != "" {
		arg.Default(a.Default)
	}

	switch {
	case len(a.Enum) > 0:
		arguments[a.Name] = arg.Enum(a.Enum...)
	default:
		arguments[a.Name] = arg.String()
	}
}

// GenericFlag is a standard command line flag
type GenericFlag struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Required    bool     `json:"required"`
	PlaceHolder string   `json:"placeholder"`
	Enum        []string `json:"enum"`
	Default     any      `json:"default"`
	Bool        bool     `json:"bool"`
	EnvVar      string   `json:"env"`
	Short       string   `json:"short"`
}

// AddToFiskCommand adds a flag to a command
func (f *GenericFlag) AddToFiskCommand(cmd *fisk.CmdClause, flags map[string]any) {
	flag := cmd.Flag(f.Name, f.Description)
	if f.Required {
		flag.Required()
	}

	if f.PlaceHolder != "" {
		flag.PlaceHolder(f.PlaceHolder)
	}

	if f.Default != nil {
		flag.Default(fmt.Sprintf("%v", f.Default))
	}

	if f.EnvVar != "" {
		flag.Envar(f.EnvVar)
	}

	if f.Short != "" {
		flag.Short([]rune(f.Short)[0])
	}

	switch {
	case len(f.Enum) > 0:
		flags[f.Name] = flag.Enum(f.Enum...)
	case f.Bool:
		if f.Default == true || f.Default == "true" {
			flags[f.Name] = flag.Bool()
		} else {
			flags[f.Name] = flag.UnNegatableBool()
		}
	default:
		flags[f.Name] = flag.String()
	}
}

// CreateGenericCommand can be used to add all the typical flags and arguments etc if your command is based on GenericCommand. Values set in flags and arguments
// are created on the supplied maps, if flags or arguments is nil then this will not attempt to add defined flags. Use this if you wish to use GenericCommand as
// a base for your own commands while perhaps using an extended argument set
func CreateGenericCommand(app KingpinCommand, sc *GenericCommand, arguments map[string]any, flags map[string]any, b *AppBuilder, cb fisk.Action) *fisk.CmdClause {
	cmd := app.Command(sc.Name, sc.Description).Action(runWrapper(*sc, arguments, flags, b, cb))
	for _, a := range sc.Aliases {
		cmd.Alias(a)
	}

	if sc.Cheat != nil {
		name := sc.Name
		if sc.Cheat.Label != "" {
			name = sc.Cheat.Label
		}

		cmd.Cheat(name, sc.Cheat.Cheat)
	}

	if sc.ConfirmPrompt != "" {
		flags["prompt"] = cmd.Flag("prompt", "Disables the interactive confirmation prompt").Default("true").Bool()
	}

	if arguments != nil {
		for _, a := range sc.Arguments {
			a.AddToFiskCommand(cmd, arguments)
		}
	}

	if flags != nil {
		for _, f := range sc.Flags {
			f.AddToFiskCommand(cmd, flags)
		}
	}

	return cmd
}

func runWrapper(cmd GenericCommand, arguments map[string]any, flags map[string]any, b *AppBuilder, handler fisk.Action) fisk.Action {
	return func(pc *fisk.ParseContext) error {
		f := dereferenceArgsOrFlags(flags)

		if cmd.Banner != "" {
			txt, err := ParseStateTemplate(cmd.Banner, arguments, flags, b.Configuration())
			if err != nil {
				return err
			}

			if txt != "" {
				fmt.Fprintln(b.stdOut, txt)
			}
		}

		if cmd.ConfirmPrompt != "" && f["prompt"] == true {
			ans := false
			err := survey.AskOne(&survey.Confirm{Message: cmd.ConfirmPrompt, Default: false}, &ans)
			if err != nil {
				return err
			}
			if !ans {
				return fmt.Errorf("aborted")
			}
		}

		return handler(pc)
	}
}
