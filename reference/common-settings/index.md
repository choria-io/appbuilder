# Common Settings

Application definitions share a set of common settings across all command types. This section covers the standard properties, arguments, flags, validations, and other shared configuration options.

## Command Types

The core command types are `parent`, `exec`, `form`, `scaffold` and `ccm_manifest`. Additional types can be registered through the plugin system.

Most commands are made up of a generic set of options and then have one or more added in addition to specialise them.

### Common properties reference

Most commands include a standard set of fields - those that do not or have special restrictions will mention in the docs.

The following example produces this command:

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

The definition consists of a `commands` member that has these properties:

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

    # Help shown in output of 'example help say' or 'example say --help` (required)
    description: |
      Says something using the cowsay command

      The command called defaults to cowsay but can be
      configured using the Cowsay configuration item

    # Selects the kind of command, see below (required)
    type: exec # or any other known type

    # Optionally allows running 'example say hello' or 'example s hello' (optional)
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

The initial options define the application followed by commands. All the top settings are required except `help_template`, its value may be one of `compact`, `long`, `short` or `default`. When not set it defaults to `default`. Each help format presents information differently (requires version 0.0.9).

A banner can be emitted before invoking the commands in an exec, providing a warning or extra information to users before running a command. For example, a banner may warn that a config override is in use:

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

Cheat Sheet style help is supported, see the [dedicated guide](../cheats/) about that.

#### Arguments

An `argument` is a positional input to a command. `example say hello`, when the command is `say` the `hello` would be the first argument.

Arguments can have many options, the table below detail them and the version that added them.

| Option        | Description                                                                                                                             | Required | Version |
|---------------|-----------------------------------------------------------------------------------------------------------------------------------------|----------|---------|
| `name`        | A unique name for each argument                                                                                                         | yes      |         |
| `description` | A description for this argument, typically 1 line                                                                                       | yes      |         |
| `required`    | Indicates that a value for this argument must be set, which includes being set from default                                             |          |         |
| `enum`        | An array of valid values, if set the flag must be one of these values                                                                   |          | 0.0.4   |
| `default`     | Sets a default value when not passed, will satisfy enums and required. For bools must be `true` or `false`                              |          | 0.0.4   |
| `validate`    | An [expr](https://expr-lang.org) based validation expression, see [Argument and Flag Validations](#argument-and-flag-validations) below |          | 0.8.0   |


#### Flags

A `flag` is a option passed to the application using something like `--flag`, typically these are used for optional inputs. Flags can have many options, the table below detail them and the version that added them.

| Option        | Description                                                                                                                             | Required | Version |
|---------------|-----------------------------------------------------------------------------------------------------------------------------------------|----------|---------|
| `name`        | A unique name for each flag                                                                                                             | yes      |         |
| `description` | A description for this flag, typically 1 line                                                                                           | yes      |         |
| `required`    | Indicates that a value for this flag must be set, which includes being set from default                                                 |          |         |
| `placeholder` | Will show this text in the help output like `--cowfile=FILE`                                                                            |          |         |
| `enum`        | An array of valid values, if set the flag must be one of these values                                                                   |          | 0.0.4   |
| `default`     | Sets a default value when not passed, will satisfy enums and required. For bools must be `true` or `false`                              |          | 0.0.4   |
| `bool`        | Indicates that the flag is a boolean (see below)                                                                                        |          | 0.1.1   |
| `env`         | Will load the value from an environment variable if set, passing the flag specifically wins, then the env, then default                 |          | 0.1.2   |
| `short`       | A single character that can be used instead of the `name` to access this flag. ie. `--cowfile` might also be `-F`                       |          | 0.1.2   |
| `validate`    | An [expr](https://expr-lang.org) based validation expression, see [Argument and Flag Validations](#argument-and-flag-validations) below |          | 0.8.0   |

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

The `--force` flag is used to influence the command. Booleans with their default set to `true` or `"true"` will add a `--no-flag-name` option to negate it. Booleans without a `true` default do not get a negation flag.

#### Argument and Flag Validations

Input provided to commands may need validation. For example, when passing commands
to shell scripts, care must be taken to avoid [Shell Injection](https://en.wikipedia.org/wiki/Code_injection#Shell_injection).

Custom validators on Arguments and Flags are supported using the [Expr Language](https://expr.medv.io/docs/Language-Definition).

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
The standard `expr` language grammar is supported - it has a large number of functions that can assist
validation needs. A few extra functions are added that make sense for operations teams.

In each case accessing `value` would be the value passed from the user.

| Function             | Description                                                   |
|----------------------|---------------------------------------------------------------|
| `isIP(value)`        | Checks if `value` is an IPv4 or IPv6 address                  |
| `isIPv4(value)`      | Checks if `value` is an IPv4 address                          |
| `isIPv6(value)`      | Checks if `value` is an IPv6 address                          |
| `isInt(value)`       | Checks if `value` is an Integer                               |
| `isFloat(value)`     | Checks if `value` is a Float                                  |
| `isShellSafe(value)` | Checks if `value` is attempting to to do shell escape attacks |

### Confirmations

Commands can prompt for confirmation before performing an action:

```yaml
  - name: delete
    description: Delete the data
    type: exec
    confirm_prompt: "Really?"
    command: rm -rf /nonexisting
```

Before running the command the user will be prompted to confirm the action. Since version `0.2.0` an option is
added to the CLI allowing the prompt to be skipped using `--no-prompt`.

## Including other definitions

Since version 0.10.0 an entire definition can be included from another file or just the commands in a parent.

```yaml
name: include
description: An include based app
version: 0.2.2
author: another@example.net

include_file: sample-app.yaml
```

This includes the entire application from another file but overrides the name, description, version and author.

A specific `parent` can load all its commands from a file:

```yaml
  - name: include
    type: parent
    include_file: go.yaml
```

In this case the go.yaml would be the full `parent` definition.

