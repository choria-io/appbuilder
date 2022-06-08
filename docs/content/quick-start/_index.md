+++
title = "Quick Start"
toc = true
weight = 10
pre = "<b>1. </b>"
+++

## Installation

Over on our [Releases](https://github.com/choria-io/appbuilder/releases) page you will find binaries, rpms, debs, zip files and more holding the `appbuilder` command. There is just one command and you can place it anywhere in your path.

If your editor supports it there is a [JSON Schema for the definition](https://choria.io/schemas/appbuilder/v1/application.json).

We publish OS X and Linux homebrew packages:

```nohighlight
$ brew tap choria-io/tap
$ brew install choria-io/tap/appbuilder
```

## Hello World

We will make a little command that invokes `cowsay` to demonstrate some of the capabilities of the system.

We want to be able to run this command and it should invoke `cowsay`, `cowthink` or if configured to do so use `animalsay` instead of `cowsay`

```nohighlight
$ demo say "hello world"
$ demo think "hello world"
```

First we have to write a YAML file that describes our demo application, we have reference sections in the wiki for all the options, so this being an introduction, will be short on details.

```
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
$ sudo mkdir -p /etc/appbuilder/demo-app.yaml
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

Finally, if you read the YAML file, you may see we made the system configurable, lets look how that works.

Create the following in `/etc/appbuilder/demo-cfg.yaml`

```yaml
Cowsay: animalsay
```

Now when we invoke `demo say` it will use `animalsay`:

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
