+++
title = "Configuration"
description = "per-application configuration file support"
weight = 90
toc = true
+++

Items such as passwords, tokens, custom applications, and paths can be supplied via a per-application configuration file.

This file is stored in `example-cfg.yaml` in the standard [file locations](../file-locations/).

The file can contain any valid YAML, for example:

```yaml
# /etc/appbuilder/demo-cfg.yaml
Cowsay: animalsay
```

This can then be used in [templates](../templating/). When a configuration item is required, the `require` function should be used:

```yaml
command: |
   slack-notify --token "{{.Config.slack.token | require "slack token not set" }}"
```
