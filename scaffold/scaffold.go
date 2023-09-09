// Copyright (c) 2023, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package scaffold

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/kballard/go-shellquote"
)

// Config configures a scaffolding operation
type Config struct {
	// TargetDirectory is where to place the resulting rendered files, must not exist
	TargetDirectory string `yaml:"target"`
	// SourceDirectory reads templates from a directory, mutually exclusive with Source
	SourceDirectory string `yaml:"source_directory"`
	// Source reads templates from in-process memory
	Source map[string]any `yaml:"source"`
	// Post configures post processing of files using filepath globs
	Post []map[string]string `yaml:"post"`
	// SkipEmpty skips files that are 0 bytes after rendering
	SkipEmpty bool `yaml:"skip_empty"`
	// Sets a custom template delimiter, useful for generating templates from templates
	CustomLeftDelimiter string `yaml:"left_delimiter"`
	// Sets a custom template delimiter, useful for generating templates from templates
	CustomRightDelimiter string `yaml:"right_delimiter"`
}

type Logger interface {
	Debugf(format string, v ...any)
	Infof(format string, v ...any)
}

var errSkippedEmpty = errors.New("skipped rendering")

type Scaffold struct {
	cfg           *Config
	funcs         template.FuncMap
	log           Logger
	workingSource string
	currentDir    string
}

// New creates a new scaffold instance
func New(cfg Config, funcs template.FuncMap) (*Scaffold, error) {
	if cfg.TargetDirectory == "" {
		return nil, fmt.Errorf("target is required")
	}

	var err error
	cfg.TargetDirectory, err = filepath.Abs(cfg.TargetDirectory)
	if err != nil {
		return nil, fmt.Errorf("invalid target %s: %v", cfg.TargetDirectory, err)
	}

	if len(cfg.Source) == 0 && cfg.SourceDirectory == "" {
		return nil, fmt.Errorf("no sources provided")
	}

	if cfg.SourceDirectory != "" {
		_, err := os.Stat(cfg.SourceDirectory)
		if err != nil {
			return nil, fmt.Errorf("cannot read source directory: %w", err)
		}
	}

	if _, err := os.Stat(cfg.TargetDirectory); !os.IsNotExist(err) {
		return nil, fmt.Errorf("target directory exist")
	}

	return &Scaffold{cfg: &cfg, funcs: funcs}, nil
}

// Logger configures a logger to use, no logging is done without this
func (s *Scaffold) Logger(log Logger) {
	s.log = log
}

func (c *Scaffold) dumpSourceDir(source map[string]any, target string) error {
	for k, v := range source {
		if strings.Contains(k, "..") {
			return fmt.Errorf("invalid file name %v", k)
		}
		if strings.ContainsAny(k, `/\`) {
			return fmt.Errorf("invalid file name %v", k)
		}

		out := filepath.Join(target, k)

		switch e := v.(type) {
		case string: // a file
			err := os.WriteFile(out, []byte(e), 0400)
			if err != nil {
				return err
			}

		case map[string]any: // a directory
			err := os.Mkdir(out, 0700)
			if err != nil {
				return err
			}

			err = c.dumpSourceDir(e, out)
			if err != nil {
				return err
			}

		default: // a mistake
			return fmt.Errorf("invalid source entry %s: %v", k, v)
		}
	}

	return nil
}

func (c *Scaffold) createTempDirForSource() (string, error) {
	td, err := os.MkdirTemp("", "")
	if err != nil {
		return "", err
	}

	err = c.dumpSourceDir(c.cfg.Source, td)
	if err != nil {
		os.RemoveAll(td)
		return "", err
	}

	return td, nil
}

func (c *Scaffold) saveAndPostFile(f string, data string) error {
	err := c.saveFile(f, data)
	if err != nil {
		return err
	}

	return c.postFile(f)
}

func (c *Scaffold) renderAndPostFile(out string, t string, data any) error {
	err := c.renderFile(out, t, data)
	switch {
	case errors.Is(err, errSkippedEmpty):
		if c.log != nil {
			c.log.Infof("Skipping empty file %v", out)
		}

		return nil
	case err != nil:
		return err
	}

	err = c.postFile(out)
	if err != nil {
		return err
	}

	if c.log != nil {
		c.log.Infof("Rendered %s", out)
	}

	return nil
}

func (c *Scaffold) templateFuncs() template.FuncMap {
	if c.funcs == nil {
		return nil
	}

	funcs := template.FuncMap{}
	for k, v := range c.funcs {
		funcs[k] = v
	}

	funcs["write"] = func(out string, content string) (string, error) {
		err := c.saveAndPostFile(filepath.Join(c.cfg.TargetDirectory, out), content)
		return "", err
	}

	funcs["render"] = func(templ string, data any) (string, error) {
		res, err := c.renderTemplateFile(filepath.Join(c.workingSource, templ), data)
		return string(res), err
	}

	return funcs
}

func (c *Scaffold) renderTemplateFile(tmpl string, data any) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	templ := template.New(filepath.Base(tmpl))
	funcs := c.templateFuncs()
	if funcs != nil {
		templ.Funcs(funcs)
	}

	if c.cfg.CustomLeftDelimiter != "" && c.cfg.CustomRightDelimiter != "" {
		templ.Delims(c.cfg.CustomLeftDelimiter, c.cfg.CustomRightDelimiter)
	}

	td, err := os.ReadFile(tmpl)
	if err != nil {
		return nil, err
	}

	templ, err = templ.Parse(string(td))
	if err != nil {
		return nil, fmt.Errorf("parsing template %v failed: %w", tmpl, err)
	}

	err = templ.Execute(buf, data)
	if err != nil {
		return nil, err
	}

	if c.cfg.SkipEmpty && len(bytes.TrimSpace(buf.Bytes())) == 0 {
		return nil, errSkippedEmpty
	}

	return buf.Bytes(), nil
}

func (c *Scaffold) saveFile(out string, content string) error {
	absOut, err := filepath.Abs(out)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(absOut, c.cfg.TargetDirectory) {
		return fmt.Errorf("%s is not in target directory %s", out, c.cfg.TargetDirectory)
	}

	return os.WriteFile(out, []byte(content), 0755)
}

func (c *Scaffold) renderFile(out string, t string, data any) error {
	res, err := c.renderTemplateFile(t, data)
	if err != nil {
		return err
	}

	return c.saveFile(out, string(res))
}

func (c *Scaffold) postFile(f string) error {
	for _, p := range c.cfg.Post {
		for g, v := range p {
			matched, err := filepath.Match(g, filepath.Base(f))
			if err != nil {
				return err
			}

			if !matched {
				continue
			}

			cmd := ""
			args := []string{}

			parts, err := shellquote.Split(strings.ReplaceAll(v, "{}", f))
			if err != nil {
				return err
			}
			cmd = parts[0]
			if len(parts) > 1 {
				args = append(args, parts[1:]...)
			}

			if !strings.Contains(v, "{}") {
				args = append(args, f)
			}

			if c.log != nil {
				c.log.Infof("Post processing using: %s %s", cmd, strings.Join(args, " "))
			}

			out, err := exec.Command(cmd, args...).CombinedOutput()
			if err != nil {
				return fmt.Errorf("failed to post process %s\nerror: %w\noutput: %q", f, err, out)
			}
		}
	}

	return nil
}

// Render creates the target directory and place all files into it after template processing and post processing
func (c *Scaffold) Render(data any) error {
	err := os.MkdirAll(c.cfg.TargetDirectory, 0770)
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	err = os.Chdir(c.cfg.TargetDirectory)
	if err != nil {
		return err
	}
	defer os.Chdir(cwd)

	c.workingSource = c.cfg.SourceDirectory

	if c.workingSource == "" {
		// move the memory source to temp dir
		c.workingSource, err = c.createTempDirForSource()
		if err != nil {
			return err
		}
		defer func() {
			os.RemoveAll(c.workingSource)
			c.workingSource = ""
		}()
	}

	c.currentDir = c.cfg.TargetDirectory
	defer func() { c.currentDir = "" }()

	// now render both the same way
	err = filepath.WalkDir(c.workingSource, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == c.workingSource {
			return nil
		}

		if d.Name() == "_partials" {
			return filepath.SkipDir
		}

		out := filepath.Join(c.cfg.TargetDirectory, strings.TrimPrefix(path, c.workingSource))

		switch {
		case d.IsDir():
			err := os.Mkdir(out, 0775)
			if err != nil {
				return err
			}

		case d.Type().IsRegular():
			c.currentDir = filepath.Dir(out)
			err = c.renderAndPostFile(out, path, data)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("invalid file in source: %v", d.Name())
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
