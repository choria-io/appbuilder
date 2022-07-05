// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Definition", func() {
	var d *Definition
	BeforeEach(func() {
		d = &Definition{}
	})

	Describe("Validate", func() {
		It("Should require basic values", func() {
			err := d.Validate(nil)
			Expect(err).To(MatchError("invalid definition: application: name is required, description is required, version is required to be a valid semver, author is required, no commands defined"))

			d.Name = "ginkgo"
			d.Version = "1.2.3"
			d.Description = "ginkgo example"
			d.Author = "Ginkgo Tests"
			d.Commands = []json.RawMessage{[]byte("{}")}
			d.HelpTemplate = "x"
			Expect(d.Validate(nil)).To(MatchError("invalid definition: application: help_template must be one of compact, long, default, columns or unset"))

			d.HelpTemplate = "compact"
			Expect(d.Validate(nil)).To(Succeed())
		})
	})
})
