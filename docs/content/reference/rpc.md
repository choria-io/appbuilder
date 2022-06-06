+++
title = "Choria RPC Command Type"
weight = 50
toc = true
+++

The RPC command interact with the Choria RPC system used execute actions on remote nodes.

Since this is built into Choria it will simply use your Choria Client configuration for the user executing the command
to find the Choria Brokers and more. It supports the usual override methods such as creating a `choria.conf` file in
your project working directory. No connection properties are required or supported.

Before using this command type I suggest reading about [Choria Concepts](https://choria.io/docs/concepts/).

{{% notice secondary "Version Hint" code-branch %}}
This feature is only available when hosting App Builder applications within the Choria Server version 0.26.0 or newer
{{% /notice %}}

## Overview

This command supports all the standard properties like Arguments, Flags, Banners and more, it also incorporates the
discovery features of the [Discover Command Type](../discover/) in order to address nodes.

Below a simple RPC request.

```yaml
name: stop
description: Stops the Service gracefully
type: rpc

request:
  agent: service
  action: stop
  inputs:
    service: httpd
```

This will look and behave exactly like `choria req service stop service=httpd`.

## Adjusting CLI Behavior

A number of settings exist to adjust the behavior or add flags to the CLI at runtime.  Generally you can either allow users
to supply values sugh as `--json`, or force the output to be JSON but you cannot allow both at present:

| Setting                    | Description                                                                                  |
|----------------------------|----------------------------------------------------------------------------------------------|
| `std_filters`              | Enables standard filter flags like `-C`, `-W` and more                                       |
| `output_format`            | Forces a specific output format, one of `senders`, `json` or `table`                         |
| `output_format_flags`      | Enables `--senders`, `--json` and `--table` options, cannot be set with `output_format`      |
| `display`                  | Supplies a setting to the typical `--display` option, one of `ok`, `failed`, `all` or `none` |
| `display_flag`             | Enables the `--display` flag on the CLI, cannot be used with `display`                       |
| `batch_flags`              | Adds the `--batch` and `--batch-sleep` flags                                                 |
| `batch`, `batch_sleep`     | Supplies values for `--batch` and `--batch-sleep`, cannot be used with `batch_flags`         |
| `no_progress`              | Disables the progress bar`                                                                   |
| `all_nodes_confirm_prompt` | A confirmation prompt shown when an empty filter is used                                     |

## Request Parameters

Every RPC request needs `request` specified that must have at least `agent` and `action` set.

Inputs are allowed as a string hash - equivalent to how one would type inputs on the `choria req` CLI.

It also accepts a `filter` option that is the same as that in the [discover command](../discover/).

```yaml
name: stop
description: Stops the Service gracefully
type: rpc

request:
  agent: service
  action: stop
  inputs:
    service: httpd
  filter:
    classes:
      - roles::apache
```

## Filtering Replies

Results can be filtered using a [result filter](https://choria.io/docs/concepts/cli/#filtering-results), this allows you
to exclude/include specific replies before rendering the results.

Here's an example that will find all Choria Servers with a few flags to match versions, it invokes the `rpcutil#daemon_states`
action and then filters results matching a query. Only the matching node names are shown.

```yaml
name: busy
description: Find Choria Agents matching certain versions
type: rpc

# list only the names
output_format: senders

flags:
  - name: ne
    description: Finds nodes with version not equal to the given
    placeholder: VERSION
    reply_filter: ok() && semver(data("version", "!= {{.Flags.ne}}
  - name: eq
    description: Finds nodes with version equal to the given
    placeholder: VERSION
    reply_filter: ok() && semver(data("version", "== {{.Flags.eq}}

request:
  agent: rpcutil
  action: daemon_stats
```

## Transforming Results

Results can be transformed using GOJQ syntax, here's one that gets the state of a particular autonomous agent:

```yaml
name: state
description: Obtain the state of the service operator
type: rpc
transform:
  query: |
    .replies | .[] | select(.statuscode==0) | .sender + ": " + .data.state
request:
  agent: choria_util
  action: machine_state
  inputs:
    name: nats
```

When run it will just show lines like:

```nohighlight
n1-lon: RUN
n3-lon: RUN
n2-lon: RUN
```
