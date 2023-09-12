+++
title = "Experiments"
toc = true
weight = 30
pre = "<b>3. </b>"
+++

Some features are ongoing experiments and not part of the supported feature set, this section will call them out.

## Local Task Mode

While it's nice to have a formal machine-wide command that behaves like a normal Unix CLI command I found I would like
to use this same framework to build project specific helpers.

{{% notice secondary "Version Hint" code-branch %}}
This is available since version `0.6.2`.
{{% /notice %}}

Imagine you have a development project and have utility commands to update dependencies, serve the documentation in
preview mode, run tests or build custom binaries.  This is a lot of different commands and tools to learn, wouldn't it
be nice if there was a single command you can run in any of your projects to get a project specific custom app?

```nohighlight
$ abt
usage: abt [<flags>] <command> [<args> ...]

App Builder Task

Help: https://choria-io.github.io/appbuilder

Commands:
  help [<command>...]
  dependencies
    update [<flags>]
  test [<dir>]
  docs
    serve [<flags>]
  build
    binary [<flags>]
    snapshot
```

Here I run `abt` in this project directory, if I ran it elsewhere or in my home directory I would get a different command.

This isn't targeting general build pipelines but rather a way to make per project/directory utilities.

The full capabilities of the core App Builder definitions are available, the only thing that really change is how definitions and configurations are found.

### App Definition Locations

The `abt` command will search from the current directory upward until it finds one of these files:

 * `ABTaskFile.dist.yaml`
 * `ABTaskFile.dist.yml`
 * `ABTaskFile.yaml`
 * `ABTaskFile.yml`
 * `ABTaskFile`

In this manner a project can ship a default task file and users can provide local overrides.

It's common that a task file will want to run something in a known directory relative to its location, to facilitate this
the `exec` command has some new template behaviors that can be used with the new `dir` property to achieve this regardless
of the working directory the user is in.

 * `{{ UserWorkingDir }}` - the directory the user ran the command in
 * `{{ AppDir }}` or `{{ TaskDir }}` - the directory the task file is located in

An example can be found in the source repository for this project.

### Configuration

Configuration is looked for in the local directory in the `.abtenv` file.  At present this is not searched for in parent
directories.

## Argument and Flag Validations

One might need to ensure that the input provided by a user passes some validation, for example when passing commands
to shell scripts one has to be careful about [Shell Injection](https://en.wikipedia.org/wiki/Code_injection#Shell_injection).

We support custom validators on Arguments and Flags using the [Expr Language](https://expr.medv.io/docs/Language-Definition)

{{% notice secondary "Version Hint" code-branch %}}
This is available since version `0.8.0`.
{{% /notice %}}

Based on the Getting Started example that calls `cowsay` we might wish to limit the length of the message to what 
would work well with `cowsay` and also ensure there is no shell escaping happening.

```yaml
arguments:
 - name: message
   description: The message to display
   required: true
   validate: len(value) < 20 && is_shellsafe(value)
```
We support the standard `expr` language grammar - that has a large number of functions that can assist the 
validation needs - we then add a few extra functions that makes sense for operation teams.

In each case accessing `value` would be the value passed from the user

| Function              | Description                                                     |
|-----------------------|-----------------------------------------------------------------|
| `is_ip(value)`        | Checks if `value` is a IPv4 or IPv6 address                     |
| `is_ipv4(value)`      | Checks if `value` is a IPv4 address                             |
| `is_ipv6(value)`      | Checks if `value` is a IPv6 address                             |
| `is_shellsafe(value)` | Checks if `value` is a attempting to to do shell escape attacks |

## Compiled Applications

It's nice that you do not need to compile App Builder apps into binaries as it allows for fast iteration, but sometimes
it might be desired.

As of version `0.7.2` we support compiling binaries that contain an application.

Given an application in `app.yaml` we can create a small go stub:

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

Here we mount the application at the top level of the `myapp` binary, but you could also mount it later on - perhaps you
have other compiled in behaviors you wish to surface:

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

Here we would end up with `myapp embedded [app commands]` - the command being mounted at a deeper level in the resulting
compiled application.  This way you can plug a App Builder command into any level programmatically.