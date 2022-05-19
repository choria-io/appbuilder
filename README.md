![Choria App Builder](https://github.com/choria-io/appbuilder/raw/main/images/logo.png)

## Overview

This is a tool to help operations teams wrap their myriad shell scripts, multi line kubectl invocations, jq commands and
more all in one friendly CLI tool that's easy to use and share.

## Quick Start

We'll show how to make a tiny app that just wraps `echo`:

```nohighlight
usage: mycorp [<flags>] <command> [<args> ...]

A hello world sample Choria App

Contact: R.I.Pienaar <rip@devco.net>

Flags:
  --help  Show context-sensitive help (also try --help-long and --help-man).

Commands:
  help [<command>...]
    Show help.

  demo hello
    Displays the word 'hello'

  demo world
    Displays the word 'world'

  echo <word>
    Simple echo demonstration
```

To make this install the `appbuilder` command in your shell path, then place the below YAML file in `~/.config/choria/builder/mycorp-app.yaml`
or in your current working directory:

```yaml
name: hello
description: A hello world sample Choria App
version: 0.0.1
author: R.I.Pienaar <rip@devco.net>

commands:
    - name: demo
      description: A simple Hello World demonstration
      type: parent
      commands:
        - name: hello
          description: Displays the word 'hello'
          type: exec
          command: echo hello

        - name: world
          description: Displays the word 'world'
          type: exec
          confirm_prompt: Really show the word 'world'
          command: echo world

    - name: echo
      description: Simple echo demonstration
      type: exec
      command: echo {{.Arguments.word}}
      arguments:
      - name: word
        description: The word to show
        required: true
```

Then create a symlink for `mycorp`:

```nohighlight
$ ls -l mycorp
lrwxrwxrwx 1 rip rip  24 May 19 22:02 mycorp -> /home/rip/bin/appbuilder
```

Now simply run `mycorp`.

## Status

This is a brand new project that's under active development, we are extracting the core out of the maain Choria project
into a standalone component.

There is a [video introducing the idea behind it](https://youtu.be/wbu3N63WY7Y).
