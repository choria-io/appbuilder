// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/alecthomas/kingpin.v2"
)

var _ = Describe("GenericCommand", func() {
	var (
		cb  = func(_ *kingpin.ParseContext) error { return nil }
		def *GenericCommand
	)

	BeforeEach(func() {
		def = &GenericCommand{Name: "ginkgo", Description: "help", Aliases: []string{"x"}}
	})

	Describe("CreateGenericCommand", func() {
		It("Should create a basic command without flags or arguments", func() {
			def.Flags = []GenericFlag{
				{Name: "arg1", Description: "help1", Default: "default1", Enum: []string{"default1", "default2"}, Required: true},
			}
			def.Arguments = []GenericArgument{
				{Name: "arg1", Description: "help1", Default: "default1", Enum: []string{"default1", "default2"}, Required: true},
			}

			cmd := CreateGenericCommand(kingpin.New("app", "app"), def, nil, nil, cb)
			model := cmd.Model()

			Expect(model.Name).To(Equal("ginkgo"))
			Expect(model.Help).To(Equal("help"))
			Expect(model.Aliases).To(Equal([]string{"x"}))
			Expect(model.Flags).To(HaveLen(0))
			Expect(model.Args).To(HaveLen(0))
		})

		It("Should create a basic command with arguments", func() {
			def.Arguments = []GenericArgument{
				{Name: "arg1", Description: "help1", Default: "default1", Enum: []string{"default1", "default2"}, Required: true},
			}

			cmd := CreateGenericCommand(kingpin.New("app", "app"), def, map[string]*string{}, nil, cb)
			model := cmd.Model()
			Expect(model.Args).To(HaveLen(1))
			Expect(model.Args[0].Name).To(Equal("arg1"))
			Expect(model.Args[0].Help).To(Equal("help1"))
			Expect(model.Args[0].Default).To(Equal([]string{"default1"}))
		})

		It("Should create a basic command with flags", func() {
			def.Flags = []GenericFlag{
				{Name: "flag1", Description: "help1", Default: "default1", Enum: []string{"default1", "default2"}, Required: true},
			}

			cmd := CreateGenericCommand(kingpin.New("app", "app"), def, nil, map[string]*string{}, cb)
			model := cmd.Model()
			Expect(model.Flags).To(HaveLen(1))
			Expect(model.Flags[0].Name).To(Equal("flag1"))
			Expect(model.Flags[0].Help).To(Equal("help1"))
			Expect(model.Flags[0].Default).To(Equal([]string{"default1"}))
		})
	})
})
