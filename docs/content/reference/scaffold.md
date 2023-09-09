+++
title = "Scaffold Command Type"
weight = 35
toc = true
+++

Use the `scaffold` command to create directories of files based on templates.

{{% notice secondary "Version Hint" code-branch %}}
This was added in version 0.7.0
{{% /notice %}}

## Scaffolding files

First we will just show the most basic example:

```yaml
name: scaffold
description: Demonstrate scaffold features by creating some go files
type: scaffold
arguments:
  - name: target
    description: The target to create the files in
    required: true

target: "{{ .Arguments.target }}"
source:
  "main.go": |
    // Copyright {{ .Arguments.author }} {{ now | date "2006" }}
    
    package main
    import "{{ .Arguments.package }}/cmd"
    func main() {cmd.Run()}
```

This generates a file `main.go` in the directory set using the `target` argument.  The target directory must not exist.

Complex trees can be created like this:

```yaml
source:
  "cmd":
    "cmd.go": |
      // content not shown
  "main.go": |
    // content not shown
```

Here we will have a directory `cmd` with `cmd/cmd.go` inside along with top level `main.go`.

## Storing files externally

In the example we have the template embedded in the YAML file, its functional but does not really scale well.

You can create a directory full of template files that mirror the target directory layer do this instead:

```yaml
name: scaffold
description: Demonstrate scaffold features by creating some go files
type: scaffold
arguments:
  - name: target
    description: The target to create the files in
    required: true
flags:
  - name: template
    description: The template to use
    default: golang
    
target: "{{ .Arguments.target }}"
source_directory: /usr/local/templates/{{ .Flags.template }}
```

Now we will use `/usr/local/template/golang` by default and whatever is passed in `--template` instead of `golang` 
otherwise.

## Post processing files

In the first example we showed a poorly formatted go file, the result will be equally badly formatted.

Here we show how to post process the files using `gofmt`:

```yaml
name: scaffold
description: Demonstrate scaffold features by creating some go files
type: scaffold
arguments:
  - name: target
    description: The target to create the files in
    required: true

target: "{{ .Arguments.target }}"
source_directory: /usr/local/templates/default

post:
  - "*.go": "gofmt -w"
  - "*.go": "goimports -w '{}'"
```

The new `post` structure defines a list of processors based on a file pattern match done using `filepath.Match`.

As shown the same pattern can be matched multiple times to run multiple commands on the file.

If the string `{}` is in the file it will be replaced with the full path to the file otherwise the path is set as 
last argument. When using this format it's suggested you use quotes like in the example.

## Conditional rendering

By default all files are rendered even when the result is empty, by setting `skip_empty: true` any file that results in
empty content will be skipped.

```yaml
name: scaffold
description: Demonstrate scaffold features by creating some go files
type: scaffold
arguments:
- name: target
  description: The target to create the files in
  required: true
flags:
- name: gitignore
  description: Create a .gitignore file
  bool: true
  default: true
  
target: "{{ .Arguments.target }}"
source_directory: /usr/local/templates/default
skip_empty: true
```

We can now create a template for the `.gitignore` file like this:

```
{{ if .Flags.gitignore }}
# content here
{{ end }}
```

This will result in a file that is empty - or rather just white space in this case - this file will be ignored and not
written to disk. 

## Rendering partials

We support partials that can be reused, any files in the `_partials` directory will be skipped for normal processing,
you can reference these files from other files:

{{% notice secondary "Version Hint" code-branch %}}
This was added in version 0.7.4
{{% /notice %}}

```
{{ render "_partials/go_copyright" . }}

package main

func main() {
}
```

If you now made a file `_partials/go_copyright` in your source templates holding the following:

```
// Copyright {{ .Arguments.author }} {{ now | date "2006" }}
```

You can easily reuse the content of the Copyright strings and update all in one place later.

## Rendering files from templates

It's often the case that you need to create new files that is not in the actual template source.  Perhaps you ask a
user how many of a certain thing they need and then you need to create that many files.  This means you will likely
have a Partial that can be used to make the file and need to invoke it many times.

{{% notice secondary "Version Hint" code-branch %}}
This was added in version 0.7.4
{{% /notice %}}

To use this you can store a template in the `_partials` directory and then render files like this:

```
{{- $flags := .Flags }}
{{- range $cluster := $flags.Clusters | atoi | seq | split " " }}
{{- $config :=  cat "cluster-" $cluster ".conf" | nospace }} 
{{- render "_partials/cluster.conf" $flags | write $config  }}
{{- end }}
```

This will render and, using the `write` helper, save `cluster-{1,2,3,...}.conf` for how many ever clusters you had in 
Flags. The file will be post processed as normal and written relative to the target directory.

We save `.Flags` in `$flags` because within the `range` the `.` will not point to the top anymore, so this ensures we
can access the passed in flags in the `_partials/cluster.conf` template.

If you place this loop in a file that is only there to generate these other files then the resulting empty 
file can be ignored using `skip_empty: true` in the scaffold definition.

## Custom template delimiter

When generating Go projects you might find you want to place template tags into the final project, for example when
generating a `ABTaskFile`.

With the final `ABTaskFile` having the same template delimiters will cause havoc.

You can change the delimiters of the template source to avoid this:

```yaml
name: scaffold
description: Demonstrate scaffold features by creating some go files
type: scaffold
arguments:
- name: target
  description: The target to create the files in
  required: true
  
target: "{{ .Arguments.target }}"
source_directory: /usr/local/templates/default
skip_empty: true
left_delimiter: "[["
right_delimiter: "]]"
```

Our earlier .gitignore would now be:

```
[[ if .Flags.gitignore ]]
# content here {{ these will not be changed }}
[[ end ]]
```

