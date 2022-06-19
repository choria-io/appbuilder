// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"os"
	"testing"

	"github.com/choria-io/appbuilder/builder"
	"github.com/choria-io/fisk"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestExec(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ExecCommand")
}

var _ = Describe("Exec", func() {
	var p *Exec

	BeforeEach(func() {
		p = &Exec{def: &Command{}, b: &builder.AppBuilder{}}
		p.def.Type = "exec"
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

	Describe("Validate", func() {
		It("Should do generic validations", func() {
			err := p.Validate(nil)
			Expect(err).To(MatchError("name is required, description is required, a command or script is required"))

			p.def.GenericCommand.Name = "ginkgo"
			p.def.GenericCommand.Description = "ginkgo description"
			err = p.Validate(nil)
			Expect(err).To(MatchError("a command or script is required"))
		})

		It("Should require a command", func() {
			p.def.GenericCommand.Name = "ginkgo"
			p.def.GenericCommand.Description = "ginkgo description"

			err := p.Validate(nil)
			Expect(err).To(MatchError("a command or script is required"))

			p.def.Command = "/bin/echo hello"
			err = p.Validate(nil)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should prevent command and script", func() {
			p.def.Script = "script"
			p.def.Command = "x"
			p.def.Name = "x"
			p.def.Description = "x"

			err := p.Validate(nil)
			Expect(err).To(MatchError("only one of command or script is allowed"))
		})
	})

	Describe("findShell", func() {
		It("Should support shell property in the definition", func() {
			p.def.Shell = "/bin/ginkgo"
			Expect(p.findShell()).To(Equal([]string{"/bin/ginkgo", "-c"}))

			p.def.Shell = "/bin/ginkgo -c"
			Expect(p.findShell()).To(Equal([]string{"/bin/ginkgo", "-c"}))
		})

		It("Should support SHELL", func() {
			pre := os.Getenv("SHELL")
			defer func() { os.Setenv("SHELL", pre) }()

			os.Setenv("SHELL", "/bin/ginkgo")
			Expect(p.findShell()).To(Equal([]string{"/bin/ginkgo", "-c"}))
		})

		It("Should fall back shell", func() {
			pre := os.Getenv("SHELL")
			defer func() { os.Setenv("SHELL", pre) }()

			os.Setenv("SHELL", "")
			Expect(p.findShell()).To(HaveLen(2))
		})
	})
})
