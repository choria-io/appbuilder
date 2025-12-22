// Copyright (c) 2022-2025, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/choria-io/fisk"
	"github.com/goccy/go-yaml"
	"github.com/tidwall/gjson"
	"github.com/xlab/tablewriter"
)

type KingpinCommand interface {
	Flag(name, help string) *fisk.FlagClause
	Command(name, help string) *fisk.CmdClause
}

// Command is the interface a command plugin should implement
type Command interface {
	// CreateCommand should add all the flags, sub commands, arguments and more to the app
	CreateCommand(app KingpinCommand) (*fisk.CmdClause, error)
	// SubCommands is the list of defined sub commands, nil if none
	SubCommands() []json.RawMessage
	// Validate should validate the properties of the command after creation
	Validate(Logger) error
	// String should describe the plugin, usually in the form 'name (kind)'
	String() string
}

type TemplateState struct {
	Arguments any
	Flags     any
	Config    any
	Input     any
}

// AppBuilder is the main runner and configuration handler
type AppBuilder struct {
	ctx            context.Context
	def            *Definition
	name           string
	appPath        string
	definitionPath string
	userWorkingDir string
	cfg            map[string]any
	cfgSources     []string
	stdOut         io.Writer
	stdErr         io.Writer
	log            Logger
	exitWithUsage  bool
}

var (
	ErrorDuplicatePlugin  = errors.New("duplicate plugin")
	ErrorUnknownPlugin    = errors.New("unknown plugin")
	ErrDefinitionNotfound = errors.New("definition not found")
	ErrConfigNotFound     = errors.New("config file not found")
	ErrCommandHasNoType   = errors.New("command has no type defined")
	ErrInvalidDefinition  = errors.New("invalid definition")

	Version = "development"

	taskFileNames = []string{
		"ABTaskFile.dist.yaml",
		"ABTaskFile.dist.yml",
		"ABTaskFile.yaml",
		"ABTaskFile.yml",
		"ABTaskFile",
	}

	requireDescription   = true
	appDefPattern        = "%s-app.yaml"
	appCfgPatten         = "%s-cfg.yaml"
	defaultDescription   = ""
	defaultUsageTemplate = fisk.CompactMainUsageTemplate
	descriptionFmt       = `%s

Contact: %s
`
)

// New creates a new CLI Builder
func New(ctx context.Context, name string, opts ...Option) (*AppBuilder, error) {
	builder := &AppBuilder{
		cfg:    make(map[string]any),
		ctx:    ctx,
		name:   name,
		stdOut: os.Stdout,
		stdErr: os.Stderr,
		log:    NewDefaultLogger(),
		cfgSources: []string{
			filepath.Join(xdg.ConfigHome, "appbuilder"),
			"/etc/appbuilder",
		},
	}

	var err error

	builder.userWorkingDir, err = os.Getwd()
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		if opt != nil {
			err = opt(builder)
			if err != nil {
				return nil, err
			}
		}
	}

	return builder, nil
}

// UserWorkingDirectory is the user is in when executing the command
func (b *AppBuilder) UserWorkingDirectory() string {
	return b.userWorkingDir
}

// DefinitionDirectory is the directory where the definition is stored
func (b *AppBuilder) DefinitionDirectory() string {
	if b.definitionPath == "" {
		return ""
	}

	return filepath.Dir(b.definitionPath)
}

// Stdout is the target for writing errors
func (b *AppBuilder) Stdout() io.Writer {
	return b.stdOut
}

// Stderr is the target for writing errors
func (b *AppBuilder) Stderr() io.Writer {
	return b.stdErr
}

// Configuration is the loaded configuration, valid only after LoadConfig() is called, usually done during RunCommand()
func (b *AppBuilder) Configuration() map[string]any {
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

func (b *AppBuilder) CreateBuilderApp(cmd KingpinCommand) {
	validate := cmd.Command("validate", "Validates a application definition").Action(b.validateAction)
	validate.Arg("definition", "Path to the definition to validate").Required().ExistingFileVar(&b.appPath)

	cmd.Command("info", "Shows information about the App Builder setup").Action(b.infoAction)
	cmd.Command("list", "List applications").Action(b.listAction)
}

// RunBuilderCLI runs the builder command, used to validate apps and more
func (b *AppBuilder) RunBuilderCLI() error {
	help := `Choria Application Builder

This is the builder helper allowing you to validate and
inspect the configuration.

If you expected your own command here be sure to create
a symlink from your command name to this file and then 
always invoke the symlink.

For help see https://choria-io.github.io/appbuilder/
`
	cmd := fisk.New(b.name, help)
	cmd.Version(Version)
	cmd.Author("R.I.Pienaar <rip@devco.net>")
	cmd.HelpFlag.Hidden()
	cmd.VersionFlag.Hidden()
	cmd.HelpFlag.Hidden()
	cmd.GetFlag("help").Hidden()
	cmd.UsageWriter(b.stdErr)
	cmd.ErrorWriter(b.stdErr)

	b.CreateBuilderApp(cmd)

	if b.exitWithUsage {
		cmd.MustParseWithUsage(os.Args[1:])
		return nil
	}

	_, err := cmd.Parse(os.Args[1:])
	return err
}

func (b *AppBuilder) listAction(_ *fisk.ParseContext) error {
	sources := append([]string{"."}, b.cfgSources...)
	var found []string

	for _, source := range sources {
		if !fileExist(source) {
			continue
		}

		entries, err := os.ReadDir(source)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), "-app.yaml") {
				abs := filepath.Join(source, entry.Name())
				if err != nil {
					return err
				}

				found = append(found, abs)
			}
		}
	}

	if len(found) == 0 {
		fmt.Println("No Application Definitions found")
	}

	table := tablewriter.CreateTable()
	table.UTF8Box()
	table.AddTitle("Known Applications")
	table.AddHeaders("Name", "Location", "Description")

	for _, source := range found {
		d, err := b.loadDefinition(source)
		if err != nil {
			return err
		}

		app := strings.TrimSuffix(source, "-app.yaml")
		table.AddRow(filepath.Base(app), source, d.Description)
	}

	fmt.Println(table.Render())

	return nil
}

func (b *AppBuilder) infoAction(_ *fisk.ParseContext) error {
	fmt.Println("Choria Application Builder")
	fmt.Println()
	fmt.Printf("        Debug Logging (BUILDER_DEBUG): %t\n", os.Getenv("BUILDER_DEBUG") != "")
	if os.Getenv("BUILDER_CONFIG") != "" {
		fmt.Printf("  Configuration File (BUILDER_CONFIG): %s\n", os.Getenv("BUILDER_CONFIG"))
	} else {
		fmt.Printf("  Configuration File (BUILDER_CONFIG): not specified\n")
	}
	if os.Getenv("BUILDER_APP") != "" {
		fmt.Printf("        Definition File (BUILDER_APP): %s\n", os.Getenv("BUILDER_APP"))
	} else {
		fmt.Printf("        Definition File (BUILDER_APP): not specified\n")
	}

	fmt.Printf("                     Source Locations: %s\n", strings.Join(b.cfgSources, ", "))

	return nil
}

func (b *AppBuilder) validateAction(_ *fisk.ParseContext) error {
	d, err := b.LoadDefinition()
	if err != nil {
		return err
	}

	errs := make(chan string, 10000)

	err = d.Validate(nil)
	if err != nil {
		errs <- err.Error()
	}

	b.validateCommands([]string{"root"}, errs, d.commands...)

	close(errs)

	if len(errs) > 0 {
		fmt.Printf("Application definition %s not valid:\n", b.appPath)
		fmt.Println()
		for e := range errs {
			fmt.Printf("   %v\n", e)
		}

		os.Exit(1)
	} else {
		fmt.Printf("Application definition %s is valid\n", b.appPath)
		return nil
	}

	return err
}

// HasDefinition determines if the named definition can be found on the node
func (b *AppBuilder) HasDefinition() bool {
	name := appDefPattern
	if strings.Contains(name, "%") {
		name = fmt.Sprintf(appDefPattern, b.name)
	}

	source, _ := b.findConfigFile(name, b.appPath)
	if source == "" {
		return false
	}

	return fileExist(source)
}

func (b *AppBuilder) loadDefinitionBytes(cfg []byte, path string) (*Definition, error) {
	d := &Definition{}
	cfgj, err := yaml.YAMLToJSON(cfg)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidDefinition, err)
	}

	err = json.Unmarshal(cfgj, d)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidDefinition, err)
	}

	if d.IncludeFile != "" {
		f, err := os.ReadFile(d.IncludeFile)
		if err != nil {
			return nil, err
		}
		cfgj, err := yaml.YAMLToJSON(f)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidDefinition, err)
		}

		def := &Definition{}
		err = json.Unmarshal(cfgj, def)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidDefinition, err)
		}

		if d.Name != "" {
			def.Name = d.Name
		}
		if d.Author != "" {
			def.Author = d.Author
		}
		if d.Description != "" {
			def.Description = d.Description
		}
		if d.Version != "" {
			def.Version = d.Version
		}

		d = def
	}

	err = b.createCommands(d, d.Commands)
	if err != nil {
		return nil, err
	}

	b.definitionPath = path

	return d, nil
}

func (b *AppBuilder) loadDefinition(source string) (*Definition, error) {
	if b.log != nil {
		b.log.Debugf("Loading application definition %v", source)
	}

	cfg, err := os.ReadFile(source)
	if err != nil {
		return nil, err
	}

	return b.loadDefinitionBytes(cfg, source)
}

// LoadDefinition loads the definition for the name from file, creates the command structure and validates everything
func (b *AppBuilder) LoadDefinition() (*Definition, error) {
	name := appDefPattern
	if strings.Contains(appDefPattern, "%") {
		name = fmt.Sprintf(appDefPattern, b.name)
	}

	source, err := b.findConfigFile(name, b.appPath)
	if err != nil {
		return nil, ErrDefinitionNotfound
	}

	d, err := b.loadDefinition(source)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (b *AppBuilder) createCommands(d *Definition, defs []json.RawMessage) error {
	if len(defs) == 0 {
		return fmt.Errorf("%w: no definitions found", ErrInvalidDefinition)
	}

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
func (b *AppBuilder) LoadConfig() (map[string]any, error) {
	fname := appCfgPatten
	if strings.Contains(appCfgPatten, "%") {
		fname = fmt.Sprintf(appCfgPatten, b.name)
	}

	source, err := b.findConfigFile(fname, "")
	if err != nil || source == "" {
		source, err = b.findConfigFile(fname, "")
		if err != nil {
			return nil, ErrConfigNotFound
		}
	}

	return b.loadConfigFile(source)
}

func (b *AppBuilder) loadConfigFile(source string) (map[string]any, error) {
	b.log.Debugf("Loading configuration file %s", source)

	if !fileExist(source) {
		return nil, ErrConfigNotFound
	}

	cfgb, err := os.ReadFile(source)
	if err != nil {
		return nil, err
	}

	cfgj, err := yaml.YAMLToJSON(cfgb)
	if err != nil {
		return nil, err
	}

	cfg := map[string]any{}

	err = json.Unmarshal(cfgj, &cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (b *AppBuilder) createAppCLI() (*fisk.Application, error) {
	description := b.def.Description
	if description == "" {
		description = defaultDescription
	}

	matches := strings.Count(descriptionFmt, "%")
	if matches == 1 {
		description = fmt.Sprintf(descriptionFmt, description)
	} else if matches == 2 {
		description = fmt.Sprintf(descriptionFmt, description, b.def.Author)
	}

	cmd := fisk.New(b.name, description)
	cmd.Version(b.def.Version)
	cmd.Author(b.def.Author)
	cmd.HelpFlag.Hidden()
	cmd.VersionFlag.Hidden()
	cmd.HelpFlag.Hidden()
	cmd.GetFlag("help").Hidden()
	cmd.UsageWriter(b.stdErr)
	cmd.ErrorWriter(b.stdErr)

	switch strings.TrimSpace(strings.ToLower(b.def.HelpTemplate)) {
	case "", "default":
		cmd.UsageTemplate(defaultUsageTemplate)
		cmd.ErrorUsageTemplate(defaultUsageTemplate)
	case "compact":
		cmd.UsageTemplate(fisk.CompactUsageTemplate)
		cmd.ErrorUsageTemplate(fisk.CompactUsageTemplate)
	case "short":
		cmd.UsageTemplate(fisk.ShorterMainUsageTemplate)
		cmd.ErrorUsageTemplate(fisk.ShorterMainUsageTemplate)
	case "long":
		cmd.UsageTemplate(fisk.KingpinDefaultUsageTemplate)
		cmd.ErrorUsageTemplate(fisk.KingpinDefaultUsageTemplate)
	}

	cheats := b.def.Cheats
	if cheats != nil && (cheats.Enabled || cheats.Cheat != "") {
		cmd.WithCheats(cheats.Tags...)
		cmd.CheatCommand.Hidden()

		name := b.name
		if cheats.Label != "" {
			name = cheats.Label
		}

		cmd.Cheat(name, cheats.Cheat)
		cmd.Help = fmt.Sprintf("%s\n\nUse '%s cheat' to access cheat sheet style help", cmd.Help, b.name)
	}

	err := b.registerCommands(cmd, b.def.commands...)
	if err != nil {
		return nil, err
	}

	return cmd, nil
}

// FiskApplication loads the definition and returns a fisk application
func (b *AppBuilder) FiskApplication() (*fisk.Application, error) {
	var err error

	b.cfg, err = b.LoadConfig()
	if err != nil && !errors.Is(err, ErrConfigNotFound) {
		return nil, err
	}

	b.def, err = b.LoadDefinition()
	if err != nil {
		return nil, err
	}

	return b.createAppCLI()
}

func (b *AppBuilder) runTaskCLI() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	tf, err := b.findFirstTaskFile(wd)
	if err != nil {
		return err
	}

	b.cfg, err = b.loadConfigFile(".abtenv")
	if err != nil && !errors.Is(err, ErrConfigNotFound) {
		return err
	}

	b.def, err = b.loadDefinition(tf)
	if err != nil {
		return err
	}

	cmd, err := b.createAppCLI()
	if err != nil {
		return err
	}

	if b.exitWithUsage {
		cmd.MustParseWithUsage(os.Args[1:])
		return nil
	}

	_, err = cmd.Parse(os.Args[1:])
	return err
}

func (b *AppBuilder) runCLI() error {
	var err error

	cmd, err := b.FiskApplication()
	if err != nil {
		return err
	}

	if b.exitWithUsage {
		cmd.MustParseWithUsage(os.Args[1:])
		return nil
	}

	_, err = cmd.Parse(os.Args[1:])
	return err
}

func (b *AppBuilder) findFirstTaskFile(path string) (string, error) {
	for _, f := range taskFileNames {
		tf := filepath.Join(path, f)
		b.log.Debugf("Looking for task file %s", tf)
		if fileExist(tf) {
			return tf, nil
		}
	}

	parent := filepath.Dir(path)

	if parent != path {
		return b.findFirstTaskFile(parent)
	}

	return "", ErrDefinitionNotfound
}

func (b *AppBuilder) findConfigFile(name string, override string) (string, error) {
	sources := b.cfgSources

	cur, err := filepath.Abs(".")
	if err == nil {
		sources = append([]string{cur}, sources...)
	}

	b.log.Debugf("Searching for config %s in %s with override %q", name, strings.Join(sources, ", "), override)

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

func (b *AppBuilder) validateCommands(bread []string, errs chan string, cmds ...Command) {
	for _, c := range cmds {
		bread = append(bread, c.String())
		b.log.Debugf("Validating %s", c)
		err := c.Validate(b.log)
		if err != nil {
			errs <- fmt.Sprintf("%s: %s", strings.Join(bread, " -> "), err)
		}

		for _, sub := range c.SubCommands() {
			sc, err := b.createCommand(sub)
			if err != nil {
				errs <- fmt.Sprintf("%s: %s", strings.Join(bread, " ->  "), err.Error())
			}

			b.validateCommands(bread, errs, sc)
		}
	}
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

	cmd, err := cons(b, def, b.log)
	if err != nil {
		return nil, err
	}

	return cmd, nil
}

// TemplateFuncs returns standard template funcs, set all to also include sprig functions
func (b *AppBuilder) TemplateFuncs(all bool) template.FuncMap {
	funcs := TemplateFuncs(all)

	funcs["UserWorkingDir"] = func() string {
		return b.UserWorkingDirectory()
	}

	funcs["AppDir"] = func() string {
		return b.DefinitionDirectory()
	}
	funcs["TaskDir"] = funcs["AppDir"]

	return funcs
}
