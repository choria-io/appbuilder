// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Utils", func() {
	Describe("fileExist", func() {
		It("Should detect files correctly", func() {
			Expect(fileExist("/nonexisting")).To(BeFalse())
			Expect(fileExist("/")).To(BeTrue())
		})
	})
})
