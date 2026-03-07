+++
title = "Cheat Sheets"
description = "cheat sheet style help for App Builder applications"
toc = true
weight = 100
+++

While output from `--help` can be useful, many people do not read it or understand the particular format and syntax
shown.  Instead, a quick cheat sheet style help can often be more helpful.

The [cheat](https://github.com/cheat/cheat) utility solves this problem in a generic manner,
by allowing searching, indexing and rendering of cheat sheets in the terminal.

```nohighlight
$ cheat tar
# To extract an uncompressed archive:
tar -xvf /path/to/foo.tar

# To extract a .tar in specified Directory:
tar -xvf /path/to/foo.tar -C /path/to/destination/
```

This format is well suited to App Builder applications. Since `0.0.7` it is possible to add cheat
sheets to an application, access them without needing to install the `cheat` command, and also integrate them with that
command if desired.

Cheats are grouped by label, so while your application might have `natsctl report jetstream` the cheats are only 1 level
deep and does not need to match the names of commands.

## Example

The following example updates the quick start application to include cheats:

```yaml
name: demo
description: Demo application for Choria App Builder
author: https://github.com/choria-io/appbuilder

cheat:
  tags:
    - mycorp
    - cows
  label: demo # this would be the default if not given
  cheat: |
    # To say something using a cow
    demo say hello

    # To think something using a cow
    demo think hello

commands:
  - name: say
    description: Say something using the configured command
    type: exec
    cheat:
      cheat: |
        # This command can be configured using the Cowsay configuration
        Cowsay: /usr/bin/animalsay
    command: |
      {{ default .Config.Cowsay "cowsay" }} {{ .Arguments.message | escape }}
    arguments:
      - name: message
        description: The message to send to the terminal
        required: true
```

Running the application produces:

```nohighlight
usage: demo [<flags>] <command> [<args> ...]

Demo application for Choria App Builder

Contact: https://github.com/choria-io/appbuilder

Use 'demo cheat' to access cheat sheet style help

Commands:
  say <message>
....
```

Since 2 cheats were added, running `demo cheat` shows a list:

```nohighlight
$ demo cheat
Available Cheats:

    demo
    say
```

The cheat sheet is accessible directly:

```nohighlight
$ demo cheat demo
# To say something using a cow
demo say hello

# To think something using a cow
demo think hello
```

## Integrate with `cheat`

The [cheat](https://github.com/cheat/cheat) utility is worth investigating. With it installed,
all cheats from an App Builder application can be exported into it:

```nohighlight
$ demo cheat --save /home/rip/.config/cheat/cheatsheets/personal/demo
Saved cheat to /home/rip/.config/cheat/cheatsheets/personal/demo/demo
Saved cheat to /home/rip/.config/cheat/cheatsheets/personal/demo/say
```

With this done, `cheat demo/say` retrieves the saved cheat, or all cheats tagged `mycorp` (one of the tags
added above) can be listed:

```nohighlight
$ cheat -l -t mycorp
title:    file:                                                  tags:
demo/demo /home/rip/.config/cheat/cheatsheets/personal/demo/demo cows,mycorp,personal
demo/say  /home/rip/.config/cheat/cheatsheets/personal/demo/say  cows,mycorp,personal
```
