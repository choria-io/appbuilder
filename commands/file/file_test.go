// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package file

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/choria-io/appbuilder/builder"
	"github.com/choria-io/fisk"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFile(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FileCommand")
}

var _ = Describe("File", func() {
	var (
		p   *File
		b   *builder.AppBuilder
		out *bytes.Buffer
		td  string
		err error
	)

	// creates a file below the temporary directory and returns its full path
	createFile := func(name string, contents []byte) string {
		path := filepath.Join(td, name)

		err := os.MkdirAll(filepath.Dir(path), 0700)
		Expect(err).ToNot(HaveOccurred())

		err = os.WriteFile(path, contents, 0600)
		Expect(err).ToNot(HaveOccurred())

		return path
	}

	strPtr := func(s string) *string { return &s }

	BeforeEach(func() {
		td, err = os.MkdirTemp("", "")
		Expect(err).ToNot(HaveOccurred())
		DeferCleanup(func() { os.RemoveAll(td) })

		out = bytes.NewBuffer(nil)

		b, err = builder.New(context.Background(), "ginkgo",
			builder.WithLogger(&builder.NoopLogger{}),
			builder.WithStdout(out),
			builder.WithStderr(out))
		Expect(err).ToNot(HaveOccurred())

		p = &File{
			def:       &Command{},
			b:         b,
			ctx:       context.Background(),
			log:       &builder.NoopLogger{},
			arguments: map[string]any{},
			flags:     map[string]any{},
		}
		p.def.Type = "file"
		p.def.Name = "ginkgo"
		p.def.Description = "ginkgo description"
	})

	Describe("NewFileCommand", func() {
		It("Should detect invalid definitions", func() {
			_, err := NewFileCommand(b, json.RawMessage(`{"file":1}`), &builder.NoopLogger{})
			Expect(err).To(MatchError(builder.ErrInvalidDefinition))
		})

		It("Should parse the definition", func() {
			c, err := NewFileCommand(b, json.RawMessage(`{"name":"ginkgo","description":"ginkgo description","type":"file","file":"/etc/motd","chroot":"/etc","action":"read"}`), &builder.NoopLogger{})
			Expect(err).ToNot(HaveOccurred())

			f := c.(*File)
			Expect(f.def.File).To(Equal("/etc/motd"))
			Expect(f.def.Chroot).To(Equal("/etc"))
			Expect(f.def.Action).To(Equal("read"))
			Expect(f.String()).To(Equal("ginkgo (file)"))
		})
	})

	Describe("Validate", func() {
		It("Should ensure the command is a file command", func() {
			p.def.Type = "exec"
			Expect(p.Validate(nil)).To(MatchError("not a file command"))
		})

		It("Should do generic validations", func() {
			p.def.Name = ""
			p.def.Description = ""

			err := p.Validate(nil)
			Expect(err).To(MatchError(ErrorInvalidConfiguration))
			Expect(err.Error()).To(Equal("invalid configuration: name is required, description is required, file is required"))
		})

		It("Should require a file", func() {
			err := p.Validate(nil)
			Expect(err.Error()).To(Equal("invalid configuration: file is required"))

			p.def.File = "/etc/motd"
			Expect(p.Validate(nil)).ToNot(HaveOccurred())
		})

		It("Should default the action to read", func() {
			p.def.File = "/etc/motd"
			Expect(p.Validate(nil)).ToNot(HaveOccurred())
			Expect(p.def.Action).To(Equal("read"))
		})

		It("Should accept the read action", func() {
			p.def.File = "/etc/motd"
			p.def.Action = "read"
			Expect(p.Validate(nil)).ToNot(HaveOccurred())
			Expect(p.def.Action).To(Equal("read"))
		})

		It("Should reject other actions", func() {
			p.def.File = "/etc/motd"

			for _, action := range []string{"write", "delete", "READ", "append"} {
				p.def.Action = action
				err := p.Validate(nil)
				Expect(err).To(MatchError(ErrorInvalidConfiguration))
				Expect(err.Error()).To(Equal(fmt.Sprintf("invalid configuration: invalid action %q, only read is supported", action)))
			}
		})
	})

	Describe("SubCommands", func() {
		It("Should expose the sub commands", func() {
			Expect(p.SubCommands()).To(BeNil())

			p.def.Commands = []json.RawMessage{[]byte(`{"name":"sub"}`)}
			Expect(p.SubCommands()).To(HaveLen(1))
		})
	})

	Describe("CreateCommand", func() {
		It("Should create the command with aliases", func() {
			cmd := fisk.New("x", "y")
			p.def.Name = "ptest"
			p.def.Description = "ptest description"
			p.def.Aliases = []string{"p", "x"}
			p.def.Flags = []builder.GenericFlag{
				{Name: "flag", Description: "flag description"},
			}

			sub, err := p.CreateCommand(cmd)
			Expect(err).ToNot(HaveOccurred())

			model := sub.Model()
			Expect(model.Aliases).To(Equal([]string{"p", "x"}))
			Expect(model.Name).To(Equal("ptest"))
			Expect(model.Help).To(Equal("ptest description"))
		})
	})

	Describe("runCommand", func() {
		Describe("Without a chroot", func() {
			It("Should write the file bytes to stdout", func() {
				p.def.File = createFile("motd.txt", []byte("hello world\n"))

				Expect(p.runCommand(nil)).ToNot(HaveOccurred())
				Expect(out.String()).To(Equal("hello world\n"))
			})

			It("Should not modify the bytes in any way", func() {
				data := []byte{0x00, 0x01, 0xff, 'a', '\n', 'b'}
				p.def.File = createFile("binary.dat", data)

				Expect(p.runCommand(nil)).ToNot(HaveOccurred())
				Expect(out.Bytes()).To(Equal(data))
			})

			It("Should handle empty files", func() {
				p.def.File = createFile("empty.txt", []byte{})

				Expect(p.runCommand(nil)).ToNot(HaveOccurred())
				Expect(out.Len()).To(Equal(0))
			})

			It("Should fail for missing files", func() {
				p.def.File = filepath.Join(td, "missing.txt")

				err := p.runCommand(nil)
				Expect(err).To(MatchError(os.ErrNotExist))
				Expect(out.Len()).To(Equal(0))
			})

			It("Should fail when the file is a directory", func() {
				p.def.File = td

				Expect(p.runCommand(nil)).To(HaveOccurred())
			})
		})

		Describe("With a chroot", func() {
			BeforeEach(func() {
				p.def.Chroot = td
			})

			It("Should read files relative to the chroot", func() {
				createFile("motd.txt", []byte("hello chroot\n"))
				p.def.File = "motd.txt"

				Expect(p.runCommand(nil)).ToNot(HaveOccurred())
				Expect(out.String()).To(Equal("hello chroot\n"))
			})

			It("Should support files in sub directories", func() {
				createFile("sub/dir/motd.txt", []byte("hello sub dir\n"))
				p.def.File = "sub/dir/motd.txt"

				Expect(p.runCommand(nil)).ToNot(HaveOccurred())
				Expect(out.String()).To(Equal("hello sub dir\n"))
			})

			It("Should not allow traversal out of the chroot", func() {
				outside := createFile("outside.txt", []byte("secret\n"))
				p.def.Chroot = createDir(td, "root")
				p.def.File = filepath.Join("..", filepath.Base(outside))

				Expect(p.runCommand(nil)).To(HaveOccurred())
				Expect(out.Len()).To(Equal(0))
			})

			It("Should not allow absolute paths out of the chroot", func() {
				outside := createFile("outside.txt", []byte("secret\n"))
				p.def.Chroot = createDir(td, "root")
				p.def.File = outside

				Expect(p.runCommand(nil)).To(HaveOccurred())
				Expect(out.Len()).To(Equal(0))
			})

			It("Should fail for missing files", func() {
				p.def.File = "missing.txt"

				err := p.runCommand(nil)
				Expect(err).To(MatchError(os.ErrNotExist))
				Expect(out.Len()).To(Equal(0))
			})

			It("Should fail for a missing chroot", func() {
				p.def.Chroot = filepath.Join(td, "missing")
				p.def.File = "motd.txt"

				Expect(p.runCommand(nil)).To(HaveOccurred())
				Expect(out.Len()).To(Equal(0))
			})
		})

		Describe("Templates", func() {
			It("Should render the file using arguments and flags", func() {
				createFile("ginkgo.txt", []byte("templated\n"))

				p.def.Chroot = td
				p.def.File = "{{ .Arguments.name }}.{{ .Flags.extension }}"
				p.arguments["name"] = strPtr("ginkgo")
				p.flags["extension"] = strPtr("txt")

				Expect(p.runCommand(nil)).ToNot(HaveOccurred())
				Expect(out.String()).To(Equal("templated\n"))
			})

			It("Should render the chroot using arguments and flags", func() {
				createFile("root/motd.txt", []byte("templated chroot\n"))

				p.def.Chroot = filepath.Join(td, "{{ .Arguments.dir }}")
				p.def.File = "motd.txt"
				p.arguments["dir"] = strPtr("root")

				Expect(p.runCommand(nil)).ToNot(HaveOccurred())
				Expect(out.String()).To(Equal("templated chroot\n"))
			})

			It("Should support sprig functions", func() {
				createFile("ginkgo.txt", []byte("sprig\n"))

				p.def.Chroot = td
				p.def.File = "{{ .Arguments.name | lower }}.txt"
				p.arguments["name"] = strPtr("GINKGO")

				Expect(p.runCommand(nil)).ToNot(HaveOccurred())
				Expect(out.String()).To(Equal("sprig\n"))
			})

			It("Should detect invalid file templates", func() {
				p.def.File = "{{ .Arguments.name "

				err := p.runCommand(nil)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HavePrefix("invalid file template: "))
			})

			It("Should detect invalid chroot templates", func() {
				p.def.File = "motd.txt"
				p.def.Chroot = "{{ .Arguments.dir "

				err := p.runCommand(nil)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HavePrefix("invalid chroot template: "))
			})

			It("Should treat a chroot that renders empty as no chroot", func() {
				p.def.File = createFile("motd.txt", []byte("no chroot\n"))
				p.def.Chroot = "{{ .Flags.chroot }}"
				p.flags["chroot"] = strPtr("")

				Expect(p.runCommand(nil)).ToNot(HaveOccurred())
				Expect(out.String()).To(Equal("no chroot\n"))
			})
		})
	})

	Describe("Application", func() {
		var cmd *fisk.Application

		BeforeEach(func() {
			Register()

			createFile("motd.txt", []byte("hello application\n"))

			def := `
name: ginkgo
description: ginkgo application
version: 0.0.1
commands:
  - name: read
    description: reads a file
    type: file
    action: read
    chroot: "{{ .Flags.chroot }}"
    file: "{{ .Arguments.file }}"
    arguments:
      - name: file
        description: the file to read
        required: true
    flags:
      - name: chroot
        description: the directory to read within
`

			err := os.WriteFile(filepath.Join(td, "app.yaml"), []byte(def), 0600)
			Expect(err).ToNot(HaveOccurred())

			b, err = builder.New(context.Background(), "ginkgo",
				builder.WithAppDefinitionFile(filepath.Join(td, "app.yaml")),
				builder.WithLogger(&builder.NoopLogger{}),
				builder.WithContextualUsageOnError(),
				builder.WithStdout(out),
				builder.WithStderr(out))
			Expect(err).ToNot(HaveOccurred())

			cmd, err = b.FiskApplication()
			Expect(err).ToNot(HaveOccurred())
			cmd.Terminate(func(int) {})
			cmd.Writer(out)
		})

		It("Should read files using the definition", func() {
			cmd.MustParseWithUsage(strings.Fields(fmt.Sprintf("read %s", filepath.Join(td, "motd.txt"))))
			Expect(out.String()).To(Equal("hello application\n"))
		})

		It("Should read files within the chroot", func() {
			cmd.MustParseWithUsage(strings.Fields(fmt.Sprintf("read motd.txt --chroot %s", td)))
			Expect(out.String()).To(Equal("hello application\n"))
		})

		It("Should not read files outside of the chroot", func() {
			createFile("root/keep.txt", []byte("keep\n"))

			cmd.MustParseWithUsage(strings.Fields(fmt.Sprintf("read ../motd.txt --chroot %s", filepath.Join(td, "root"))))
			Expect(out.String()).ToNot(ContainSubstring("hello application"))
			Expect(out.String()).To(ContainSubstring("error"))
		})
	})
})

// creates a directory below parent and returns its full path
func createDir(parent string, name string) string {
	path := filepath.Join(parent, name)

	err := os.MkdirAll(path, 0700)
	Expect(err).ToNot(HaveOccurred())

	return path
}
