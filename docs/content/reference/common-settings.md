+++
title = "Common Settings"
toc = true
weight = 10
+++

## Command Types

As we saw above we have `parent` and `exec` types of commands. Users can add more but these are the core ones we support today.

Most commands are made up of a generic set of options and then have one or more added in addition to specialise them.

### Common properties reference

Most commands include a standard set of fields - those that do not or have special restritions will mention in the docs.

Lets look at how we can produce this command:

```nohighlight
usage: demo say [<flags>] <message>

Says something using the cowsay command

The command called defaults to cowsay but can be configured using the Cowsay configuration item

Flags:
  --help             Show context-sensitive help (also try --help-long and --help-man).
  --cowfile=FILE     Use a specific cow file

Args:
  <message>  The message to display
```

It's made up of a `commands` member that has these properties:

```yaml
name: example
description: Example application
version: 1.0.0
author: Operations team <ops@example.net>
help_template: default # optional

commands:
  - 
    # The name in the command: 'example say ....' (required)
    name: say

    # Help showng in output of 'example help say' or 'example say --help` (required)
    description: |
      Says something using the cowsay command

      The command called defaults to cowsay but can be
      configured using the Cowsay configuration item

    # Selects the kind of command, see below (required)
    type: exec # or any other known type

    # Optionally you can run 'example say hello' or 'example s hello' (optional)
    aliases:
     - s 

    # Arguments to accept (optional)
    arguments:
     - name: message
       description: The message to display
       required: true

    # Flags to accept (optional)
    flags:
      - name: cowfile
        description: Use a specific cow file
        placeholder: FILE

    # Sub commands to create below this one (optional, but see specific references)
    commands: []
```

Here we show the initial options that define the application followed by commands.  All the top settings are required except `help_template`, it's value may be one of `compact`, `long`, `short` or `default`.  When not set it will equal `default`. Experiment with these options to see which help format suits your app best (requires version 0.0.9). 

You can emit a banner before invoking the commands in an exec, use this to show a warning or extra information to users before running a command.  Perhaps to warn them that a config override is in use like here:

```yaml
  - name: say
    description: Say something using the configured command
    type: exec
    command: |
      {{ default .Config.Cowsay "cowsay" }} {{ .Arguments.message | escape }}
    banner: |
      {{- if (default .Config.Cowsay "") -}}
      >>
      >> Using the {{ .Config.Cowsay }} command
      >>
      {{- end -}}
    arguments:
      - name: message
        description: The message to send to the terminal
        required: true
```

We support Cheat Sheet style help, see the [dedicated guide](../cheats/) about that.

#### Arguments

An `argument` is a positional input to a command. `example say hello`, when the command is `say` the `hello` would be the first argument.

Arguments can have many options, the table below detail them and the version that added them.

| Option        | Description                                                                                                             | Required | Version |
|---------------|-------------------------------------------------------------------------------------------------------------------------|----------|---------|
| `name`        | A unique name for each flag                                                                                             | yes      |         |
| `description` | A description for this flag, typically 1 line                                                                           | yes      |         |
| `required`    | Indicates that a value for this flag must be set, which includes being set from default                                 |          |         |
| `enum`        | An array of valid values, if set the flag must be one of these values                                                   |          | 0.0.4   |
| `default`     | Sets a default value when not passed, will satisfy enums and required. For bools must be `true` or `false`              |          | 0.0.4   |


#### Flags

A `flag` is a option passed to the application using something like `--flag`, typically these are used for optional inputs. Flags can have many options, the table below detail them and the version that added them.

| Option        | Description                                                                                                             | Required | Version |
|---------------|-------------------------------------------------------------------------------------------------------------------------|----------|---------|
| `name`        | A unique name for each flag                                                                                             | yes      |         |
| `description` | A description for this flag, typically 1 line                                                                           | yes      |         |
| `required`    | Indicates that a value for this flag must be set, which includes being set from default                                 |          |         |
| `placeholder` | Will show this text in the help output like `--cowfile=FILE`                                                            |          |         |
 | `enum`        | An array of valid values, if set the flag must be one of these values                                                   |          | 0.0.4   |
| `default`     | Sets a default value when not passed, will satisfy enums and required. For bools must be `true` or `false`              |          | 0.0.4   |
| `bool`        | Indicates that the flag is a boolean (see below)                                                                        |          | 0.1.1   |
| `env`         | Will load the value from an environment variable if set, passing the flag specifically wins, then the env, then default |          | 0.1.2   |
| `short`       | A single character that can be used instead of the `name` to access this flag. ie. `--cowfile` might also be `-F`       |          | 0.1.2   |

##### Boolean Flags

```yaml
  - name: delete
    description: Delete the data
    type: exec
    command: |
      {{if .Flags.force}}
      rm -rfv /nonexisting
      {{else}}
      echo "Please pass --force to delete the data"
      {{end}}
    flags:
      - name: force
        description: Required to pass when removing data
        bool: true
```

Here we have a `--force` flag that is used to influence the command.  Booleans can have their default set to `true` or `"true`" which will then add a `--no-flag-name` option added to negate it.

#### Argument and Flag Validations

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

| Function             | Description                                                   |
|----------------------|---------------------------------------------------------------|
| `isIP(value)`        | Checks if `value` is an IPv4 or IPv6 address                  |
| `isIPv4(value)`      | Checks if `value` is an IPv4 address                          |
| `isIPv6(value)`      | Checks if `value` is an IPv6 address                          |
| `isInt(value)`       | Checks if `value` is an Integer                               |
| `isFloat(value)`     | Checks if `value` is a Float                                  |
| `isShellSafe(value)` | Checks if `value` is attempting to to do shell escape attacks |

### Confirmations

You can prompt for confirmation from a user for performing an action:

```yaml
  - name: delete
    description: Delete the data
    type: exec
    confirm_prompt: "Really?"
    command: rm -rf /nonexisting
```

Before running the command the user will be prompted to confirm he wish to do it.  Since version `0.2.0` an option will
be added to the CLI allowing you to skip the prompt using `--no-prompt`.

