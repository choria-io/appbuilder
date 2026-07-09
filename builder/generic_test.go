// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"context"
	"fmt"
	"sort"

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

		DescribeTable("Should map flag types to the correct fisk value",
			func(typ string, expected string) {
				flags := map[string]any{}
				def.Flags = []GenericFlag{{Name: "f", Description: "help", Type: typ}}
				CreateGenericCommand(fisk.New("app", "app"), def, nil, flags, nil, cb)
				Expect(fmt.Sprintf("%T", flags["f"])).To(Equal(expected))
			},
			Entry("no type defaults to string", "", "*string"),
			Entry("string", "string", "*string"),
			Entry("bool", "bool", "*bool"),
			Entry("int", "int", "*int"),
			Entry("integer alias", "integer", "*int"),
			Entry("int64", "int64", "*int64"),
			Entry("uint", "uint", "*uint"),
			Entry("float", "float", "*float64"),
			Entry("float32", "float32", "*float32"),
			Entry("counter", "counter", "*int"),
			Entry("existing_file", "existing_file", "*string"),
			Entry("case insensitive and trimmed", " INT ", "*int"),
			Entry("unknown falls back to string", "nope", "*string"),
		)

		It("Should honor the legacy bool field over a set type", func() {
			flags := map[string]any{}
			def.Flags = []GenericFlag{{Name: "f", Description: "help", Bool: true, Type: "int"}}
			CreateGenericCommand(fisk.New("app", "app"), def, nil, flags, nil, cb)
			Expect(fmt.Sprintf("%T", flags["f"])).To(Equal("*bool"))
		})

		It("Should place arguments and flags in their own maps", func() {
			args := map[string]any{}
			flags := map[string]any{}
			def.Arguments = []GenericArgument{{Name: "a", Description: "help", Type: "int"}}
			def.Flags = []GenericFlag{{Name: "f", Description: "help", Type: "string"}}

			CreateGenericCommand(fisk.New("app", "app"), def, args, flags, nil, cb)

			Expect(args).To(HaveKey("a"))
			Expect(args).ToNot(HaveKey("f"))
			Expect(flags).To(HaveKey("f"))
			Expect(flags).ToNot(HaveKey("a"))
			Expect(fmt.Sprintf("%T", args["a"])).To(Equal("*int"))
			Expect(fmt.Sprintf("%T", flags["f"])).To(Equal("*string"))
		})

		It("Should enforce types and expose typed values when parsing", func() {
			build := func() (map[string]any, *fisk.Application) {
				flags := map[string]any{}
				d := &GenericCommand{Name: "ginkgo", Type: "exec", Description: "help"}
				d.Flags = []GenericFlag{{Name: "count", Description: "help", Type: "int", Default: "3"}}
				app := fisk.New("app", "app")
				app.Terminate(func(int) {})
				CreateGenericCommand(app, d, nil, flags, &AppBuilder{cfg: map[string]any{}}, cb)
				return flags, app
			}

			flags, app := build()
			_, err := app.Parse([]string{"ginkgo"})
			Expect(err).ToNot(HaveOccurred())
			Expect(*(flags["count"].(*int))).To(Equal(3))

			flags, app = build()
			_, err = app.Parse([]string{"ginkgo", "--count", "10"})
			Expect(err).ToNot(HaveOccurred())
			Expect(*(flags["count"].(*int))).To(Equal(10))

			_, app = build()
			_, err = app.Parse([]string{"ginkgo", "--count", "abc"})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Validate", func() {
		var valErr func(func(*GenericCommand)) error

		BeforeEach(func() {
			valErr = func(fn func(*GenericCommand)) error {
				d := &GenericCommand{Name: "n", Type: "exec", Description: "help"}
				fn(d)
				return d.Validate(nil)
			}
		})

		It("Should accept supported flag and argument types", func() {
			Expect(valErr(func(d *GenericCommand) {
				d.Flags = []GenericFlag{
					{Name: "s", Description: "h", Type: "string"},
					{Name: "i", Description: "h", Type: "int", Default: "5"},
					{Name: "b", Description: "h", Type: "bool", Default: true},
					{Name: "leg", Description: "h", Bool: true},
					{Name: "e", Description: "h", Enum: []string{"a", "b"}},
					{Name: "ci", Description: "h", Type: " INT "},
				}
				d.Arguments = []GenericArgument{
					{Name: "f", Description: "h", Type: "existing_file"},
				}
			})).To(Succeed())
		})

		It("Should reject an unknown type and list the valid types", func() {
			err := valErr(func(d *GenericCommand) {
				d.Flags = []GenericFlag{{Name: "x", Description: "h", Type: "integr"}}
			})
			Expect(err).To(MatchError(ContainSubstring(`flag "x" has unknown type "integr"`)))
			Expect(err).To(MatchError(ContainSubstring("valid types are: bool, counter")))
		})

		It("Should reject non string or bool defaults with a usable quoting hint", func() {
			// YAML numbers arrive as float64 after YAMLToJSON, and naive formatting
			// would render 1000000 as 1e+06 in the hint.
			err := valErr(func(d *GenericCommand) {
				d.Flags = []GenericFlag{{Name: "c", Description: "h", Type: "int", Default: float64(1000000)}}
			})
			Expect(err).To(MatchError(ContainSubstring("default must be a string or boolean")))
			Expect(err).To(MatchError(ContainSubstring(`default: "1000000"`)))
		})

		It("Should allow enum with a blank or string type", func() {
			Expect(valErr(func(d *GenericCommand) {
				d.Flags = []GenericFlag{
					{Name: "e", Description: "h", Enum: []string{"a", "b"}},
					{Name: "es", Description: "h", Type: "string", Enum: []string{"a", "b"}},
					{Name: "ews", Description: "h", Type: " String ", Enum: []string{"a", "b"}},
				}
				d.Arguments = []GenericArgument{
					{Name: "ea", Description: "h", Type: "string", Enum: []string{"a", "b"}},
				}
			})).To(Succeed())
		})

		It("Should reject combining a non string type with enum", func() {
			for _, typ := range []string{"int", "bool", "float", "existing_file"} {
				Expect(valErr(func(d *GenericCommand) {
					d.Flags = []GenericFlag{{Name: "c", Description: "h", Type: typ, Enum: []string{"1", "2"}}}
				})).To(MatchError(ContainSubstring("sets both type and enum")), "type %q should conflict with enum", typ)
			}
		})

		It("Should reject combining the legacy bool field with a non bool type", func() {
			Expect(valErr(func(d *GenericCommand) {
				d.Flags = []GenericFlag{{Name: "c", Description: "h", Bool: true, Type: "int"}}
			})).To(MatchError(ContainSubstring(`sets both bool and type "int"`)))
		})

		It("Should validate arguments the same as flags", func() {
			err := valErr(func(d *GenericCommand) {
				d.Arguments = []GenericArgument{{Name: "a", Description: "h", Type: "nope", Default: float64(42)}}
			})
			Expect(err).To(MatchError(ContainSubstring(`argument "a" has unknown type "nope"`)))
			Expect(err).To(MatchError(ContainSubstring(`argument "a" default must be a string or boolean`)))
		})
	})

	Describe("input type helpers", func() {
		It("normalizeType lower-cases and trims", func() {
			Expect(normalizeType("  INT ")).To(Equal("int"))
			Expect(normalizeType("Existing_File")).To(Equal("existing_file"))
		})

		DescribeTable("isTrueDefault",
			func(in any, expected bool) {
				Expect(isTrueDefault(in)).To(Equal(expected))
			},
			Entry("bool true", true, true),
			Entry("bool false", false, false),
			Entry("string true", "true", true),
			Entry("string True", "True", true),
			Entry("string 1", "1", true),
			Entry("string false", "false", false),
			Entry("unparsable string", "yes", false),
			Entry("nil", nil, false),
			Entry("number", float64(5), false),
		)

		It("defaultHint formats numbers without scientific notation", func() {
			Expect(defaultHint(float64(1000000))).To(Equal("1000000"))
			Expect(defaultHint(float64(3.5))).To(Equal("3.5"))
			Expect(defaultHint("abc")).To(Equal("abc"))
		})

		It("knownTypeNames is sorted, includes bool and excludes duration", func() {
			names := knownTypeNames()
			Expect(sort.StringsAreSorted(names)).To(BeTrue())
			Expect(names).To(ContainElement("bool"))
			Expect(names).To(ContainElement("string"))
			Expect(names).To(ContainElement("int"))
			Expect(names).ToNot(ContainElement("duration"))
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
			_, err := trans.TransformBytes(context.Background(), []byte(`{"hello":"world"`), nil, nil, &AppBuilder{cfg: map[string]any{}})
			Expect(err).To(MatchError(ErrInvalidTransform))

			trans.Query = ".hello"
			Expect(trans.Validate(nil)).To(Succeed())

			_, err = trans.TransformBytes(context.Background(), []byte(`{`), nil, nil, nil)
			Expect(err).To(MatchError("json input parse error: unexpected end of JSON input"))

			res, err := trans.TransformBytes(context.Background(), []byte(`{"hello":"world"}`), nil, nil, &AppBuilder{cfg: map[string]any{}})
			Expect(err).ToNot(HaveOccurred())
			Expect(string(res)).To(Equal("world\n"))
		})
	})
})
