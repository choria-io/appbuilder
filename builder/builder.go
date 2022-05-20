// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/ghodss/yaml"
	"github.com/tidwall/gjson"
	"gopkg.in/alecthomas/kingpin.v2"
)

type KingpinCommand interface {
	Flag(name, help string) *kingpin.FlagClause
	Command(name, help string) *kingpin.CmdClause
}

// Command is the interface a command plugin should implement
type Command interface {
	// CreateCommand should add all the flags, sub commands, arguments and more to the app
	CreateCommand(app KingpinCommand) (*kingpin.CmdClause, error)
	// SubCommands is the list of defined sub commands, nil if none
	SubCommands() []json.RawMessage
	// Validate should validate the properties of the command after creation
	Validate(Logger) error
	// String should describe the plugin, usually in the form 'name (kind)'
	String() string
}

type templateState struct {
	Arguments interface{}
	Flags     interface{}
	Config    interface{}
}

// AppBuilder is the main runner and configuration handler
type AppBuilder struct {
	ctx        context.Context
	def        *Definition
	name       string
	appPath    string
	cfg        map[string]interface{}
	cfgSources []string
	log        Logger
}

var (
	ErrorDuplicatePlugin  = errors.New("duplicate plugin")
	ErrorUnknownPlugin    = errors.New("unknown plugin")
	ErrDefinitionNotfound = errors.New("definition not found")
	ErrConfigNotFound     = errors.New("config file not found")
	ErrCommandHasNoType   = errors.New("command has no type defined")
	ErrInvalidDefinition  = errors.New("invalid definition")

	appDefPattern  = "%s-app.yaml"
	appCfgPatten   = "%s-cfg.yaml"
	descriptionFmt = `%s

Contact: %s
`
)

// New creates a new CLI Builder
func New(ctx context.Context, name string, opts ...Option) (*AppBuilder, error) {
	builder := &AppBuilder{
		cfg:  make(map[string]interface{}),
		ctx:  ctx,
		name: name,
		log:  &defaultLogger{},
		cfgSources: []string{
			filepath.Join(xdg.ConfigHome, "appbuilder"),
			"/etc/appbuilder",
		},
	}

	for _, opt := range opts {
		err := opt(builder)
		if err != nil {
			return nil, err
		}
	}

	return builder, nil
}

// Configuration is the loaded configuration, valid only after LoadConfig() is called, usually done during RunCommand()
func (b *AppBuilder) Configuration() map[string]interface{} {
	return b.cfg
}

// Context gives access to the context used to control app execution and shutdown
func (b *AppBuilder) Context() context.Context {
	return b.ctx
}

// RunCommand prepares the CLI and runs it, including parsing all flags etc
func (b *AppBuilder) RunCommand() error {
	return b.runCLI()
}

// HasDefinition determines if the named definition can be found on the node
func (b *AppBuilder) HasDefinition() bool {
	source, _ := b.findConfigFile(fmt.Sprintf(appDefPattern, b.name), b.appPath)
	if source == "" {
		return false
	}

	return fileExist(source)
}

// LoadDefinition loads the definition for the name from file, creates the command structure and validates everything
func (b *AppBuilder) LoadDefinition() (*Definition, error) {
	source, err := b.findConfigFile(fmt.Sprintf(appDefPattern, b.name), b.appPath)
	if err != nil {
		return nil, ErrDefinitionNotfound
	}

	if b.log != nil {
		b.log.Debugf("Loading application definition %v", source)
	}

	cfg, err := os.ReadFile(source)
	if err != nil {
		return nil, err
	}

	d := &Definition{}
	cfgj, err := yaml.YAMLToJSON(cfg)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(cfgj, d)
	if err != nil {
		return nil, err
	}

	return d, b.createCommands(d, d.Commands)
}

func (b *AppBuilder) createCommands(d *Definition, defs []json.RawMessage) error {
	for _, c := range defs {
		cmd, err := b.createCommand(c)
		if err != nil {
			return err
		}

		d.commands = append(d.commands, cmd)
	}

	return nil
}

// LoadConfig loads the configuration if possible, does not error if nothing is found only if loading fails
func (b *AppBuilder) LoadConfig() (map[string]interface{}, error) {
	fname := fmt.Sprintf(appCfgPatten, b.name)

	source, err := b.findConfigFile(fname, "")
	if err != nil || source == "" {
		source, err = b.findConfigFile(fname, "")
		if err != nil {
			return nil, ErrConfigNotFound
		}
	}

	b.log.Debugf("Loading configuration file %s", source)

	cfgb, err := os.ReadFile(source)
	if err != nil {
		return nil, err
	}

	cfgj, err := yaml.YAMLToJSON(cfgb)
	if err != nil {
		return nil, err
	}

	cfg := map[string]interface{}{}

	err = json.Unmarshal(cfgj, &cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (b *AppBuilder) runCLI() error {
	var err error

	b.cfg, err = b.LoadConfig()
	if err != nil && !errors.Is(err, ErrConfigNotFound) {
		return err
	}

	b.def, err = b.LoadDefinition()
	if err != nil {
		return err
	}

	cmd := kingpin.New(b.name, fmt.Sprintf(descriptionFmt, b.def.Description, b.def.Author))
	cmd.Version(b.def.Version)
	cmd.Author(b.def.Author)
	cmd.VersionFlag.Hidden()

	err = b.registerCommands(cmd, b.def.commands...)
	if err != nil {
		return err
	}

	_, err = cmd.Parse(os.Args[1:])
	return err
}

func (b *AppBuilder) findConfigFile(name string, override string) (string, error) {
	sources := b.cfgSources

	cur, err := filepath.Abs(".")
	if err == nil {
		sources = append([]string{cur}, sources...)
	}

	if b.log != nil {
		b.log.Debugf("Searching for config %s in %s with override %q", name, strings.Join(sources, ", "), override)
	}

	source := override
	if source == "" {
		for _, s := range sources {
			path := filepath.Join(s, name)
			b.log.Debugf("Looking for file %s", path)
			if fileExist(path) {
				source = path
				break
			}
		}
	}

	if source == "" {
		return "", fmt.Errorf("%w: %s in %s", ErrConfigNotFound, name, strings.Join(sources, ", "))
	}

	return source, nil
}

func (b *AppBuilder) registerCommands(cli KingpinCommand, cmds ...Command) error {
	bread := []string{"root"}

	for _, c := range cmds {
		bread = append(bread, c.String())
		b.log.Debugf("Registering %s", c)
		err := c.Validate(b.log)
		if err != nil {
			return fmt.Errorf("%w: %s: %s", ErrInvalidDefinition, strings.Join(bread, " -> "), err)
		}

		cmd, err := c.CreateCommand(cli)
		if err != nil {
			return err
		}

		subs := c.SubCommands()
		if len(subs) > 0 {
			for _, sub := range subs {
				subCommand, err := b.createCommand(sub)
				if err != nil {
					return err
				}

				err = b.registerCommands(cmd, subCommand)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (b *AppBuilder) createCommand(def json.RawMessage) (Command, error) {
	t := gjson.GetBytes(def, "type")
	if !t.Exists() {
		return nil, fmt.Errorf("%w:\n%s", ErrCommandHasNoType, string(def))
	}

	cons, err := commandByKind(t.String())
	if err != nil {
		return nil, err
	}

	return cons(b, def, b.log)
}
