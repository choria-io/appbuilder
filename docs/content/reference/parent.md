+++
title = "Parent Command"
toc = true
weight = 20
+++

A parent is a placeholder, you can have a command like `example deploy status` and `example deploy upgrade`, here the `deploy` is a `parent`. It's just there to group related commands and takes no action on it's own.

It requires the the `name`, `description`, `type` and `commands` and the optional `aliases`.

It does not accept `flags`, `arguments` or `confirm_prompt`.

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
