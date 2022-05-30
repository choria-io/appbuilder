// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"testing"

	"github.com/alecthomas/kingpin"
	"github.com/choria-io/appbuilder/builder"
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
			cmd := kingpin.New("x", "y")
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
			Expect(err).To(MatchError("name is required, description is required, a command is required"))

			p.def.GenericCommand.Name = "ginkgo"
			p.def.GenericCommand.Description = "ginkgo description"
			err = p.Validate(nil)
			Expect(err).To(MatchError("a command is required"))
		})

		It("Should require a command", func() {
			p.def.GenericCommand.Name = "ginkgo"
			p.def.GenericCommand.Description = "ginkgo description"

			err := p.Validate(nil)
			Expect(err).To(MatchError("a command is required"))

			p.def.Command = "/bin/echo hello"
			err = p.Validate(nil)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
