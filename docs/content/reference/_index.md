+++
title = "Reference"
toc = true
weight = 20
pre = "<b>2. </b>"
+++

## Command Components

The system primarily is there to execute a type of command somewhere in the hierarchy of a CLI tools sub commands.

Lets say we have an app called `demo` that has commands `demo say` and `demo think` the `say` and `think` bits are commands. In this example these are commands of type `exec` - they run a shell command.

If we had a command `demo deploy status` and `demo deploy upgrade` then generally the `deploy` would not perform any action, it's there mainly to achor sub commands and show help information. Here the `deploy` command would be of type `parent`.

Generally I would suggest nested commands are structured as `root -> parent -> parent -> exec` and never `root -> parent -> exec -> exec`. If you do decide to do that I strongly suggest the first exec is a read only action like showing some status. User should feel safe to execute parents without unintended side effects.

## Flags and Arguments

Often we need to pass some parameters to commands, for example if we have one to upgrade some software it might be `demo upgrade 1.2.3`.  Here the `1.2.3` is an argument, you can have a number of arguments and they can be set to be required or optional.  If you have multiple arguments an optional one can not be before a required one.

Flags are generally kept for optional items like `demo upgrade 1.2.3 --channel=nightly`, here we pass a flag `--channel`. At present we only support flags with string values.  We intend to support enums of valid values and boolean flags.
