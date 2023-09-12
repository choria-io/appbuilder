// Copyright (c) 2023, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBuilder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Validator")
}

var _ = Describe("Validator", func() {
	Describe("is_ip", func() {
		It("Should validate correctly", func() {
			ok, err := Validate("1.1.1.1", "is_ip(value)")
			Expect(err).ToNot(HaveOccurred())
			Expect(ok).To(BeTrue())

			ok, err = Validate("2a00:1450:4002:405::20", "is_ip(value)")
			Expect(err).ToNot(HaveOccurred())
			Expect(ok).To(BeTrue())

			ok, err = Validate("bob", "is_ip(value)")
			Expect(err.Error()).To(ContainSubstring("bob is not an IP address"))
			Expect(ok).To(BeFalse())
		})
	})

	Describe("is_ipv4", func() {
		It("Should validate correctly", func() {
			ok, err := Validate("1.1.1.1", "is_ipv4(value)")
			Expect(err).ToNot(HaveOccurred())
			Expect(ok).To(BeTrue())

			ok, err = Validate("2a00:1450:4002:405::20", "is_ipv4(value)")
			Expect(err.Error()).To(ContainSubstring("2a00:1450:4002:405::20 is not an IPv4 address"))
			Expect(ok).To(BeFalse())
		})
	})

	Describe("is_ipv6", func() {
		It("Should validate correctly", func() {
			ok, err := Validate("2a00:1450:4002:405::20", "is_ipv6(value)")
			Expect(err).ToNot(HaveOccurred())
			Expect(ok).To(BeTrue())

			ok, err = Validate("1.1.1.1", "is_ipv6(value)")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("1.1.1.1 is not an IPv6 address"))
			Expect(ok).To(BeFalse())
		})
	})

	Describe("shellsafe", func() {
		It("Should match bad strings", func() {
			badchars := []string{"`", "$", ";", "|", "&&", ">", "<"}

			for _, c := range badchars {
				ok, err := Validate(fmt.Sprintf("thing%sthing", c), "is_shellsafe(value)")
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("may not contain '%s'", c)))
				Expect(ok).To(BeFalse())
			}
		})

		It("Should allow good things", func() {
			Expect(Validate("ok", "is_shellsafe(value)")).To(BeTrue())
			Expect(Validate("", "is_shellsafe(value)")).To(BeTrue())
			Expect(Validate("ok ok ok", "is_shellsafe(value)")).To(BeTrue())
		})
	})
})
