// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package example

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/choria-io/appbuilder/builder"
	"github.com/choria-io/appbuilder/commands/exec"
	"github.com/choria-io/appbuilder/commands/parent"
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

		cmd.Writer(usageBuf)
	})

	Describe("Top Level", func() {
		It("Should have the correct help", func() {
			cmd.MustParseWithUsage(strings.Fields(""))
			out := usageBuf.String()
			Expect(out).To(ContainSubstring("example: error: command not specified"))
			Expect(out).To(ContainSubstring("A sample application demonstrating App Builder features"))
			Expect(out).To(ContainSubstring("Use 'example cheat' to access cheat sheet style help"))
			Expect(out).To(ContainSubstring("Demonstrates transforming data using jq"))
		})
	})

	Describe("Cheats", func() {
		It("Should have a top level cheat", func() {
			cmd.MustParseWithUsage(strings.Fields("cheat sample"))
			Expect(usageBuf.String()).To(ContainSubstring("to see all the commands"))
		})

		It("Should have command cheats", func() {
			cmd.MustParseWithUsage(strings.Fields("cheat exec"))
			Expect(usageBuf.String()).To(ContainSubstring("to be asked a confirmation"))
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
	})

	Describe("exec", func() {
		Describe("confirm", func() {
			It("Should be able to override confirmation", func() {
				cmd.MustParseWithUsage(strings.Fields("exec confirm --no-prompt"))
				Expect(usageBuf.String()).To(ContainSubstring("execution confirmed"))
			})
		})

		Describe("banner", func() {
			It("Should print the banner and command output", func() {
				cmd.MustParseWithUsage(strings.Fields("exec banner"))
				out := usageBuf.String()
				Expect(out).To(ContainSubstring("This is a banner"))
				Expect(out).To(ContainSubstring("Command output"))
			})
		})

		Describe("env", func() {
			It("Should set the environment variable", func() {
				cmd.MustParseWithUsage(strings.Fields("exec env GINKGO"))
				Expect(usageBuf.String()).To(ContainSubstring("The supplied value set in APPVAR: \"GINKGO\""))
			})
		})

		Describe("jq", func() {
			It("Should get the data for known repos", func() {
				cmd.MustParseWithUsage(strings.Fields("exec jq"))
				Expect(usageBuf.String()).To(ContainSubstring("The latest release is: Release"))
			})

			It("Should accept arguments and handle failure", func() {
				cmd.MustParseWithUsage(strings.Fields("exec jq ripienaar"))
				Expect(usageBuf.String()).To(ContainSubstring("Release lookup failed"))
			})
		})
	})
})
