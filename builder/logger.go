// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"log"
)

// Logger is a pluggable logger interface
type Logger interface {
	Debugf(format string, v ...any)
	Infof(format string, v ...any)
	Warnf(format string, v ...any)
	Errorf(format string, v ...any)
}

// Default console logger
type defaultLogger struct{}

func (l *defaultLogger) Infof(format string, v ...any) {
	log.Printf(format, v...)
}

func (l *defaultLogger) Warnf(format string, v ...any) {
	log.Printf(format, v...)
}

func (l *defaultLogger) Errorf(format string, v ...any) {
	log.Printf(format, v...)
}

func (l *defaultLogger) Debugf(format string, v ...any) {
	log.Printf(format, v...)
}

type NoopLogger struct{}

func (n NoopLogger) Debugf(format string, v ...any) {}

func (n NoopLogger) Infof(format string, v ...any) {}

func (n NoopLogger) Warnf(format string, v ...any) {}

func (n NoopLogger) Errorf(format string, v ...any) {}
