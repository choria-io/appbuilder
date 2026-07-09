// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/choria-io/fisk"
	"github.com/choria-io/validator"
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
	Tags          []string             `json:"tags,omitempty"`
	Secrets       []GenericSecret      `json:"secrets,omitempty"`
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

	for _, a := range c.Arguments {
		errs = append(errs, validateInput("argument", a.Name, a.Type, a.Default, len(a.Enum) > 0, false)...)
	}

	for _, f := range c.Flags {
		if len(f.Short) > 1 {
			errs = append(errs, fmt.Sprintf("short flag for %s must be 1 character", f.Name))
		}
		errs = append(errs, validateInput("flag", f.Name, f.Type, f.Default, len(f.Enum) > 0, f.Bool)...)
	}

	seenSecrets := map[string]struct{}{}
	for _, s := range c.Secrets {
		err := s.Validate()
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}

		if _, dup := seenSecrets[s.Name]; dup {
			errs = append(errs, fmt.Sprintf("duplicate secret name %q", s.Name))
		}
		seenSecrets[s.Name] = struct{}{}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, ", "))
	}

	return nil
}

// GenericArgument is a standard command line argument
type GenericArgument struct {
	Name                 string   `json:"name"`
	Description          string   `json:"description"`
	Required             bool     `json:"required"`
	Enum                 []string `json:"enum"`
	Default              any      `json:"default"`
	ValidationExpression string   `json:"validate"`
	Type                 string   `json:"type"`
}

// GenericFlag is a standard command line flag
type GenericFlag struct {
	Name                 string   `json:"name"`
	Description          string   `json:"description"`
	Required             bool     `json:"required"`
	PlaceHolder          string   `json:"placeholder"`
	Enum                 []string `json:"enum"`
	Default              any      `json:"default"`
	Bool                 bool     `json:"bool"`
	EnvVar               string   `json:"env"`
	Short                string   `json:"short"`
	ValidationExpression string   `json:"validate"`
	Type                 string   `json:"type"`
}

// parserClause is the shared surface of fisk's *ArgClause and *FlagClause that we use to
// configure a typed input. Both satisfy it via the embedded parserMixin.
type parserClause interface {
	String() *string
	Bool() *bool
	UnNegatableBool() *bool
	Enum(...string) *string
	ExistingFile() *string
	ExistingDir() *string
	Counter() *int
	Int() *int
	Int8() *int8
	Int16() *int16
	Int32() *int32
	Int64() *int64
	Uint() *uint
	Uint8() *uint8
	Uint16() *uint16
	Uint32() *uint32
	Uint64() *uint64
	Float() *float64
	Float32() *float32
	Float64() *float64
}

var (
	_ parserClause = (*fisk.ArgClause)(nil)
	_ parserClause = (*fisk.FlagClause)(nil)
)

// inputTypes is the single source of truth for the flag and argument type names and the
// fisk parser that enforces each. bool and enum are handled separately by applyInputType
// because they depend on the default value and the enum list.
var inputTypes = map[string]func(parserClause) any{
	"string":        func(c parserClause) any { return c.String() },
	"existing_file": func(c parserClause) any { return c.ExistingFile() },
	"existing_dir":  func(c parserClause) any { return c.ExistingDir() },
	"counter":       func(c parserClause) any { return c.Counter() },
	"int":           func(c parserClause) any { return c.Int() },
	"integer":       func(c parserClause) any { return c.Int() },
	"int8":          func(c parserClause) any { return c.Int8() },
	"int16":         func(c parserClause) any { return c.Int16() },
	"int32":         func(c parserClause) any { return c.Int32() },
	"int64":         func(c parserClause) any { return c.Int64() },
	"uint":          func(c parserClause) any { return c.Uint() },
	"uint8":         func(c parserClause) any { return c.Uint8() },
	"uint16":        func(c parserClause) any { return c.Uint16() },
	"uint32":        func(c parserClause) any { return c.Uint32() },
	"uint64":        func(c parserClause) any { return c.Uint64() },
	"float":         func(c parserClause) any { return c.Float() },
	"float32":       func(c parserClause) any { return c.Float32() },
	"float64":       func(c parserClause) any { return c.Float64() },
}

// normalizeType lower-cases and trims a user supplied type so matching is forgiving and
// CreateGenericCommand and Validate always agree on which type was requested.
func normalizeType(t string) string {
	return strings.ToLower(strings.TrimSpace(t))
}

// knownInputType reports whether dType, already normalized, is one of the mapped types.
func knownInputType(dType string) bool {
	_, ok := inputTypes[dType]
	return ok
}

// knownTypeNames returns all supported type names including bool, sorted for stable help
// and error output.
func knownTypeNames() []string {
	names := make([]string, 0, len(inputTypes)+1)
	for name := range inputTypes {
		names = append(names, name)
	}
	names = append(names, "bool")
	sort.Strings(names)

	return names
}

// isTrueDefault reports whether a default represents boolean true, used to decide if a bool
// should be negatable. It mirrors fisk which parses bool defaults via strconv.ParseBool.
func isTrueDefault(dflt any) bool {
	switch v := dflt.(type) {
	case bool:
		return v
	case string:
		b, err := strconv.ParseBool(v)
		return err == nil && b
	default:
		return false
	}
}

// applyInputType configures c for the requested type and returns the fisk value pointer.
// enum takes precedence over type, then bool, then the mapped types, finally a plain string.
func applyInputType(c parserClause, dType string, enum []string, dflt any) any {
	switch {
	case len(enum) > 0:
		return c.Enum(enum...)
	case dType == "bool":
		if isTrueDefault(dflt) {
			return c.Bool()
		}
		return c.UnNegatableBool()
	}

	if fn, ok := inputTypes[dType]; ok {
		return fn(c)
	}

	return c.String()
}

// defaultHint renders a default value for an error message, avoiding the scientific notation
// fmt uses for large numbers so the suggested quoted value stays usable.
func defaultHint(dflt any) string {
	if f, ok := dflt.(float64); ok {
		return strconv.FormatFloat(f, 'f', -1, 64)
	}

	return fmt.Sprintf("%v", dflt)
}

// validateInput checks the type and default of a flag or argument. kind is "flag" or
// "argument" and is used in messages, legacyBool is the deprecated bool flag field.
func validateInput(kind, name, typ string, dflt any, hasEnum, legacyBool bool) []string {
	var errs []string

	// fisk only takes string defaults, so numbers and the like must be quoted to reach it
	// unambiguously. Booleans are allowed as a convenience.
	switch dflt.(type) {
	case nil, string, bool:
	default:
		errs = append(errs, fmt.Sprintf("%s %q default must be a string or boolean, quote the value like default: %q", kind, name, defaultHint(dflt)))
	}

	dType := normalizeType(typ)
	if dType == "" {
		return errs
	}

	switch {
	case hasEnum && !(dType == "string" || dType == ""):
		errs = append(errs, fmt.Sprintf("%s %q sets both type and enum, remove one", kind, name))
	case legacyBool && dType != "bool":
		errs = append(errs, fmt.Sprintf("%s %q sets both bool and type %q, remove one", kind, name, dType))
	case dType == "bool", knownInputType(dType):
		// supported
	default:
		errs = append(errs, fmt.Sprintf("%s %q has unknown type %q, valid types are: %s", kind, name, dType, strings.Join(knownTypeNames(), ", ")))
	}

	return errs
}

// CreateGenericCommand can be used to add all the typical flags and arguments etc if your command is based on GenericCommand. Values set in flags and arguments
// are created on the supplied maps, if flags or arguments is nil then this will not attempt to add defined flags. Use this if you wish to use GenericCommand as
// a base for your own commands while perhaps using an extended argument set
func CreateGenericCommand(app KingpinCommand, sc *GenericCommand, arguments map[string]any, flags map[string]any, b *AppBuilder, cb fisk.Action) *fisk.CmdClause {
	description := sc.Description
	if len(sc.Secrets) > 0 {
		description = fmt.Sprintf("%s\n\nRequires the 1Password CLI and an active session.", description)
	}

	cmd := app.Command(sc.Name, description).Action(runWrapper(*sc, arguments, flags, b, cb))
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
			arg := cmd.Arg(a.Name, a.Description)
			if a.Required {
				arg.Required()
			}

			if a.Default != nil {
				arg.Default(fmt.Sprintf("%v", a.Default))
			}

			if a.ValidationExpression != "" {
				arg.Validator(validator.FiskValidator(a.ValidationExpression))
			}

			arguments[a.Name] = applyInputType(arg, normalizeType(a.Type), a.Enum, a.Default)
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

			if f.Default != nil {
				flag.Default(fmt.Sprintf("%v", f.Default))
			}

			if f.EnvVar != "" {
				flag.Envar(f.EnvVar)
			}

			if f.Short != "" {
				flag.Short([]rune(f.Short)[0])
			}

			if f.ValidationExpression != "" {
				flag.Validator(validator.FiskValidator(f.ValidationExpression))
			}

			dType := normalizeType(f.Type)
			if f.Bool {
				dType = "bool"
			}

			flags[f.Name] = applyInputType(flag, dType, f.Enum, f.Default)
		}
	}

	for _, t := range sc.Tags {
		cmd.Tag(t)
	}

	return cmd
}

func runWrapper(cmd GenericCommand, arguments map[string]any, flags map[string]any, b *AppBuilder, handler fisk.Action) fisk.Action {
	return func(pc *fisk.ParseContext) error {
		f := dereferenceArgsOrFlags(flags)

		// Reset first so a reused builder (library/test usage) never bleeds the previous
		// command's secrets into this one if resolution is skipped below.
		b.secrets = nil

		if cmd.Banner != "" {
			// Banners render before resolution so .Secrets is intentionally unavailable here.
			txt, err := b.RenderTemplate(cmd.Banner, arguments, flags, WithSprig())
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

		// Resolve after the confirm prompt so no op/biometric prompt fires before the user
		// confirms, and before the handler so secret-bearing templates can use .Secrets.
		if len(cmd.Secrets) > 0 {
			if os.Getenv("BUILDER_DRY_RUN") != "" {
				b.secrets = dryRunSecrets(cmd.Secrets)
			} else {
				secrets, err := resolveSecrets(b.Context(), cmd.Secrets)
				if err != nil {
					return err
				}
				b.secrets = secrets
			}
		}

		return handler(pc)
	}
}
