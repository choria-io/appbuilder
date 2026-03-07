+++
title = "App Builder"
description = "declarative CLI application builder"
toc = true
weight = 5
archetype = "home"
+++

The App Builder project builds CLI applications from YAML definitions, encapsulating shell scripts, piped commands, `kubectl` invocations, and other operational tools into a single discoverable command.

Operations teams tend to rely on a large selection of shell scripts and ad-hoc commands in their day-to-day work. This tribal knowledge is a hurdle for new team members. Wiki-based run books that capture these commands do not scale well and do not stay up to date.

App Builder solves this by specifying a model for a CLI application in a YAML file and building custom interfaces on the fly. When the underlying commands change, only the app model needs updating - the CLI interface and any wiki pages that reference it remain stable.

A [video introduction](https://youtu.be/-IUwoXEJK0c) covers the motivation behind App Builder.

## Example

The following example shows App Builder in use:

```nohighlight
$ natsctl --help
usage: natsctl [<flags>] <command> [<args> ...]

NATS Stress Test Cluster Controller

Contact: R.I.Pienaar <rip@devco.net>

Flags:
  --help  Show context-sensitive help (also try --help-long and --help-man).

Commands:
  help [<command>...]
    Show help.

  report servers
    Shows list of NATS servers

  report jetstream
    Shows JetStream status

  service list [<flags>]
    List the servers running the service

  service state [<flags>]
    Obtain the state of the service operator
...
```

In the example above, commands like `natsctl report servers` or `natsctl report jetstream` invoke something like `nats server list --user system --password secret --server nats.example.net:4222`.

The `natsctl service state` command invokes a Choria RPC API in a subset of fleet nodes and passes the result through a JQ query `.replies | .[] | select(.statuscode==0) | .sender + ": " + .data.state` to transform the API output.

Wrapping these in `natsctl report servers` or `natsctl service state` keeps the detail managed separately. Wikis can reference `natsctl report servers`, and if the underlying command ever changes, only the app model needs updating.

All sub-commands include built-in help. Shell completion for `bash` and `zsh` is provided.
