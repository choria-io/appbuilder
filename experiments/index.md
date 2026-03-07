# Experiments

This section documents experimental features that are not yet part of the supported feature set.

## CCM Manifest Command Type

The `ccm_manifest` command type invokes a [Choria Configuration Manager](https://choria-io.github.io/ccm/) manifest directly. All flags and arguments are set as manifest data.

```yaml
name: demo
description: Demo application for Choria App Builder
author: https://github.com/choria-io/appbuilder
commands:
  - name: docker
    description: Install docker using CCM
    type: ccm_manifest
    flags:
      - name: version
        description: Version to install
        required: true
        default: latest
    manifest: obj://CCM/simple.tar.gz
    nats_context: CCM
    render_summary: true
    no_render_messages: false
```

The command accepts the standard [common settings](../reference/common-settings/) including flags, arguments and sub commands. It also supports [data transformations](../reference/transformations/) over the manifest session summary data.

| Option                         | Description                                                                         |
|--------------------------------|-------------------------------------------------------------------------------------|
| `manifest`                     | The URL to the manifest to execute, supports [templating](../reference/templating/) |
| `nats_context`                 | The NATS context to use when invoking the manifest, defaults to `CCM`               |
| `render_summary` (boolean)     | When true displays the summary statistics as text on STDOUT                         |
| `no_render_messages` (boolean) | When true suppresses the pre- or post-message defined in the manifest               |

## Choria Configuration Manager Transform

The [Choria Configuration Manager](https://choria-io.github.io/ccm/) is a configuration management tool that is part of the Choria project.

CCM manifests take data and facts as input. The `ccm_manifest` transform executes a manifest with custom data.

This combines well with the form-based wizards to collect configuration interactively. All flags and arguments are also set as data.

```yaml
name: demo
description: Demo application for Choria App Builder
author: https://github.com/choria-io/appbuilder
commands:
  - name: docker
    type: form
    properties:
      - name: version
        description: Version to install
        required: true
        default: latest

    transform:
      ccm_manifest:
        manifest: obj://CCM/simple.tar.gz
        nats_context: CCM
        render_summary: true
        no_render_messages: false
```

This sets `version` in the data supplied to the manifest and executes it. Setting `render_summary` renders the summary to STDOUT rather than returning it as JSON. Setting `no_render_messages` suppresses the pre- and post-messages in the manifest.

A top-level command type is also available:

```yaml
name: demo
description: Demo application for Choria App Builder
author: https://github.com/choria-io/appbuilder
commands:
  - name: docker
    type: ccm_manifest
    flags:
      - name: version
        description: Version to install
        required: true
        default: latest
    manifest: obj://CCM/simple.tar.gz
    nats_context: CCM
    render_summary: true
    no_render_messages: false
```

Here a flag supplies the version instead of a form.