+++
title = "Exec Command Type"
weight = 30
toc = true
+++

Use the `exec` command to execute commands found in your shell, and, optionally format their output through JQ.

The `exec` command supports [data transformations](../transformations).

## Running commands

An exec runs a command, it is identical to the [generic example](../common-settings/) shown earlier and accepts flags, arguments and sub commands.  The only difference is that it adds a `command`, `environment` (since `0.0.3`) and `transform` (since `0.0.5`) items.

Below the example that runs cowsay integrated with [configuration](Configuration):

```yaml
name: say
description: Says something using the cowsay command
type: exec

environment:
  - "MESSAGE={{ .Arguments.message}}"

command: |
      {{ default .Config.Cowsay "cowsay" }} "{{ .Arguments.message | escape }}"

arguments:
   - name: message
     description: The message to display
     required: true
```

The `command` is how the shell command is specified and we show some [templating](../templating).  This will read the `.Config` hash for a value `Cowsay` if it does not exist it will default to `"cowsay"`. We also see how we can reference the `.Arguments` to access the value supplied by the user, we escape it for shell safety.

We also show how to set environment variables using `environment`, this too will be templated.

Since version `0.0.9` setting environment variable `BUILDER_DRY_RUN` to any value will enable debug logging, log the command and terminate without calling your command.

## Shell scripts

A shell script can be added directly to your app, setting `shell` will use that command to run the script, if not set it will use `$SHELL`, `/bin/bash` or `/bin/sh` which ever is found first.

The script is parsed through [templating](../templating).

```yaml
name: script
description: A shell script
type: exec
shell: /bin/zsh
script: |
  for i in {1..5}
  do
    echo "hello world"
  done
```
