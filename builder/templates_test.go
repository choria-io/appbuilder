// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"encoding/json"
	"text/template"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Templates", func() {
	var (
		b      *AppBuilder
		argVal string
	)

	BeforeEach(func() {
		argVal = "world"
		b = &AppBuilder{
			cfg:     map[string]any{"environment": "production"},
			secrets: Secrets{"api_token": "s3cr3t-value"},
		}
	})

	args := func() map[string]any {
		return map[string]any{"name": &argVal}
	}

	Describe("RenderTemplate", func() {
		It("should expose arguments and config", func() {
			out, err := b.RenderTemplate(`{{ .Arguments.name }}-{{ .Config.environment }}`, args(), nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(Equal("world-production"))
		})

		It("should expose .Input via WithInput", func() {
			out, err := b.RenderTemplate(`{{ .Input }}`, nil, nil, WithInput("the-input"))
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(Equal("the-input"))
		})

		It("should always provide the builder directory functions", func() {
			b.userWorkingDir = "/work/dir"
			out, err := b.RenderTemplate(`{{ UserWorkingDir }}`, nil, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(Equal("/work/dir"))
		})

		It("should not enable sprig functions by default", func() {
			_, err := b.RenderTemplate(`{{ "x" | upper }}`, nil, nil)
			Expect(err).To(HaveOccurred())
		})

		It("should enable sprig functions via WithSprig", func() {
			out, err := b.RenderTemplate(`{{ "x" | upper }}`, nil, nil, WithSprig())
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(Equal("X"))
		})

		It("should add caller functions via WithFuncs", func() {
			funcs := template.FuncMap{"greet": func() string { return "hi" }}
			out, err := b.RenderTemplate(`{{ greet }}`, nil, nil, WithFuncs(funcs))
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(Equal("hi"))
		})

		It("should resolve explicit secret references", func() {
			out, err := b.RenderTemplate(`{{ .Secrets.api_token }}`, nil, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(Equal("s3cr3t-value"))
		})
	})

	Describe("NewTemplateState", func() {
		It("should fill config and secrets from the builder", func() {
			state := b.NewTemplateState(args(), nil)
			Expect(state.Config).To(Equal(map[string]any{"environment": "production"}))
			Expect(state.Secrets).To(Equal(Secrets{"api_token": "s3cr3t-value"}))
			Expect(state.Arguments).To(Equal(map[string]any{"name": "world"}))
		})

		It("should set input via WithInput", func() {
			state := b.NewTemplateState(nil, nil, WithInput(42))
			Expect(state.Input).To(Equal(42))
		})
	})

	Describe("Secrets redaction", func() {
		It("should redact whole-state text dumps", func() {
			out, err := b.RenderTemplate(`{{ . }}`, args(), nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(ContainSubstring(secretRedaction))
			Expect(out).ToNot(ContainSubstring("s3cr3t-value"))
		})

		It("should redact whole-state JSON dumps", func() {
			out, err := b.RenderTemplate(`{{ toJson . }}`, args(), nil, WithSprig())
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(ContainSubstring(secretRedaction))
			Expect(out).ToNot(ContainSubstring("s3cr3t-value"))
		})

		It("should still resolve explicit secret indexes", func() {
			out, err := b.RenderTemplate(`{{ .Secrets.api_token }}`, nil, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(Equal("s3cr3t-value"))
		})

		It("should redact via String and MarshalJSON only when populated", func() {
			Expect(Secrets{"x": "y"}.String()).To(Equal(secretRedaction))
			Expect(Secrets(nil).String()).To(Equal(""))

			j, err := json.Marshal(Secrets{"x": "y"})
			Expect(err).ToNot(HaveOccurred())
			Expect(string(j)).To(Equal(`"[REDACTED]"`))

			j, err = json.Marshal(Secrets(nil))
			Expect(err).ToNot(HaveOccurred())
			Expect(string(j)).To(Equal(`""`))
		})
	})
})
