+++
title = "Runtime Settings and Tools"
toc = true
index = 70
+++

When just invoking `appbuilder` various utilities are exposed, your apps also take some Environment Variables as runtime
configuration.

## Builder Info

General runtime information can be printed:

```nohighlight
$ appbuilder info
Choria Application Builder

        Debug Logging (BUILDER_DEBUG): false
  Configuration File (BUILDER_CONFIG): not specified
        Definition File (BUILDER_APP): not specified
                     Source Locations: /home/example/.config/appbuilder, /etc/appbuilder

```

Here we can see where applications are loaded from and more.

## Run Time Configuration

As seen above a few variables are consulted, below a list with details:

| Variable         | Description                                                                                                |
|------------------|------------------------------------------------------------------------------------------------------------|
| `BUILDER_DEBUG`  | When set to any level debug logging will be shown to screen                                                |
| `BUILDER_CONFIG` | When invoking a command a custom configuration file can be loaded by setting the path in this variable     |
| `BUILDER_APP`    | When invoking a command a custom application definition can be loaded by setting the path in this variable |

With these variables set the `appbuilder info` command will update accordingly

## Finding Commands

All applications stored in source locations can be listed:

```nohighlight
$ appbuilder list
╭─────────────────────────────────────────────────────────────────────────────────────────╮
│                                   Known Applications                                    │
├────────┬──────────────────────────────────────────────┬─────────────────────────────────┤
│ Name   │ Location                                     │ Description                     │
├────────┼──────────────────────────────────────────────┼─────────────────────────────────┤
│ mycorp │ /home/rip/.config/appbuilder/mycorp-app.yaml │ A hello world sample Choria App │
╰────────┴──────────────────────────────────────────────┴─────────────────────────────────╯
```

## Validating Definitions

A recursive deep validate can be run across the entire definition which will highlight multiple errors in commands
and sub commands:

```nohighlight
$ appbuilder validate mycorp-app.yaml
Application definition mycorp-app.yaml not valid:

   root -> demo (parent): parent requires sub commands
   root -> demo (parent) -> echo (exec): a command is required
```
