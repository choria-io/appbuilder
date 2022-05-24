// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/choria-io/appbuilder/builder"
	"github.com/choria-io/appbuilder/commands/exec"
	"github.com/choria-io/appbuilder/commands/parent"
)

func main() {
	parent.MustRegister()
	exec.MustRegister()

	name := filepath.Base(os.Args[0])
	var err error

	if strings.HasPrefix(name, "appbuilder") {
		err = builder.RunBuilderCLI(context.Background(), true)
	} else {
		err = builder.RunStandardCLI(context.Background(), name, true, nil)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s failed: %v\n", name, err)
		os.Exit(1)
	}
}
