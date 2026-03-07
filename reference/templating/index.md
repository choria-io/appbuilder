# Templating

Templates allow interpolation of values from Flags, Arguments, and Configuration into certain aspects of commands.

For example, the [exec](../exec) command type supports templates for placing arguments into the command being run.

The Go template language is used for all template processing.

Only some fields are parsed through templates, the documentation for each [command type](../common-settings) will call out what is supported.

## Reference

An example template use was shown in the [exec](../exec) documentation:

```yaml
command: |
      {{ default .Config.Cowsay "cowsay" }} "{{ .Arguments.message | escape }}"
```

This example demonstrates accessing the `.Config` and `.Arguments` structures and using some functions.

### Available Data

| Key          | Description                                                                  |
|--------------|------------------------------------------------------------------------------|
| `.Config`    | Data stored in the configuration file for this application                   |
| `.Arguments` | Data supplied by users using command arguments                               |
| `.Flags`     | Data supplied by users using command flags                                   |
| `.Input`     | Parsed JSON input from a previous step, available in transform contexts only |

### Available Functions

| Function         | Description                                                                            | Example                                 |
|------------------|----------------------------------------------------------------------------------------|-----------------------------------------|
| `require`        | Asserts that some data is available, errors with the given message on failure or a default message when empty  | `{{ require .Config.Password "Password not set in the configuration" }}`|
| `escape`         | Escapes a string for use in shell arguments                                            | `{{ escape .Arguments.message }}`|
| `read_file`      | Reads a file                                                                           | `{{ read_file .Arguments.file }}`       |
| `default`        | Checks a value, if its not supplied uses a default                                     | `{{ default .Config.Cowsay "cowsay" }}` |
| `env`            | Reads an environment variable                                                          | `{{ env "HOME" }}`                      |
| `UserWorkingDir` | Returns the directory the user is in when running the command                           | `{{ UserWorkingDir }}`                  |
| `AppDir`         | Returns the directory the application definition is in                                 | `{{ AppDir }}`                          |
| `TaskDir`        | Alias for `AppDir`                                                                     | `{{ TaskDir }}`                         |

In addition to the above, the [Sprig](http://masterminds.github.io/sprig/) functions library is available in most template contexts including commands, scaffolds and transforms.
