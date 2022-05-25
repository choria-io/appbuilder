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

```
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
author: Opertions team <ops@example.net>

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
