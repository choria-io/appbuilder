// Copyright (c) 2023, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package forms

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBuilder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Builder")
}

var _ = Describe("Forms", func() {
	Describe("Graph", func() {
		It("Should generate correct values", func() {
			root := newObjectEntry(map[string]any{})
			root.addChild(newObjectEntry(map[string]any{"listen": "localhost:-1"}))

			ln, _ := root.addChild(newObjectEntry(map[string]any{"leafnode": nil}))
			ln.addChild(newObjectEntry(map[string]any{"credentials": "/x.cred"}))
			ln.addChild(newObjectEntry(map[string]any{"url": "connect.ngs.global:4222"}))
			urls, _ := ln.addChild(newObjectEntry(map[string]any{"urls": []any{}}))
			urls.addChild(newArrayEntry([]any{"x", "y"}))

			accounts, _ := root.addChild(newObjectEntry(map[string]any{"accounts": nil}))
			users, _ := accounts.addChild(newStringEntry("USERS"))
			uc, _ := users.addChild(newObjectEntry(map[string]any{"users": []any{}}))
			uc.addChild(newArrayEntry([]any{
				map[string]any{"username": "bob", "password": "b0b"},
				map[string]any{"username": "jill", "password": "j1ll"},
			}))

			expected := map[string]any{
				"accounts": map[string]any{
					"USERS": map[string]any{
						"users": []any{
							map[string]any{"password": "b0b", "username": "bob"},
							map[string]any{"password": "j1ll", "username": "jill"},
						},
					},
				},
				"leafnode": map[string]any{
					"credentials": "/x.cred",
					"url":         "connect.ngs.global:4222",
					"urls":        []any{"x", "y"},
				},
				"listen": "localhost:-1",
			}

			_, v := root.combinedValue()

			Expect(v).To(Equal(expected))
		})
	})
})
