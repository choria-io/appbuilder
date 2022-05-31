+++
title = "KV Command Type"
weight = 30
toc = true
+++

The KV command interact with the Choria Key-Value Store and supports usual operations such as Get, Put, Delete and more.

{{% notice secondary "Version Hint" code-branch %}}
This feature is only available when hosting App Builder applications within the Choria Server version 0.26.0 or newer
{{% /notice %}}

## Overview

All variations of this command have a number of required properties, here's the basic `get` operation, all these keys are required:

```yaml
name: version
description: Retrieve the `version` key
type: kv

action: get
bucket: DEPLOYMENT
key: version
```

Usual standard properties like `flags`, `arguments`, `commands` and so forth are all supported. The `bucket` and `key` flags supports [templating](../templating/).

## Writing data using `put`

Data can be written to the bucket, it's identical to the above example with the addition of the `value` property that supports [templating](../templating/).

```yaml
name: version
description: Stores a new version for the deployment
type: kv

action: put
bucket: DEPLOYMENT
key: version
value: '{{- .Arguments.version -}}'
arguments:
  - name: version
    description: The version to store
    required: true
```

## Retrieving data and transformations using `get`

Stored data can be retrieved and rendered to the screen, typically the value is just dumped. Keys and Values however have
additional metadata that can be rendered in JSON format.

```yaml
name: version
description: Retrieve the `version` key
type: kv

action: get
bucket: DEPLOYMENT
key: state

# Triggers rendering the KV entry as JSON that will include metadata ab out the value.
json: true
```
Further if it's known that the entry holds JSON data it can be formatted using GOJQ:

```yaml
name: version
description: Retrieve the `version` key
type: kv

action: get
bucket: DEPLOYMENT
key: state

transform:
  query: .state
```

## Deleting data using `del`

Deleting a specific key is very similar to a basic retrieve, just use a different `action`:

```yaml
name: version
description: Deletes the deployment configuration property
type: kv

action: del
bucket: DEPLOYMENT
key: configuration
```

## Viewing key history using `history`

Choria Key-Value store optionally has historical data for keys, the data can be shown in tabular (default) or JSON formats:

```yaml
name: version
description: Deploy version history
type: kv

action: history
bucket: DEPLOYMENT
key: version

# optionally renders the result as JSON
json: true
```
