+++
title = "Form Command Type"
weight = 37
toc = true
+++

Use the `form` command to create guided wizard style question-and-answer sessions that construct complex data from user input.

The general use case is to guide users through creating complex configuration files. The gathered data is output as JSON and can be sent to [transforms](../transformations) for scaffolding or templating into a final form.

The `form` command supports [data transformations](../transformations), flags, arguments and sub commands.

{{% notice secondary "Version Hint" code-branch %}}
This was added in version 0.9.0
{{% /notice %}}

## Collecting data

A basic example that collects a network address and user accounts:

```yaml
name: configuration
description: Generate a configuration file
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
? Add accounts entry No
{
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

## Properties reference

The `form` command is a generic command with the only addition being an array of `properties` making up the questions:

| Property      | Description                                                                                                                            |
|---------------|----------------------------------------------------------------------------------------------------------------------------------------|
| `name`        | Unique name for each property, in objects this would be the name of the key in the object                                              |
| `description` | Information shown to the user before asking the questions                                                                               |
| `help`        | Help shown when the user enters `?` in the prompt                                                                                      |
| `empty`       | What data to create when no values are given, one of `array`, `object`, `absent`                                                       |
| `type`        | The type of data to gather, one of `string`, `integer`, `float`, `bool`, `password`, `object` or `array`. Objects and Arrays will nest |
| `conditional` | An `expr` expression that looks back at the already-entered data and can be used to skip certain questions                             |
| `validation`  | A validation expression that will validate user input and ask the user to enter the value again on fail                                |
| `required`    | A value that is required cannot be skipped                                                                                             |
| `default`     | Default value to set                                                                                                                   |
| `enum`        | Will only allow one of these values to be set, presented as a select list                                                              |
| `properties`  | Nested questions to ask, array of properties as described in this table                                                                |

## Validations

Validation uses the validators described in [Argument and Flag Validations](../common-settings/#argument-and-flag-validations) with `value` being the data just-entered by the user.

## Conditional questions

Conditional queries are handled using `expr`, the example below looks back at the `accounts` entry and will only ask this `thing` when the user opted to add accounts:

```yaml
  - name: thing
    description: Adds a thing if accounts are set
    empty: absent
    conditional: Input.accounts != nil
```

## Transforming output

The form output is JSON and can be processed through [transforms](../transformations). This combines well with the [scaffold](../scaffold) transform to generate files from the collected data:

```yaml
name: configuration
description: Generate configuration from user input
type: form

properties:
  - name: listen
    description: The network address to listen on
    required: true
    default: 127.0.0.1:4222

transform:
  scaffold:
    target: /etc/myapp
    source_directory: /usr/local/templates/config
```

A full example can be seen in the `example` directory of the project.