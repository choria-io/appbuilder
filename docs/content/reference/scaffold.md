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

