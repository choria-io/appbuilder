+++
title = "Exec Command Type"
weight = 20
toc = true
+++

Use the `exec` command to execute commands found in your shell, and, optionally format their output through JQ.

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

We also show how to set environment variables using `environment`, this too will be templated. This was added in version `0.0.3`.

## Shell scripts

A shell script can be added directly to your app, setting `shell` will use that command to run the script, if not set it will use `$SHELL`, `/bin/bash` or `/bin/sh` which ever is found first.

{{% notice secondary "Version Hint" code-branch %}}
Added in version 0.0.8
{{% /notice %}}

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

## Transformation using JQ

If you have a command that is known to emit JSON data you can ask `appbuilder` to transform that data using a dialect of JQ called [GoJQ](https://github.com/itchyny/gojq), the resulting data will be printed to STDOUT.

{{% notice secondary "Version Hint" code-branch %}}
Added in version 0.0.5
{{% /notice %}}

```yaml
name: ghd
description: Gets the description of a Github Repo
type: exec
command: |
  curl -H "Accept: application/vnd.github.v3+json"
     https://api.github.com/repos/{{ .Arguments.owner }}/{{ .Arguments.repo }}

transform:
  query: .description

arguments:
  - name: owner
    description: The repo owner
    required: true

  - name: repo
    description: The repo name
    require: true
```

Here we fetch data from the GitHub API and use the internal JQ to transform it by extracting just the one item.

```nohighlight
$ demo ghd choria-io appbuilder
Tool to create friendly wrapping command lines over operations tools
```
