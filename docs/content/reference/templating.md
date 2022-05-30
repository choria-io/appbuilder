+++
title = "Templating"
toc = true
weight = 25
+++

Templates allow you to interpolate values from Flags, Arguments and Configuration into some aspects of commands.

For example the [exec](../exec) command type allows you to use templates to put arguments into the command being run.

We use the Go template language at the moment, it's not the best we might look at something else later.

Only some fields are parsed through templates, the documentation for each [command type](../common-settings) will call out what is supported.

## Reference

An example template use was shown in the [exec](../exec) documentation:

```yaml
command: |
      {{ default .Config.Cowsay "cowsay" }} "{{ .Arguments.message | escape }}"
```

Here we have examples of accessing the `.Config` and `.Arguments` structures and using some functions.

### Available Data

| Key          | Description                                                |
|--------------|------------------------------------------------------------|
| `.Config`    | Data stored in the configuration file for this application |
| `.Arguments` | Data supplied by users using command arguments             |
| `.Flags`     | Data supplied by users using command flags                 |

### Available Functions

| Function    | Description                                                                            | Example                                 |
|-------------|----------------------------------------------------------------------------------------|-----------------------------------------|
| `require`   | Asserts that some data is available, errors with an optional custom message on failure | `{{ .Config.Password \| require "Password not set in the configuration" }}`|
| `escape`    | Escapes a string for use in shell arguments                                            | `{{ .Arguments.message \| escape }}`|
| `read_file` | Reads a file                                                                           | `{{ read_file .Arguments.file }}`       |
| `default`   | Checks a value, if its not supplied uses a default                                     | `{{ default .Config.Cowsay "cowsay" }}` |
