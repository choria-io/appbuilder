// Copyright (c) 2023, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"github.com/choria-io/appbuilder/commands/exec"
	"github.com/choria-io/appbuilder/commands/form"
	"github.com/choria-io/appbuilder/commands/parent"
	"github.com/choria-io/appbuilder/commands/scaffold"
)

func MustRegisterStandardCommands() {
	parent.MustRegister()
	exec.MustRegister()
	scaffold.MustRegister()
	form.MustRegister()
}
