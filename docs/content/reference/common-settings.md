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

Here we show the initial options that define the application followed by commands.  All the top settings are required except `help_template`, it's value may be one of `compact`, `long` or `default`.  When not set it will equal `default`. Experiment with these options to see which help format suits your app best (requires version 0.0.9). 

Since version `0.0.4` if a specific flag or argument has a finite number of options, you can limit it using the `enum` option and we have a `default` option to complement it, here's an example:

```yaml
flags:
  - name: eyes
    description: Control the eyes of the cow
    enum: ["*", "+", "x", "@"]
    default: "+"
```

If any option other than those are supplied an error will be raised. If `--eyes` is not given it will default to `+`.

Since version `0.0.6` you can emit a banner before invoking the commands in an exec, use this to show a warning or extra
information to users before running a command.  Perhaps to warn them that a config override is in use like here:

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

Since version `0.0.7` we support Cheat Sheet style help, see the [dedicated guide](../cheats/) about that.

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

### Boolean Flags

We support boolean flags since version `0.1.1`:

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

