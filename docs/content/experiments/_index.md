+++
title = "Experiments"
toc = true
weight = 40
pre = "<b>4. </b>"
+++

Some features are ongoing experiments and not part of the supported feature set, this section will call them out.

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

| Option               | Description                                                                         |
|----------------------|-------------------------------------------------------------------------------------|
| `manifest`           | The URL to the manifest to execute, supports [templating](../reference/templating/) |
| `nats_context`       | The NATS context to use when invoking the manifest, defaults to `CCM`               |
| `render_summary`     | When true displays the summary statistics as text on STDOUT                         |
| `no_render_messages` | When true will not show the pre- or post-message defined in the manifest            |

## Choria Configuration Manager Transform

The [Choria Configuration Manager](https://choria-io.github.io/ccm/) is a new Configuration Management tool that is part of the Choria project.

CCM manifests takes Data and Facts as input, we are adding a transform can execute a manifest with custom data.

This combines well with the new Form Based Wizards to ask users for configuration interactively. It also sets all flags and arguments as data.

```
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

This will set `version` in the data supplied to the manifest and execute the manifest. Setting `render_summary` will render the summary to STDOUT rather than return it as JSON.  Setting `no_render_messages` will avoid rendering the pre- and post-messages in the Manifest

There is also a top level command that can be used directly:

```
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

Here instead of a form we have a flag to pass