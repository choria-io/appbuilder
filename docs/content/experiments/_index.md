+++
title = "Experiments"
toc = true
weight = 40
pre = "<b>4. </b>"
+++

Some features are ongoing experiments and not part of the supported feature set, this section will call them out.

## Including other definitions

Since version 0.10.0 an entire definition can be included from another file or just the commands in a parent.

```yaml
name: include
description: An include based app
version: 0.2.2
author: another@example.net

include_file: sample-app.yaml
```

Here we include the entire application from another file but we override the name, description, version and author.

A specific `parent` can load all it's commands from a file:

```yaml
  - name: include
    type: parent
    include_file: go.yaml
```

In this case the go.yaml would be the full `parent` definition.

## Form based data generation wizards

The general flow of applications is to expose Arguments and Flags when then can be used in templates to create files
or render some output.  This works quite well but can be limiting for more complex needs.

So we are introducing a full wizard style question-and-answer system that let you guide users through help, questions, 
validations and more to construct complex data.  The generated data supports almost everything JSON supports and can 
be deeply nested.

The general use case is to guide users through creating complex configuration files.

{{% notice secondary "Version Hint" code-branch %}}
This was added in version 0.9.0
{{% /notice %}}

It supports skipping sections of questions based on previous answers and generally tries to be a fully generic tool 
for getting data from users.

The gathered data can be sent to transforms for scaffolding or templating into a final form.

```yaml
commands:
  - name: configuration
    type: form
    properties:
      - name: listen
        description: The network address to listen on
        required: true
        default: 127.0.0.1:-1
        help: Examples include localhost:4222, 192.168.1.1:4222 or 127.0.0.1:4222
      - name: accounts
        description: Local accounts
        help: Sets up a local account for user access.
        type: object
        empty: absent
        properties:
        - name: users
          description: Users to add to the account
          required: true
          type: array
          properties:
            - name: user
              description: The username to connect as
              required: true
            - name: password
              description: The password to connect with
              type: password
              required: true
```

When run this looks a bit like this, with no transform the final data is just dumped to STDOUT:

```nohighlight
$ abt form
Demonstrates use of the form based data generator

? Press enter to start 

The network address and port to listen on

? listen 127.0.0.1:-1

Multiple accounts

? Add accounts entry Yes
? Unique name for this entry USERS

The username to connect as

? user user1

The password to connect with

? password ******
? Add additional 'users' entry No
? Add accounts entry Yes
? Unique name for this entry SYSTEM

The username to connect as

? user system

The password to connect with

? password ******
? Add additional 'users' entry No
? Add accounts entry No
{
  "SYSTEM": {
    "users": [
      {
        "password": "secret",
        "user": "system"
      }
    ]
  },
  "USERS": {
    "users": [
      {
        "password": "secret",
        "user": "user1"
      }
    ]
  },
  "listen": "127.0.0.1:-1"
}                              
```

The `form` command is a generic command with the only addition being an array of making up the questions `properties`, 
these are defined as below:

| Property      | Description                                                                                                                            |
|---------------|----------------------------------------------------------------------------------------------------------------------------------------|
| `name`        | Unique name for each property, in objects this would be the name of the key in the object                                              |
| `description` | Information shown to the user before asking the questions                                                                              |
| `help`        | Help shown when the user enters `?` in the prompt                                                                                      |
| `empty`       | What data to create when no values are given, one of `array`, `object`, `absent`                                                       |
| `type`        | The type of data to gather, one of `string`, `integer`, `float`, `bool`, `password`, `object` or `array`. Objects and Arrays will nest |
| `conditional` | An `expr` expression that looks back at the already-entered data and can be used to skip certain questions                             |
| `validation`  | A validation expression that will validate user input and ask the user to enter the value again on fail                                |
| `required`    | A value that is required cannot be skipped                                                                                             |
| `default`     | Default value to set                                                                                                                   |
| `enum`        | Will only allow one of these values to be set, presented as a select list                                                              |
| `properties`  | Nested questions to ask, array of properties as described in this table                                                                |

A full example can be seen in the `example` directory of the project.

Validation uses the validators shown in the next section - `Argument and Flag Validations` with `value` being the 
data just-entered by the user.

Conditional queries are also handled using `expr`, the example below looks back at the `accounts` entry (see example 
above) and will only ask this `thing` when the user opted to add accounts:

```yaml
  - name: thing
    description: Adds a thing if accounts are set
    empty: absent
    conditional: Input.accounts != nil
```

## Compiled Applications

It's nice that you do not need to compile App Builder apps into binaries as it allows for fast iteration, but sometimes
it might be desired.

As of version `0.7.2` we support compiling binaries that contain an application.

Given an application in `app.yaml` we can create a small go stub:

```go
package main

import (
	"context"
	_ "embed"
	"os"

	"github.com/choria-io/appbuilder/builder"
	"github.com/choria-io/fisk"
)

//go:embed app.yaml
var def []byte

func main() {
	builder.MustRegisterStandardCommands()

	cmd := fisk.Newf("myapp", "My compiled App Builder application")

	err := builder.MountAsCommand(context.TODO(), cmd, def, nil)
	if err != nil {
		panic(err)
	}

	cmd.MustParseWithUsage(os.Args[1:])
}
```

When you compile this as a normal Go application your binary will be an executable version of the app.

Here we mount the application at the top level of the `myapp` binary, but you could also mount it later on - perhaps you
have other compiled in behaviors you wish to surface:

```go
func main() {
	builder.MustRegisterStandardCommands()

	cmd := fisk.Newf("myapp", "My compiled App Builder application")
	embedded := cmd.Command("embedded","Embedded application goes here")

	err := builder.MountAsCommand(context.TODO(), embedded, def, nil)
	if err != nil {
		panic(err)
	}

	cmd.MustParseWithUsage(os.Args[1:])
}
```

Here we would end up with `myapp embedded [app commands]` - the command being mounted at a deeper level in the resulting
compiled application.  This way you can plug a App Builder command into any level programmatically.

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