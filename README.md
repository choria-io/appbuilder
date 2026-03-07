![Choria App Builder](https://github.com/choria-io/appbuilder/raw/main/images/logo.png)

## Overview

App Builder creates CLI applications from YAML definitions. Operations teams can wrap shell scripts, multi-line `kubectl` invocations, `jq` pipelines, and other operational tools into a single discoverable command.

Two binaries are produced from the same codebase:

* `appbuilder` - validates and inspects application definitions
* `abt` - project-specific task runner that searches for `ABTaskFile` definitions in the current directory tree

## Features

* Declarative YAML-based CLI definitions with nested sub-commands, flags, and arguments
* Command types: `exec` (shell commands and scripts), `parent` (command grouping), `form` (interactive wizards), `scaffold` (multi-file template generation), and `ccm_manifest` (Choria Config Manager)
* Output transformation pipeline with built-in JQ, ASCII graphs, Go templates, reports, and file writing
* Input validation using [expr](https://expr-lang.org) expressions
* Go template interpolation with [Sprig](https://masterminds.github.io/sprig/) functions
* Per-application configuration files
* Shell completion for `bash` and `zsh`

## Installation

Binary releases, RPMs, DEBs, and zip archives are available on the [Releases](https://github.com/choria-io/appbuilder/releases) page.

Homebrew packages are available for macOS and Linux:

```nohighlight
brew tap choria-io/tap
brew install choria-io/tap/appbuilder
```

A [JSON Schema](https://choria.io/schemas/appbuilder/v1/application.json) is available for editor integration.

## Quick Example

The following definition creates a `demo say` command that wraps `cowsay`:

```yaml
name: demo
description: Demo application
author: Operations <ops@example.net>
version: 1.0.0

commands:
  - name: say
    description: Say something using cowsay
    type: exec
    command: |
      {{ default .Config.Cowsay "cowsay" }} {{ .Arguments.message | escape }}
    arguments:
      - name: message
        description: The message to display
        required: true
```

```nohighlight
$ demo say "hello world"
 _____________
< hello world >
 -------------
        \   ^__^
         \  (oo)\_______
            (__)\       )\/\
                ||----w |
                ||     ||
```

## Links

* [Documentation](https://choria-io.github.io/appbuilder/)
* [Video Introduction](https://youtu.be/-IUwoXEJK0c)
* [Community](https://github.com/choria-io/appbuilder/discussions)
* [Code of Conduct](https://github.com/choria-io/.github/blob/master/CODE_OF_CONDUCT.md)
* [Contribution Guide](https://github.com/choria-io/.github/blob/master/CONTRIBUTING.md)
