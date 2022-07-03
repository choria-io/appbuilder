+++
title = "Cheat Sheets"
toc = true
weight = 100
+++

While output from `--help` can be useful, many people just don't read it or understand the particular format and syntax
shown.  Instead, a quick cheat sheet style help can often be more helpful.

There is a great utility called [cheat](https://github.com/cheat/cheat) that solves this problem in a generic manner, 
by allowing searching, indexing and rendering of cheat sheets in your terminal.

```nohighlight
$ cheat tar
# To extract an uncompressed archive:
tar -xvf /path/to/foo.tar

# To extract a .tar in specified Directory:
tar -xvf /path/to/foo.tar -C /path/to/destination/
```

We like this format and want to make it available to your App Builder apps, since `0.0.7` it is possible to add cheat
sheets to your application, access them without needing to install the `cheat` command but also integrate them with that
command if you choose.

Cheats are grouped by label, so while your application might have `natsctl report jetstream` the cheats are only 1 level
deep and does not need to match the names of commands.

## Example

Let's see an example, we'll update the example from the quick start to have cheats:

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

When we run it we see:

```nohighlight
usage: demo [<flags>] <command> [<args> ...]

Demo application for Choria App Builder

Contact: https://github.com/choria-io/appbuilder

Use 'demo cheat' to access cheat sheet style help

Commands:
  say <message>
....
```

Since we added 2 cheats just running `demo cheat` will show a list:

```nohighlight
$ demo cheat
Available Cheats:

    demo
    say
```

And we can access the cheat sheet directly:

```nohighlight
$ demo cheat demo
# To say something using a cow
demo say hello

# To think something using a cow
demo think hello
```

## Integrate with `cheat`

The [cheat](https://github.com/cheat/cheat) is great, and I really suggest you check it out, if you have it installed
you can export all the cheats from your builder app into it:

```nohighlight
$ demo cheat --save /home/rip/.config/cheat/cheatsheets/personal/demo
Saved cheat to /home/rip/.config/cheat/cheatsheets/personal/demo/demo
Saved cheat to /home/rip/.config/cheat/cheatsheets/personal/demo/say
```

With this done you can simply do `cheat demo/say`, or find all the cheats tagged `mycorp` which is one of the tags
we added to ours:

```nohighlight
$ cheat -l -t mycorp
title:    file:                                                  tags:
demo/demo /home/rip/.config/cheat/cheatsheets/personal/demo/demo cows,mycorp,personal
demo/say  /home/rip/.config/cheat/cheatsheets/personal/demo/say  cows,mycorp,personal
```
