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

dir: /tmp

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

Since version `0.9.0` setting `dir` will execute the command in that directory. This setting supports [templating](../templating) and has an sets extra variables `UserWorkingDir` for the directory the user is in before running the command, `AppDir` and `TaskDir` indicating the directory the definition is in.

Setting environment variable `BUILDER_DRY_RUN` to any value will enable debug logging, log the command and terminate without calling your command.

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

## Common helper functions

We provide a basic helper shell script that can be used to echo text to the screen in various ways. To use this you can 
source the script:

{{% notice secondary "Version Hint" code-branch %}}
Added in version 0.6.3
{{% /notice %}}

```yaml
name: script
description: A shell script
type: exec
shell: /bin/zsh
script: |
  set -e

  . "{{ BashHelperPath }}"
  
  ab_announce Hello World
```

This will output:

```nohighlight
>>> Hello World
```

It provides a few functions:

 * `ab_say` prefix the message using a single prefix `>>>`
 * `ab_announce` prefix the message with `>>>` with a line of `>>>` before and after the message
 * `ab_error` prefix the message with `!!!`
 * `ab_panic` prefix the message with `!!!` and exit the script with code 1

The `>>>` can be configured by setting `AB_SAY_PREFIX` and the `!!!` by setting `AB_ERROR_PREFIX` after sourcing the helper.

The output can have time stamps added to the lines by setting `AB_HELPER_TIME_STAMP` shell variable to `T` for time and `D` for time and date prefixes.

## Retrying failed executions

Failing executions can be tried based on a backoff policy, here we configure a maximum of 10 attempts with varying sleep
times that would include randomized jitter.

Scripts can detect if they are running in a retry by inspecting the `BUILDER_TRY` environment variable.

```yaml
name: retry
description: A shell script execution with backoff retries
type: exec
command: ./script.sh
backoff:
  # Maximum amount of retries, required
  max_attempts: 10
  # Maximum sleep time + jitter, optional
  max_sleep: 20s
  # Minimum sleep time + jitter, optional
  min_sleep: 1s
  # Number of steps in the backoff policy, once the max is reached
  # further retries will jitter around max_sleep, optional
  steps: 5
```

Only the `max_attempts` setting is required, `min_sleep` defaults to `500ms` and `max_sleep` defaults to `20s` with steps
defaulting to `max_attempts`.
