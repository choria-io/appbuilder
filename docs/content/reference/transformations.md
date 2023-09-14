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
  jq:
    query: .description
```

The `query` supports [Templating](../templating). 

Since version `0.5.0` an optional `yaml_input` boolean can be set true to allow YAML input to be processed using JQ.

## To JSON Transform

The `to_json` transform can convert YAML or JSON format input into JSON format output. By default the output JSON will be compact unindented JSON, prefix and indent strings can be configured.

```yaml
# unindented JSON output
transform:
  to_json: {}
```

```yaml
# Indented JSON output with a custom prefix
transform:
  to_json:
    indent: "  "
    prefix: "  "
```

## To YAML Transform

The `to_yaml` transform can convert JSON format input into YAML format output.

```yaml
transform:
  to_yaml: {}
```

The `to_yaml` transform has no options.

## Bar Graph Transform

This transform takes a JSON document like `{"x": 1, "y": 2}` as input and renders bars for the values.

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
| `caption`   | The caption to show above the graph, supports [Templating](../templating)  |

## Templates

The `template` transform uses Golang templates and the [sprig](http://masterminds.github.io/sprig/) functions
to facilitate creation of text output using a template language.

```yaml
name: template
type: exec
description: Demonstrates template processing of JSON input
command: |
  echo '{"name": "James", "surname":"Bond"}'
transform:
  template:
    contents: |
      Hello {{ .Input.name }} {{ .Input.surname | swapcase }}
```

| Option     | Description                                                                                   |
|------------|-----------------------------------------------------------------------------------------------|
| `contents` | The body of the template embedded in the application yaml file                                |
| `source`   | The file name holding the template, the file name is parsed using [Templating](../templating) |

## Writing to a file

Data entering a the `write_file` transform is written to disk and also returned, but optionally a message can be
returned.

```yaml
name: template
type: exec
description: Demonstrates template processing of JSON input
command: |
  echo '{"name": "James", "surname":"Bond"}'
transform:
  pipeline:
    - write_file:
        file: /tmp/name.txt
        replace: true

    - template:
      contents: |
        Hello {{ .Input.name }} {{ .Input.surname | swapcase }}
```

Above the `/tmp/name.txt` would hold the initial JSON data.

If the `write_file` is the only transform or in a pipeline like here the data received is simply passed on to the next 
step, this can be annoying when writing large files as they will be dumped to the screen.

```yaml
transform:
  - write_file:
    file: /tmp/report.txt
    replace: true
    message: Wrote {{.IBytes}} to {{.Target}}
```

In this case the message `Wrote 1.8 KiB to /tmp/report.txt` would be printed. You can use `.Bytes`, `.IBytes`, `.Target` and `.Contents` in the `message`.

| Option    | Description                                                                  |
|-----------|------------------------------------------------------------------------------|
| `file`    | The file to write, the file name is parsed using [Templating](../templating) |
| `message` | A message to emit from the transform instead of the contents received by it  |
| `replace` | Set to `true` to always overwrite the file                                   |

## Row orientated Reports

These reports allow you to produce text reports for data found in JSON files.  It reports on Array data and produce 
paginated reports with optional headers and footers.

```yaml
name: report
type: exec
description: Demonstrates using a report writer transform
command: curl -s https://api.github.com/repos/choria-io/appbuilder/releases/latest
transform:
  report:
    name: Asset Report
    initial_query: assets
    header: |+
      @|||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||
      data.name
      --------------------------------------------------------------------------------

    body: |
      Name: @<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<< Size: @B###### Downloads: @##
      row.name, row.size,              row.download_count
    footer: |+2
      
                                                                  ====================
                                                                  Total Downloads: @##
      report.summary.download_count
```

Here we fetch the latest release information from GitHub and produce a report with header,
footer and body. Since the JSON data from GitHub is a object we use the `assets` GJSON 
query to find the rows of data to report on.

See the [goform](https://github.com/choria-io/goform) project for a full reference to
the formatting language.

```nohighlight
                                 Release 0.2.1                                  
--------------------------------------------------------------------------------

Name: appbuilder-0.2.1-amd64.deb                   Size: 2.3 MiB  Downloads: 2  
Name: appbuilder-0.2.1-arm6.deb                    Size: 2.2 MiB  Downloads: 0  
Name: appbuilder-0.2.1-arm6.rpm                    Size: 2.1 MiB  Downloads: 0  
....
Name: appbuilder-0.2.1-windows-arm7.zip            Size: 2.1 MiB  Downloads: 0  
Name: appbuilder-0.2.1-x86_64.rpm                  Size: 2.2 MiB  Downloads: 2    

                                                            ====================
                                                            Total Downloads: 20 
```

| Option          | Description                                                                                                                                                |
|-----------------|------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `name`          | The name of the report, parsed using [Templating](../templating)                                                                                           |
| `header`        | The report header                                                                                                                                          |
| `body`          | The report body                                                                                                                                            |
| `footer`        | The report footer                                                                                                                                          |
| `rows_per_page` | How many rows to print per page, pages each have `header` and `footer`                                                                                     |
| `initial_query` | The initial GJSON query to use to find the row orientated data to report                                                                                   |
| `source_file`   | A file holding the report rather than inline, `name`, `header`, `body` and `footer` are read from here. File name parsed using [Templating](../templating) |

## Scaffold

The `scaffold` transform takes JSON data and can generate multiple files using that output.

This is essentially the [Scaffold Command](../scaffold/) in transform form, we suggest you read the Command 
documentation for full details on the underlying feature.  Here we'll just cover what makes the transform unique.

{{% notice secondary "Version Hint" code-branch %}}
This was added in version 0.9.0
{{% /notice %}}

| Option            | Description                                  |
|-------------------|----------------------------------------------|
| `target`          | The directory to write the data into         |
| `post`            | Post processing directives                   |
| `skip_empty`      | Skips files that would be empty when written |
| `left_delimiter`  | Custom template delimiter                    |
| `right_delimiter` | Custom template delimiter                    |

These settings all correspond to the same ones in the command so we won't cover them in full detail here.

The `scaffold` transform returns the input JSON on its output.


## Pipelines

We've seen a few example transform pipelines above, like this one here:

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

