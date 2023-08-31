// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/choria-io/appbuilder/builder"
	"github.com/choria-io/appbuilder/commands"
)

func main() {
	name := filepath.Base(os.Args[0])

	var err error

	commands.MustRegisterStandardCommands()

	if strings.HasPrefix(name, "appbuilder") {
		err = builder.RunBuilderCLI(context.Background(), true, builder.WithContextualUsageOnError())
	} else if strings.HasPrefix(name, "abt") {
		err = builder.RunTaskCLI(context.Background(), true, builder.WithContextualUsageOnError())
	} else {
		err = builder.RunStandardCLI(context.Background(), name, true, nil, builder.WithContextualUsageOnError())
	}

	if errors.Is(err, builder.ErrInvalidDefinition) {
		fmt.Fprintln(os.Stderr, "error: Invalid definition, please use validate to test your definition")
		os.Exit(1)
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
