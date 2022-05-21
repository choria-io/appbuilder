+++
title = "File Locations"
toc = true
index = 30
+++

The only configuration you should be concerned about is your Application Definition and optional Application Configuration.

We support the XDG Base specification, including standard environment variable based overrides like using `XDG_CONFIG_HOME`, for storing these in your home directory and have system wide fallback locations.

Files are stored in either `~/.config/appbuilder` or `/etc/appbuilder/`.

| File            | Description                                |
|-----------------|--------------------------------------------|
| `demo-app.yaml` | This is your application definition        |
| `demo-cfg.yaml` | This is your per-application configuration |
