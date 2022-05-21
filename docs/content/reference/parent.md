+++
title = "Parent Command Type"
toc = true
weight = 15
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
