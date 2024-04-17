// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package example

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/choria-io/appbuilder/builder"
	"github.com/choria-io/appbuilder/commands/exec"
	"github.com/choria-io/appbuilder/commands/form"
	"github.com/choria-io/appbuilder/commands/parent"
	"github.com/choria-io/appbuilder/commands/scaffold"
	"github.com/choria-io/fisk"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBuilder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Example")
}

var _ = Describe("Example Application", func() {
	var app *builder.AppBuilder
	var err error
	var cmd *fisk.Application
	var usageBuf = bytes.NewBuffer([]byte{})

	BeforeEach(func() {
		exec.Register()
		parent.Register()
		scaffold.Register()
		form.Register()

		app, err = builder.New(context.Background(), "example",
			builder.WithAppDefinitionFile("sample-app.yaml"),
			builder.WithLogger(&builder.NoopLogger{}),
			builder.WithContextualUsageOnError(),
			builder.WithStdout(usageBuf),
			builder.WithStderr(usageBuf))
		Expect(err).ToNot(HaveOccurred())
		cmd, err = app.FiskApplication()
		Expect(err).ToNot(HaveOccurred())
		cmd.Terminate(func(int) {})

		usageBuf.Reset()
		cmd.Writer(usageBuf)
	})

	Describe("Top Level", func() {
		It("Should have the correct help", func() {
			cmd.MustParseWithUsage(strings.Fields(""))
			out := usageBuf.String()
			Expect(out).To(ContainSubstring("A sample application demonstrating App Builder features"))
			Expect(out).To(ContainSubstring("Use 'example cheat' to access cheat sheet style help"))
			Expect(out).To(ContainSubstring("basics      Demonstrates basic features such as flags and arguments"))
		})
	})

	Describe("Cheats", func() {
		It("Should have a top level cheat", func() {
			cmd.MustParseWithUsage(strings.Fields("cheat sample"))
			Expect(usageBuf.String()).To(ContainSubstring("to see all the commands"))
		})

		It("Should have command cheats", func() {
			cmd.MustParseWithUsage(strings.Fields("cheat confirm"))
			Expect(usageBuf.String()).To(ContainSubstring("to be asked a confirmation"))
		})
	})

	Describe("Validation", func() {
		It("Should correctly validate options", func() {
			usageBuf.Reset()

			cmd.MustParseWithUsage(strings.Fields("basics required ginkgoginkgoginkgoginkgoginkgoginkgo"))
			Expect(usageBuf.String()).To(ContainSubstring(`name: validation using "len(value) < 20" did not pass`))
		})
	})

	Describe("Basics", func() {
		Describe("required", func() {
			It("Should require a name", func() {
				cmd.MustParseWithUsage(strings.Fields("basics required"))
				Expect(usageBuf.String()).To(ContainSubstring("error: required argument 'name' not provided"))

				usageBuf.Reset()
				cmd.MustParseWithUsage(strings.Fields("basics required ginkgo"))
				Expect(usageBuf.String()).To(ContainSubstring("Hello ginkgo"))
			})

			It("Should support a surname", func() {
				cmd.MustParseWithUsage(strings.Fields("basics required ginkgo bond"))
				Expect(usageBuf.String()).To(ContainSubstring("Hello bond, ginkgo bond"))
			})

			It("Should support a custom greeting as a flag", func() {
				cmd.MustParseWithUsage(strings.Fields("basics required ginkgo bond --greeting=Halo"))
				Expect(usageBuf.String()).To(ContainSubstring("Halo bond, ginkgo bond"))

				usageBuf.Reset()
				cmd.MustParseWithUsage(strings.Fields("basics required ginkgo bond -g Halo"))
				Expect(usageBuf.String()).To(ContainSubstring("Halo bond, ginkgo bond"))
			})

			It("Should support a custom greeting from the environment", func() {
				os.Setenv("GREETING", "Morning")
				cmd.MustParseWithUsage(strings.Fields("basics required ginkgo bond"))
				Expect(usageBuf.String()).To(ContainSubstring("Morning bond, ginkgo bond"))
			})

			It("Should enforce the enum", func() {
				cmd.MustParseWithUsage(strings.Fields("basics required ginkgo bond -g Foo"))
				Expect(usageBuf.String()).To(ContainSubstring("error: enum value must be one of Hello,Morning,Halo, got 'Foo'"))
			})
		})

		Describe("booleans", func() {
			It("Should have a no version of banner", func() {
				cmd.MustParseWithUsage(strings.Fields("basics booleans --help"))
				Expect(usageBuf.String()).To(ContainSubstring("--[no-]banner"))
			})

			It("Should not have a no version of silent", func() {
				cmd.MustParseWithUsage(strings.Fields("basics booleans --help"))
				Expect(usageBuf.String()).To(ContainSubstring("--silent"))
			})

			It("Values should be set correctly", func() {
				cmd.MustParseWithUsage(strings.Fields("basics booleans"))
				out := usageBuf.String()
				Expect(out).To(ContainSubstring("This is a banner"))
				Expect(out).To(ContainSubstring("Hello world"))

				usageBuf.Reset()
				cmd.MustParseWithUsage(strings.Fields("basics booleans --no-banner"))
				out = usageBuf.String()
				Expect(out).ToNot(ContainSubstring("This is a banner"))
				Expect(out).To(ContainSubstring("Hello world"))

				usageBuf.Reset()
				cmd.MustParseWithUsage(strings.Fields("basics booleans --silent"))
				Expect(usageBuf.String()).To(HaveLen(0))
			})
		})

		Describe("confirm", func() {
			It("Should be able to override confirmation", func() {
				cmd.MustParseWithUsage(strings.Fields("basics confirm --no-prompt"))
				Expect(usageBuf.String()).To(ContainSubstring("execution confirmed"))
			})
		})

		Describe("banner", func() {
			It("Should print the banner and command output", func() {
				cmd.MustParseWithUsage(strings.Fields("basics banner"))
				out := usageBuf.String()
				Expect(out).To(ContainSubstring("This is a banner"))
				Expect(out).To(ContainSubstring("Command output"))
			})
		})
	})

	Describe("exec", func() {
		Describe("env", func() {
			It("Should set the environment variable", func() {
				cmd.MustParseWithUsage(strings.Fields("exec env GINKGO"))
				Expect(usageBuf.String()).To(ContainSubstring("The supplied value set in APPVAR: \"GINKGO\""))
			})
		})

		Describe("Shell helpers", func() {
			It("Should support exposing it to scripts and commands", func() {
				cmd.MustParseWithUsage(strings.Fields("exec shell_helper"))
				Expect(usageBuf.String()).To(ContainSubstring("???\n??? Demonstrates using the shell helper\n???"))
			})
		})
	})

	Describe("transforms", func() {
		Describe("jq", func() {
			It("Should get the data for known repos", func() {
				cmd.MustParseWithUsage(strings.Fields("transforms jq"))
				Expect(usageBuf.String()).To(ContainSubstring("The latest release is: Release"))
			})

			It("Should accept arguments and handle failure", func() {
				cmd.MustParseWithUsage(strings.Fields("t jq ripienaar"))
				Expect(usageBuf.String()).To(ContainSubstring("Release lookup failed"))
			})
		})

		Describe("Templates", func() {
			It("Should parse and render the template and should include sprout functions", func() {
				cmd.MustParseWithUsage(strings.Fields("transforms template"))
				Expect(usageBuf.String()).To(ContainSubstring("Hello James bOND"))
			})
		})

		Describe("Scaffold", func() {
			It("Should render the correct files", func() {
				td, err := os.MkdirTemp("", "")
				Expect(err).ToNot(HaveOccurred())
				os.Remove(td)
				defer os.RemoveAll(td)

				cmd.MustParseWithUsage(strings.Fields(fmt.Sprintf("scaffold Ginkgo example.net/test %s", td)))

				fmt.Println(usageBuf.String())

				readFile := func(f string) string {
					b, err := os.ReadFile(filepath.Join(td, f))
					if err != nil {
						Fail(fmt.Sprintf("Reading %s failed: %v", f, err))
					}
					return string(b)
				}

				Expect(readFile("main.go")).To(SatisfyAll(
					ContainSubstring("// Copyright Ginkgo"),
					ContainSubstring("cmd.Run()"),
				))
				Expect(readFile("README.md")).To(ContainSubstring("## Copyright Ginkgo"))
				Expect(readFile("cmd/cmd.go")).To(SatisfyAll(
					ContainSubstring("// Copyright Ginkgo"),
					ContainSubstring(`fmt.Println("Scaffolded using App Builder")`),
				))
			})
		})
	})
})
