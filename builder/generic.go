// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/itchyny/gojq"
	"gopkg.in/alecthomas/kingpin.v2"
)

// GenericSubCommands is the typical sub commands most commands support, custom plugins can choose to use this if they support sub commands
type GenericSubCommands struct {
	Commands []json.RawMessage `json:"commands"`
}

// Validate is a noop here
func (c *GenericSubCommands) Validate(logger Logger) error {
	return nil
}

// GenericCommand is a typical command with the minimal options all supported
type GenericCommand struct {
	Name          string            `json:"name"`
	Description   string            `json:"description"`
	Aliases       []string          `json:"aliases"`
	Type          string            `json:"type"`
	Arguments     []GenericArgument `json:"arguments"`
	Flags         []GenericFlag     `json:"flags"`
	ConfirmPrompt string            `json:"confirm_prompt"`
	Banner        string            `json:"banner"`
}

// Validate ensures the command is well-formed
func (c *GenericCommand) Validate(logger Logger) error {
	var errs []string
	if c.Name == "" {
		errs = append(errs, "name is required")
	}

	if c.Description == "" {
		errs = append(errs, "description is required")
	}
	if c.Type == "" {
		errs = append(errs, "command type is required")
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

// GenericFlag is a standard command line flag
type GenericFlag struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Required    bool     `json:"required"`
	PlaceHolder string   `json:"placeholder"`
	Enum        []string `json:"enum"`
	Default     string   `json:"default"`
}

// GenericTransform is a generic transformation definition
type GenericTransform struct {
	Query string `json:"query"`
	q     *gojq.Query
}

// Validate parses and validates the JQ query
func (t *GenericTransform) Validate(log Logger) error {
	if t == nil || t.Query == "" {
		return fmt.Errorf("no query supplied")
	}

	var err error

	t.q, err = gojq.Parse(t.Query)
	if err != nil {
		return err
	}

	return nil
}

// FTransformJSON transforms json input via query and write the output to the writer
func (t *GenericTransform) FTransformJSON(ctx context.Context, w io.Writer, j json.RawMessage) error {
	if t.q == nil {
		return fmt.Errorf("no query")
	}

	input := map[string]interface{}{}
	err := json.Unmarshal(j, &input)
	if err != nil {
		return fmt.Errorf("json output parse error: %v", err)
	}

	iter := t.q.RunWithContext(ctx, input)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}

		switch val := v.(type) {
		case error:
			return val
		case string:
			fmt.Fprintln(w, val)
		default:
			j, err := json.MarshalIndent(val, "", "  ")
			if err != nil {
				return err
			}
			fmt.Fprintln(w, string(j))
		}
	}

	return nil
}

// CreateGenericCommand can be used to add all the typical flags and arguments etc if your command is based on GenericCommand. Values set in flags and arguments
// are created on the supplied maps, if flags or arguments is nil then this will not attempt to add defined flags. Use this if you wish to use GenericCommand as
// a base for your own commands while perhaps using an extended argument set
func CreateGenericCommand(app KingpinCommand, sc *GenericCommand, arguments map[string]*string, flags map[string]*string, cfg map[string]interface{}, cb kingpin.Action) *kingpin.CmdClause {
	cmd := app.Command(sc.Name, sc.Description).Action(runWrapper(*sc, arguments, flags, cfg, cb))
	for _, a := range sc.Aliases {
		cmd.Alias(a)
	}

	if arguments != nil {
		for _, a := range sc.Arguments {
			arg := cmd.Arg(a.Name, a.Description)
			if a.Required {
				arg.Required()
			}

			if a.Default != "" {
				arg.Default(a.Default)
			}

			if len(a.Enum) > 0 {
				arguments[a.Name] = arg.Enum(a.Enum...)
			} else {
				arguments[a.Name] = arg.String()
			}
		}
	}

	if flags != nil {
		for _, f := range sc.Flags {
			flag := cmd.Flag(f.Name, f.Description)
			if f.Required {
				flag.Required()
			}

			if f.PlaceHolder != "" {
				flag.PlaceHolder(f.PlaceHolder)
			}

			if f.Default != "" {
				flag.Default(f.Default)
			}

			if len(f.Enum) > 0 {
				flags[f.Name] = flag.Enum(f.Enum...)
			} else {
				flags[f.Name] = flag.String()
			}
		}
	}

	return cmd
}

func runWrapper(cmd GenericCommand, arguments map[string]*string, flags map[string]*string, cfg map[string]interface{}, handler kingpin.Action) kingpin.Action {
	return func(pc *kingpin.ParseContext) error {
		if cmd.ConfirmPrompt != "" {
			ans := false
			err := survey.AskOne(&survey.Confirm{Message: cmd.ConfirmPrompt, Default: false}, &ans)
			if err != nil {
				return err
			}
			if !ans {
				return fmt.Errorf("aborted")
			}
		}

		if cmd.Banner != "" {
			b, err := ParseStateTemplate(cmd.Banner, arguments, flags, cfg)
			if err != nil {
				return err
			}

			if b != "" {
				fmt.Println(b)
			}
		}

		return handler(pc)
	}
}