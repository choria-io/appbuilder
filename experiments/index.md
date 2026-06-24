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

## Secrets

The `secrets` input resolves sensitive values at command-invocation time from an external store and exposes them to [templates](../reference/templating/) as `{{ .Secrets.<name> }}`. It sits alongside `flags` and `arguments` as a third declarative input. The first supported provider is [1Password](https://developer.1password.com/docs/cli/) accessed through the `op` CLI.

Secrets are resolved only when a command actually runs, after any confirmation prompt and never during `--help` or `validate`, so no `op` or biometric prompt fires until the user commits to running the command.

```yaml
name: demo
description: Demo application for Choria App Builder
author: https://github.com/choria-io/appbuilder
commands:
  - name: deploy
    description: Deploy using a token stored in 1Password
    type: exec
    secrets:
      - name: api_token
        description: API token for the example service
        one_password:
          item: Demo API Token
          field: credential
          vault: AppBuilderDemo
    environment:
      - "API_TOKEN={{ .Secrets.api_token }}"
    script: |
      curl -H "Authorization: Bearer ${API_TOKEN}" https://api.example.com
```

This reads the `credential` field of the `Demo API Token` item in the `AppBuilderDemo` vault and makes it available as `{{ .Secrets.api_token }}`. The value is passed to the script through the standard [`environment`](../reference/exec/) list rather than inlined into the command, which keeps it out of the process argument list visible to tools like `ps`.

The resolved value is available to `command`, `script`, `dir`, `environment` and any [transformations](../reference/transformations/) but not to banners, which render before resolution. Secret values are redacted from whole-state template dumps such as `{{ . }}` and `{{ toJson . }}`, while explicit references like `{{ .Secrets.api_token }}` resolve as normal.

Setting the `BUILDER_DRY_RUN` environment variable renders the command without contacting 1Password, resolving each secret to a `<secret:NAME>` placeholder instead.

Using secrets requires the [1Password CLI](https://developer.1password.com/docs/cli/) (`op`) to be installed with an active session. Access is read-only, App Builder never writes to the store.

Each entry in the `secrets` list accepts:

| Option         | Description                                                                                       |
|----------------|--------------------------------------------------------------------------------------------------|
| `name`         | A unique name used to reference the value as `{{ .Secrets.<name> }}`, must be a valid identifier  |
| `description`  | A human friendly description of what the secret is used for                                       |
| `one_password` | Resolves the value from a 1Password item, see below                                              |

The `one_password` provider accepts:

| Option    | Description                                                                 |
|-----------|-----------------------------------------------------------------------------|
| `item`    | The item name or ID holding the secret                                       |
| `field`   | The field within the item to read                                           |
| `vault`   | The vault holding the item, required by `op` secret references              |
| `account` | Optionally selects a specific 1Password account such as `my.1password.com`  |