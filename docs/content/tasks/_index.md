+++
title = "Local Task Mode"
toc = true
weight = 30
pre = "<b>3. </b>"
+++

Some features are ongoing experiments and not part of the supported feature set, this section will call them out.

## Local Task Mode

While it's nice to have a formal machine-wide command that behaves like a normal Unix CLI command I found I would like
to use this same framework to build project specific helpers.

{{% notice secondary "Version Hint" code-branch %}}
This is available since version `0.6.2` and GA since `0.9.0`.
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
