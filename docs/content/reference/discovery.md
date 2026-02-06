+++
title = "Choria Discover Command"
weight = 40
toc = true
+++

The Discover command interact with the Choria Discovery system used to find fleet nodes based on a vast array of
possible queries and data sources.

Since this is built into Choria it will simply use your Choria Client configuration for the user executing the command
to find the Choria Brokers and more. It supports the usual override methods such as creating a `choria.conf` file in
your project working directory. No connection properties are required or supported.

Before using this command type I suggest reading about the [Choria Discovery System](https://choria.io/docs/concepts/discovery/).

{{% notice secondary "Version Hint" code-branch %}}
This feature is only available when hosting App Builder applications within the Choria Server version 0.26.0 or newer
{{% /notice %}}

## Overview

This command supports all the standard properties like Arguments, Flags, Banners and more, below is a simply command
that finds apache servers.

```yaml
name: find
description: Finds all machines tagged as Apache Servers
type: discover

std_filters: true
filter:
  classes:
    - roles::apache
```

When run it will show a list of matching nodes, one per line.  It also accepts the `--json` flag to enable returning a
JSON array of matching nodes.

Since the `std_filters` option is set the command will also accept additional filters in standard Choria format. Flags
like `-C`, `-F`, discovery mode selectors and more. User supplied options will be merged/appended with the ones supplied
in the YAML file. By default, none of the standard Choria flags will be added to the CLI. 

All the filter values, even arrays and objects, support [templating](../templating/).

## Filter Reference

The main tunable here is the filter, below a reference of available options. The examples here are a bit short, I suggest
you read the [Choria Discovery Documentation](https://choria.io/docs/concepts/discovery/) for a thorough understanding.

| Key                         | Description                                                                      | Example                                             |
|-----------------------------|----------------------------------------------------------------------------------|-----------------------------------------------------|
| `collective`                | The collective to target, defaults to main collective                            | `collective: development`                           |
| `facts`                     | List of fact filters as passed to `-F`                                           | `facts: ["country=uk"]`                             |
| `agents`                    | List of agent filters as passed to `-A`                                          | `agents: ["puppet"]`                                |
| `classes`                   | List of Config Management classes to match as passed to `-C`                     | `classes: ["apache"]`                               |
| `identities`                | List of node identities to match as passed to `-I`                               | `identities:["/^web/"]`                             |
| `combined`                  | List of Combined filters as passed to `-W`                                       | `combined:["/^web/","location=uk"]`                 |
| `compound`                  | A single Compound filter as passed to `-S`                                       | `compound: "with('apache') or with('nginx')`        |
| `discovery_method`          | A discovery method to use like `inventory` as passed to `--dm`                   | `discovery_method:"flatfile"`                       |
| `discovery_options`         | A set of discovery options, specific to the `discovery_method` chosen            | `discovery_options: {"file":"/etc/inventory.yaml"}` |
| `discovery_timeout`         | How long discovery can run, in seconds, as passed to `--discovery-timeout`       | `discovery_timeout: 2`                              |
| `dynamic_discovery_timeout` | Enables windowed dynamic timeout rather than a set discovery timeout             | `dynamic_discovery_timeout: true`                   |
| `nodes_file`                | Short cut to use flatfile discovery with a specific file, as passed to `--nodes` | `nodes_file: /etc/fleet.txt`                        |

