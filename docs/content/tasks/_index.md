+++
title = "Local Task Mode"
description = "project-specific task runner"
toc = true
weight = 30
pre = "<b>3. </b>"
+++

The `abt` command runs project-specific App Builder definitions, providing per-directory CLI utilities without installing a machine-wide command. It is not a general build pipeline tool - it focuses on wrapping project-specific operational commands.

Development projects often require utility commands to update dependencies, serve documentation in preview mode, run tests, or build custom binaries. `abt` wraps these tools into a single project-specific CLI.

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

Running `abt` in different directories produces different commands based on the task file found in each project. The full capabilities of the core App Builder definitions are available. The only difference from standard mode is how definitions and configurations are located.

## App Definition Locations

The `abt` command searches from the current directory upward until it finds one of these files:

* `ABTaskFile.dist.yaml`
* `ABTaskFile.dist.yml`
* `ABTaskFile.yaml`
* `ABTaskFile.yml`
* `ABTaskFile`

A project can ship a default task file and users can provide local overrides.

Since version `0.14.0`, the environment variable `ABTaskFile=/some/file.yaml` can be set to run a specific file without searching local directories.

Task files often need to run commands relative to the task file location. The `exec` command provides template variables and a `dir` property to handle this regardless of the working directory:

* `{{ UserWorkingDir }}` - the directory the user ran the command in
* `{{ AppDir }}` or `{{ TaskDir }}` - the directory the task file is located in

An example can be found in the [source repository](https://github.com/choria-io/appbuilder).

## Configuration

Configuration is read from the `.abtenv` file in the local directory. Parent directories are not searched for this file.
