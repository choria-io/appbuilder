+++
title = "File Locations"
toc = true
weight = 110
+++

The only configuration you should be concerned about is your Application Definition and optional Application Configuration.

We support the [XDG Base specification](https://github.com/adrg/xdg#xdg-base-directory), including standard environment variable based overrides like using `XDG_CONFIG_HOME`, for storing these in your home directory and have system wide fallback locations.

Files are stored in either `/etc/appbuilder/` or `~/.config/appbuilder` (`~/Library/Application Support/appbuilder` on a Mac).  When the symlink is created to a `choria` binary the locations `/etc/choria/builder` and  `~/.config/choria/builder` (`~/Library/Application Support/choria/builder` on a Mac) will also be searched in addition to the standard locations.

| File            | Description                                |
|-----------------|--------------------------------------------|
| `demo-app.yaml` | This is your application definition        |
| `demo-cfg.yaml` | This is your per-application configuration |
