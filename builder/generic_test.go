// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"context"

	"github.com/choria-io/fisk"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GenericCommand", func() {
	var (
		cb  = func(_ *fisk.ParseContext) error { return nil }
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

			cmd := CreateGenericCommand(fisk.New("app", "app"), def, nil, nil, nil, cb)
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

			cmd := CreateGenericCommand(fisk.New("app", "app"), def, map[string]any{}, nil, nil, cb)
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

			cmd := CreateGenericCommand(fisk.New("app", "app"), def, nil, map[string]any{}, nil, cb)
			model := cmd.Model()
			Expect(model.Flags).To(HaveLen(1))
			Expect(model.Flags[0].Name).To(Equal("flag1"))
			Expect(model.Flags[0].Help).To(Equal("help1"))
			Expect(model.Flags[0].Default).To(Equal([]string{"default1"}))
		})
	})
})

var _ = Describe("GenericTransform", func() {
	var (
		trans *Transform
	)

	BeforeEach(func() {
		trans = &Transform{}
	})

	Describe("Validate", func() {
		It("Should detect absent queries", func() {
			err := trans.Validate(nil)
			Expect(err).To(MatchError(ErrInvalidTransform))
		})
	})

	Describe("Transform", func() {
		It("Should transform using the query", func() {
			_, err := trans.TransformBytes(context.Background(), []byte(`{"hello":"world"`), nil, nil, nil)
			Expect(err).To(MatchError(ErrInvalidTransform))

			trans.Query = ".hello"
			Expect(trans.Validate(nil)).To(Succeed())

			_, err = trans.TransformBytes(context.Background(), []byte(`{`), nil, nil, nil)
			Expect(err).To(MatchError("json input parse error: unexpected end of JSON input"))

			res, err := trans.TransformBytes(context.Background(), []byte(`{"hello":"world"}`), nil, nil, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(res)).To(Equal("world\n"))
		})
	})
})
