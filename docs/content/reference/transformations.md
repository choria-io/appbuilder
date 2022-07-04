+++
title = "Transformations"
toc = true
weight = 80
+++

Transformations is like a shell pipe defined in App Builder.  We have a number of transformations and using them is entirely optional - often a shell pipe would be much better.

The reason for adding transformations like `jq` to App Builder itself is to have it function in places where that 3rd party dependency is not met.  Rather than require everyone to install JQ - and handle that dependency, we just add a JQ dialect directly to App Builder.

A basic example of transformations can be seen here:

```yaml
name: ghd
description: Gets the description of a Github Repo
type: exec
command: |
  curl -s -H "Accept: application/vnd.github.v3+json" https://api.github.com/repos/choria-io/appbuilder

transform:
  jq:
    query: .description
```

Here we call out a REST API that returns JSON payload using `curl` and then extract the `description` field from the result using a JQ transform 

```nohighlight
$ demo ghd
Tool to create friendly wrapping command lines over operations tools
```

Not every command supports transforms, so the individual command documentation will call out it out.

## JQ Transform

The `jq` transform uses a dialect of JQ called [GoJQ](https://github.com/itchyny/gojq), most of your JQ knowledge is transferable with only slight changes/additions.  This is probably the most used transform so there's a little short cut to make using it a bit easier:

```yaml
transform:
  query: .description
```

Is the same as (added in `0.2.0`):

```yaml
transform:
  jq:
    query: .description
```

The `query` parameter is all that is supported. The `query` supports [Templating](../templating).

## Bar Graph Transform

This transform takes a JSON document like `{"x": 1, "y": 2}` as input and renders bars for the values.

{{% notice secondary "Version Hint" code-branch %}}
Added in version 0.2.0
{{% /notice %}}

Here is an example that draws the sizes of the assets of the latest release:

```yaml
name: bargraph
description: Draws an ASCII bar graph
type: exec
transform:
  pipeline:
    - jq:
        query: |
          .assets|map({(.name): .size})|reduce .[] as $a ({}; . + $a)

    - bar_graph:
        caption: "Release asset sizes"
        bytes: true
script: |
  curl -s https://api.github.com/repos/choria-io/appbuilder/releases/latest
```

This uses a pipeline (see below) to transform a GitHub API request into a hash and then a `bar_graph` to render it:

```nohighlight
$ demo bargraph
Latest release asset sizes

    appbuilder-0.1.1-windows-arm64.zip: ▏ (2.0 MiB)
            appbuilder-0.1.1-arm64.rpm: ██ (2.0 MiB)
   appbuilder-0.1.1-linux-arm64.tar.gz: ███ (2.0 MiB)
            appbuilder-0.1.1-arm64.deb: ███ (2.0 MiB)
     appbuilder-0.1.1-windows-arm7.zip: █████████████ (2.1 MiB)
     appbuilder-0.1.1-windows-arm6.zip: ██████████████ (2.1 MiB)
             appbuilder-0.1.1-arm7.rpm: ██████████████ (2.1 MiB)
             appbuilder-0.1.1-arm7.deb: ███████████████ (2.1 MiB)
    appbuilder-0.1.1-linux-arm7.tar.gz: ███████████████ (2.1 MiB)
             appbuilder-0.1.1-arm6.rpm: ███████████████ (2.1 MiB)
             appbuilder-0.1.1-arm6.deb: ███████████████ (2.1 MiB)
    appbuilder-0.1.1-linux-arm6.tar.gz: ███████████████ (2.1 MiB)
    appbuilder-0.1.1-windows-amd64.zip: ███████████████████████ (2.2 MiB)
  appbuilder-0.1.1-darwin-arm64.tar.gz: █████████████████████████ (2.2 MiB)
           appbuilder-0.1.1-x86_64.rpm: ███████████████████████████ (2.2 MiB)
   appbuilder-0.1.1-linux-amd64.tar.gz: ███████████████████████████ (2.2 MiB)
            appbuilder-0.1.1-amd64.deb: ███████████████████████████ (2.2 MiB)
  appbuilder-0.1.1-darwin-amd64.tar.gz: ████████████████████████████████████████ (2.3 MiB)
```

The transform supports a few options, all are optional:

| Option    | Description                                                                                           |
|-----------|-------------------------------------------------------------------------------------------------------|
| `width`   | The width of the bar, defaults to 40                                                                  |
| `caption` | The cpation to show above the graph, supports [Templating](../templating)                             |
| `bytes`   | When set to true indicates that the numbers rendered after the bars will be bytes like in the example |

## Line Graph

This transform takes input of floats per line or a JSON document (array of floats) and turns it into a line graph.

{{% notice secondary "Version Hint" code-branch %}}
Added in version 0.2.0
{{% /notice %}}

Here we find the hourly forecast for our location and graph it:

```yaml
description: Hourly weather forecast
type: exec
transform:
  pipeline:
    - jq:
        query: |
          .weather[0].hourly|.[]|.FeelsLikeC
    - line_graph:
        width: 40
        height: 10
        caption: Hourly weather forecast (C)
command: |
  curl -s wttr.in/?format=j1
```

When run this produces:

```nohighlight
$ demo linegraph
 30.00 ┤                      ╭─────╮
 29.90 ┤                     ╭╯     │
 29.80 ┤                    ╭╯      ╰╮
 29.70 ┤                    │        │
 29.60 ┤                   ╭╯        ╰╮
 29.50 ┤                   │          │
 29.40 ┤                  ╭╯          ╰╮
 29.30 ┤                  │            ╰╮
 29.20 ┤                 ╭╯             │
 29.10 ┤                ╭╯              ╰╮
 29.00 ┼────────────────╯                ╰─────
              Hourly weather forecast (C)
```

The transform supports a few options, all are optional:

| Option      | Description                                                                |
|-------------|----------------------------------------------------------------------------|
| `width`     | The width of the graph, defaults to 40                                     |
| `height`    | The height of the graph, defaults to 20                                    |
| `precision` | The decimal precision to consider and render                               |
| `json`      | When true expects JSON input like `[1,2,3,4]` rather than a float per line |
| `caption`   | The cpation to show above the graph, supports [Templating](../templating)  |

## Pipelines

We've seen a few example transform pipelines above, like this one here:

{{% notice secondary "Version Hint" code-branch %}}
Added in version 0.2.0
{{% /notice %}}

```yaml
type: exec
transform:
  pipeline:
    - jq:
        query: |
          .weather[0].hourly|.[]|.FeelsLikeC
    - line_graph:
        width: 80
        caption: Hourly weather forecast (C)
command: |
  curl -s wttr.in/?format=j1
```

This runs the output of the `curl` command (JSON weather forecast data) through a `jq` transform that produce results like:

```nohighlight
29
29
29
29
30
30
29
29
```

We then feed that data into a `line_graph` and render it, the output from the `jq` transform is used as input to the `line_graph`.

Any failure in the pipeline will terminate processing.
