// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

// RunStandardCLI runs a standard command line instance with shutdown watchers etc. If log is nil a logger will be created
func RunStandardCLI(ctx context.Context, name string, watchInterrupts bool, log Logger, opts ...Option) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if log == nil {
		logger := logrus.New()
		log = logrus.NewEntry(logger).WithField("app", name)
		if os.Getenv("BUILDER_DEBUG") != "" {
			logger.SetLevel(logrus.DebugLevel)
		} else {
			logger.SetLevel(logrus.WarnLevel)
		}
	}

	if watchInterrupts {
		go interruptWatcher(ctx, cancel, log)
	}

	if len(opts) == 0 {
		if cfg := os.Getenv("BUILDER_CONFIG"); cfg != "" {
			opts = append(opts, WithConfigPaths(cfg))
		}
		if cfg := os.Getenv("BUILDER_APP"); cfg != "" {
			opts = append(opts, WithAppDefinitionFile(cfg))
		}
	}

	// we set the logger option first, if we made a new logger above
	// it will be set, if the user supplied one, it will be set
	//
	// but if a user later pass WithLogger() also that one will win
	// as options are processed in array order
	opts = append([]Option{WithLogger(log)}, opts...)

	bldr, err := New(ctx, name, opts...)
	if err != nil {
		return err
	}

	if !bldr.HasDefinition() {
		return fmt.Errorf("%w: %s", ErrDefinitionNotfound, name)
	}

	err = bldr.RunCommand()
	if err != nil {
		return err
	}

	return nil
}

func interruptWatcher(ctx context.Context, cancel context.CancelFunc, log Logger) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	for {
		select {
		case sig := <-sigs:
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				go func() {
					<-time.After(2 * time.Second)
					os.Exit(1)
				}()
			}

			log.Infof("Shutting down on %s", sig)
			cancel()

		case <-ctx.Done():
			return
		}
	}
}
