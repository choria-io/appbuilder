# Quick Start

This guide covers installing App Builder and creating a first application.

## Installation

The [Releases](https://github.com/choria-io/appbuilder/releases) page provides binaries, RPMs, DEBs, and zip files containing the `appbuilder` command. The binary can be placed anywhere in the system PATH.

A [JSON Schema for the definition](https://choria.io/schemas/appbuilder/v1/application.json) is available for editor integration.

OS X and Linux homebrew packages are available:

```nohighlight
brew tap choria-io/tap
brew install choria-io/tap/appbuilder
```

## Hello World

The following example creates a command that invokes `cowsay` to demonstrate some capabilities of the system.

The command supports `cowsay`, `cowthink`, and an optional configuration override to use `animalsay` instead of `cowsay`.

```nohighlight
demo say "hello world"
demo think "hello world"
```

The following YAML file (`demo-app.yaml`) describes the demo application. The reference sections cover all available options.

```yaml
name: demo
description: Demo application for Choria App Builder
author: https://github.com/choria-io/appbuilder
commands:
  - name: say
    description: Say something using the configured command
    type: exec
    command: |
      {{ default .Config.Cowsay "cowsay" }} {{ .Arguments.message | escape }}
    arguments:
      - name: message
        description: The message to send to the terminal
        required: true

  - name: think
    description: Think something using a cow
    type: exec
    command: |
      cowthink {{ .Arguments.message | escape }}
    arguments:
      - name: message
        description: The message to send to the terminal
        required: true
```

Place this file in either `/etc/appbuilder/demo-app.yaml` or `~/.config/appbuilder/demo-app.yaml` (`~/Library/Application Support/appbuilder/demo-app.yaml` on a Mac).

```nohighlight
$ sudo mkdir -p /etc/appbuilder
$ sudo cp demo-app.yaml /etc/appbuilder/
$ sudo ln -s /usr/local/bin/appbuilder /usr/bin/demo
$ demo say "hello world"
 _____________
< hello world >
 -------------
        \   ^__^
         \  (oo)\_______
            (__)\       )\/\
                ||----w |
                ||     ||

$ demo think "this is pretty cool"
 _____________________
( this is pretty cool )
 ---------------------
        o   ^__^
         o  (oo)\_______
            (__)\       )\/\
                ||----w |
                ||     ||
```

The YAML file above makes the say command configurable. The following demonstrates how that works.

Create the following in `/etc/appbuilder/demo-cfg.yaml`.

```yaml
Cowsay: animalsay
```

Now invoking `demo say` uses `animalsay`:

```nohighlight
$ demo say "hello world"
 _____________
< hello world >
 -------------
 \     ____________
  \    |__________|
      /           /\
     /           /  \
    /___________/___/|
    |          |     |
    |  ==\ /== |     |
    |   O   O  | \ \ |
    |     <    |  \ \|
   /|          |   \ \
  / |  \_____/ |   / /
 / /|          |  / /|
/||\|          | /||\/
    -------------|
        | |    | |
       <__/    \__>
```
