+++
title = "Configuration"
weight = 30
toc = true
+++

To support supplying items like passwords, tokens, custom applications or paths we support loading a per-application configuration file.

This file is stored in `example-cfg.yaml` in the standard [file locations](../file-locations/).

It's any valid YAML file, for example:

```yaml
# /etc/appbuilder/demo-cfg.yaml
Cowsay: animalsay
```

This can then we used in [templates](../templating/). If a configuration item is required I suggest always using it with the `require` function:

```yaml
command: |
   slack-notify --token "{{.Config.slack.token | require "slack token not set" }}"
```
