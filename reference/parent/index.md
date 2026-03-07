# Parent Command

A parent is a placeholder. In a command like `example deploy status` and `example deploy upgrade`, the `deploy` is a `parent`. It exists to group related commands and takes no action on its own.

It requires the `name`, `description`, `type` and `commands` and the optional `aliases` and `include_file`.

It does not accept `flags`, `arguments`, `confirm_prompt` or `banner`.

```yaml
name: deploy
description: Manage deployment of the system
type: parent

# Commands are required for the parent type and should have more than 1
commands: []
```

## Including commands from a file

The `include_file` option allows loading the parent command definition from an external YAML file. The `name` set in the parent definition is preserved while other settings are loaded from the file.

```yaml
name: deploy
description: Manage deployment of the system
type: parent
include_file: deploy_commands.yaml
```
