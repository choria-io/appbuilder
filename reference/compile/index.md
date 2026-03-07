# Compiled Applications

App Builder apps do not need to be compiled into binaries, which allows for fast iteration, but sometimes compilation might be desired.

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

Compiling this as a normal Go application produces a binary that is an executable version of the app.

## Mounting at a sub command

The previous example mounts the application at the top level of the `myapp` binary, but it can also be mounted at a sub-command level - perhaps there are other compiled-in behaviors to surface:

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

Here we would end up with `myapp embedded [app commands]` - the command being mounted at a deeper level in the resulting compiled application.  This way an App Builder command can be plugged into any level programmatically.