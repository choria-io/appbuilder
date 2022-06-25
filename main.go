// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
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

	if strings.HasPrefix(name, "appbuilder") {
		builder.RunBuilderCLI(context.Background(), true, builder.WithContextualUsageOnError())
	}

	builder.RunStandardCLI(context.Background(), name, true, nil, builder.WithContextualUsageOnError())
}
