+++
title = "Reference"
description = "reference for command components, flags, and arguments"
toc = true
weight = 20
pre = "<b>2. </b>"
+++

This section covers the core building blocks of an App Builder application - commands, flags, and arguments.

## Command Components

The system executes a type of command somewhere in the hierarchy of a CLI tool's sub commands.

Consider an app called `demo` that has commands `demo say` and `demo think` - the `say` and `think` parts are commands. In this example these are commands of type `exec` - they run a shell command.

Given a command `demo deploy status` and `demo deploy upgrade`, the `deploy` command would not perform any action. It exists mainly to anchor sub commands and show help information. Here the `deploy` command would be of type `parent`.

Nested commands should be structured as `root -> parent -> parent -> exec` and never `root -> parent -> exec -> exec`. When deviating from this pattern, the first exec should be a read-only action like showing some status. Users should feel safe to execute parents without unintended side effects.

## Flags and Arguments

Commands often need parameters. For example, a software upgrade command might look like `demo upgrade 1.2.3`. Here the `1.2.3` is an argument. Commands can have a number of arguments, and they can be set to be required or optional. When multiple arguments exist, an optional one cannot appear before a required one.

Flags are generally kept for optional items like `demo upgrade 1.2.3 --channel=nightly`, where `--channel` is a flag. At present only flags with string values are supported. Future versions intend to support enums of valid values and boolean flags.
