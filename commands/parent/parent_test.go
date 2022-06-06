// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package parent

import (
	"encoding/json"
	"testing"

	"github.com/choria-io/appbuilder/builder"
	"github.com/choria-io/fisk"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestParent(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ParentCommand")
}

var _ = Describe("Parent", func() {
	var p *Parent

	BeforeEach(func() {
		p = &Parent{def: &Command{}}
		p.def.Type = "parent"
	})

	Describe("CreateCommand", func() {
		It("Should create the command with aliases", func() {
			cmd := fisk.New("x", "y")
			p.def.Name = "ptest"
			p.def.Description = "ptest description"
			p.def.Aliases = []string{"p", "x"}

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
			Expect(err).To(MatchError("name is required, description is required, parent requires sub commands"))

			p.def.GenericCommand.Name = "ginkgo"
			p.def.GenericCommand.Description = "ginkgo description"
			err = p.Validate(nil)
			Expect(err).To(MatchError("parent requires sub commands"))
		})

		It("Should not allow flags, arguments or confirm", func() {
			p.def.GenericCommand.Name = "ginkgo"
			p.def.GenericCommand.Description = "ginkgo description"
			p.def.Commands = []json.RawMessage{[]byte("{}")}

			p.def.Flags = []builder.GenericFlag{{}}
			err := p.Validate(nil)
			Expect(err).To(MatchError("parent commands can not have flags"))

			p.def.Flags = nil
			p.def.Arguments = []builder.GenericArgument{{}}
			err = p.Validate(nil)
			Expect(err).To(MatchError("parent commands can not have arguments"))
		})

		It("Should require commands", func() {
			p.def.GenericCommand.Name = "ginkgo"
			p.def.GenericCommand.Description = "ginkgo description"

			err := p.Validate(nil)
			Expect(err).To(MatchError("parent requires sub commands"))

			p.def.Commands = []json.RawMessage{[]byte("{}")}
			err = p.Validate(nil)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
