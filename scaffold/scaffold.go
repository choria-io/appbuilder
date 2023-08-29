// Copyright (c) 2023, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package scaffold

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/kballard/go-shellquote"
)

// Config comfigures a scaffolding operation
type Config struct {
	// TargetDirectory is where to place the resulting rendered files, must not exist
	TargetDirectory string `yaml:"target"`
	// SourceDirectory reads templates from a directory, mutually exclusive with Source
	SourceDirectory string `yaml:"source_directory"`
	// Source reads templates from in-process memory
	Source map[string]any `yaml:"source"`
	// Post configures post processing of files using filepath globs
	Post []map[string]string `yaml:"post"`
}

type Logger interface {
	Debugf(format string, v ...any)
	Infof(format string, v ...any)
}

type Scaffold struct {
	cfg   *Config
	funcs template.FuncMap
	log   Logger
}

// New creates a new scaffold instance
func New(cfg Config, funcs template.FuncMap) (*Scaffold, error) {
	if cfg.TargetDirectory == "" {
		return nil, fmt.Errorf("target is required")
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

func (c *Scaffold) renderFile(out string, t string, data any) error {
	f, err := os.OpenFile(out, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	td, err := os.ReadFile(t)
	if err != nil {
		return err
	}

	templ := template.New("x")
	if c.funcs != nil {
		templ.Funcs(c.funcs)
	}

	templ, err = templ.Parse(string(td))
	if err != nil {
		return fmt.Errorf("parsing template %v failed: %w", t, err)
	}

	return templ.Execute(f, data)
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

	source := c.cfg.SourceDirectory

	if source == "" {
		// move the memory source to temp dir
		source, err = c.createTempDirForSource()
		if err != nil {
			return err
		}
		defer os.RemoveAll(source)
	}

	// now render both the same way
	err = filepath.WalkDir(source, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == source {
			return nil
		}

		out := filepath.Join(c.cfg.TargetDirectory, strings.TrimPrefix(path, source))

		switch {
		case d.IsDir():
			err := os.Mkdir(out, 0775)
			if err != nil {
				return err
			}

		case d.Type().IsRegular():
			err = c.renderFile(out, path, data)
			if err != nil {
				return err
			}

			err = c.postFile(out)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("invalid file in source: %v", d.Name())
		}

		if c.log != nil {
			c.log.Infof("Rendered %s", out)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
