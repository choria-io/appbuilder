// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"io"
)

// Option configures the builder
type Option func(*AppBuilder) error

// WithConfigPaths overrides the path to the app configuration file, should be a full absolute path
func WithConfigPaths(paths ...string) Option {
	return func(b *AppBuilder) error {
		b.cfgSources = paths
		return nil
	}
}

// WithLogger sets a custom logger to use
func WithLogger(logger Logger) Option {
	return func(b *AppBuilder) error {
		b.log = logger
		return nil
	}
}

// WithAppDefinitionBytes uses a provided app definition rather than load one from disk
func WithAppDefinitionBytes(def []byte) Option {
	return func(b *AppBuilder) (err error) {
		b.def, err = b.loadDefinitionBytes(def, "embedded")

		return err
	}
}

// WithAppDefinitionFile sets a file where the definition should be loaded from
func WithAppDefinitionFile(f string) Option {
	return func(b *AppBuilder) error {
		b.appPath = f
		return nil
	}
}

// WithContextualUsageOnError handles application termination by showing contextual help rather than returning an error
func WithContextualUsageOnError() Option {
	return func(b *AppBuilder) error {
		b.exitWithUsage = true
		return nil
	}
}

// WithStdout configures a standard out handle to output
func WithStdout(w io.Writer) Option {
	return func(b *AppBuilder) error {
		b.stdOut = w
		return nil
	}
}

// WithStderr configures a standard error out handle to output
func WithStderr(w io.Writer) Option {
	return func(b *AppBuilder) error {
		b.stdErr = w
		return nil
	}
}
