+++
title = "Compiled Applications"
weight = 140
toc = true
+++

It's nice that you do not need to compile App Builder apps into binaries as it allows for fast iteration, but sometimes it might be desired.

{{% notice secondary "Version Hint" code-branch %}}
This was added in version 0.7.2
{{% /notice %}}

## Basic compiled application

Given an application in `app.yaml` we can create a small Go stub:

```go
package main

import (
	"context"
	_ "embed"
	"os"

	"github.com/choria-io/appbuilder/builder"
	"github.com/choria-io/fisk"
)

//go:embed app.yaml
var def []byte

func main() {
	builder.MustRegisterStandardCommands()

	cmd := fisk.Newf("myapp", "My compiled App Builder application")

	err := builder.MountAsCommand(context.TODO(), cmd, def, nil)
	if err != nil {
		panic(err)
	}

	cmd.MustParseWithUsage(os.Args[1:])
}
```

When you compile this as a normal Go application your binary will be an executable version of the app.

## Mounting at a sub command

Here we mount the application at the top level of the `myapp` binary, but you could also mount it later on - perhaps you have other compiled in behaviors you wish to surface:

```go
func main() {
	builder.MustRegisterStandardCommands()

	cmd := fisk.Newf("myapp", "My compiled App Builder application")
	embedded := cmd.Command("embedded","Embedded application goes here")

	err := builder.MountAsCommand(context.TODO(), embedded, def, nil)
	if err != nil {
		panic(err)
	}

	cmd.MustParseWithUsage(os.Args[1:])
}
```

Here we would end up with `myapp embedded [app commands]` - the command being mounted at a deeper level in the resulting compiled application.  This way you can plug a App Builder command into any level programmatically.