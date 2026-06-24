// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/choria-io/fisk"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Secrets", func() {
	var origRunner func(ctx context.Context, args ...string) ([]byte, error)

	BeforeEach(func() {
		origRunner = onePasswordRunner
	})

	AfterEach(func() {
		onePasswordRunner = origRunner
	})

	validSecret := func() GenericSecret {
		return GenericSecret{Name: "tok", OnePassword: &onePasswordSecret{Item: "i", Field: "f", Vault: "v"}}
	}

	Describe("providerForSecret", func() {
		It("should dispatch on the present sub-key", func() {
			p, err := providerForSecret(validSecret())
			Expect(err).ToNot(HaveOccurred())
			Expect(p).To(BeAssignableToTypeOf(&onePasswordSecret{}))
		})

		It("should error when no provider is configured", func() {
			_, err := providerForSecret(GenericSecret{Name: "x"})
			Expect(err).To(MatchError(ErrInvalidSecret))
			Expect(err.Error()).To(ContainSubstring(`"x" has no provider configured`))
		})
	})

	Describe("GenericSecret.Validate", func() {
		It("should require a name", func() {
			err := GenericSecret{OnePassword: &onePasswordSecret{Item: "i", Field: "f", Vault: "v"}}.Validate()
			Expect(err).To(MatchError(ErrInvalidSecret))
			Expect(err.Error()).To(ContainSubstring("name is required"))
		})

		It("should reject non-identifier names and point at index access", func() {
			err := GenericSecret{Name: "weird-name", OnePassword: &onePasswordSecret{Item: "i", Field: "f", Vault: "v"}}.Validate()
			Expect(err).To(MatchError(ErrInvalidSecret))
			Expect(err.Error()).To(ContainSubstring(`index .Secrets "weird-name"`))
		})

		It("should require item, field and vault for one_password", func() {
			err := GenericSecret{Name: "tok", OnePassword: &onePasswordSecret{}}.Validate()
			Expect(err).To(MatchError(ErrInvalidSecret))
			Expect(err.Error()).To(ContainSubstring("item is required"))
			Expect(err.Error()).To(ContainSubstring("field is required"))
			Expect(err.Error()).To(ContainSubstring("vault is required"))
		})

		It("should reject reference characters op does not allow", func() {
			err := GenericSecret{Name: "tok", OnePassword: &onePasswordSecret{Item: "a/b", Field: "f", Vault: "v"}}.Validate()
			Expect(err).To(MatchError(ErrInvalidSecret))
			Expect(err.Error()).To(ContainSubstring(`item "a/b" may only contain`))
		})

		It("should accept spaces, dots, dashes and underscores in references", func() {
			err := GenericSecret{Name: "tok", OnePassword: &onePasswordSecret{Item: "Demo API Token", Field: "api.credential", Vault: "App_Builder-Demo"}}.Validate()
			Expect(err).To(Succeed())
		})

		It("should accept a well-formed secret", func() {
			Expect(validSecret().Validate()).To(Succeed())
		})
	})

	Describe("GenericCommand secret validation", func() {
		It("should reject duplicate secret names", func() {
			c := &GenericCommand{
				Name: "x", Description: "x", Type: "exec",
				Secrets: []GenericSecret{validSecret(), validSecret()},
			}
			err := c.Validate(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`duplicate secret name "tok"`))
		})
	})

	Describe("onePasswordSecret", func() {
		It("should build the op read reference and args", func() {
			s := &onePasswordSecret{Item: "Demo API Token", Field: "credential", Vault: "AppBuilderDemo"}
			Expect(s.reference()).To(Equal("op://AppBuilderDemo/Demo API Token/credential"))
			Expect(s.readArgs()).To(Equal([]string{"read", "op://AppBuilderDemo/Demo API Token/credential"}))
		})

		It("should add the account flag when set", func() {
			s := &onePasswordSecret{Item: "i", Field: "f", Vault: "v", Account: "my.1password.com"}
			Expect(s.readArgs()).To(Equal([]string{"read", "op://v/i/f", "--account", "my.1password.com"}))
		})

		It("should trim only trailing newlines, preserving other whitespace", func() {
			onePasswordRunner = func(_ context.Context, _ ...string) ([]byte, error) {
				return []byte("  spaced value \n\n"), nil
			}
			v, err := (&onePasswordSecret{Item: "i", Field: "f", Vault: "v"}).Resolve(context.Background())
			Expect(err).ToNot(HaveOccurred())
			Expect(v).To(Equal("  spaced value "))
		})

		It("should surface op stderr but never stdout on failure", func() {
			dir := GinkgoT().TempDir()
			op := filepath.Join(dir, "op")
			script := "#!/bin/sh\necho 'THE-SECRET-VALUE'\necho 'you are not currently signed in' >&2\nexit 1\n"
			Expect(os.WriteFile(op, []byte(script), 0700)).To(Succeed())

			origPath := os.Getenv("PATH")
			Expect(os.Setenv("PATH", dir+string(os.PathListSeparator)+origPath)).To(Succeed())
			DeferCleanup(func() { os.Setenv("PATH", origPath) })

			_, err := onePasswordRunner(context.Background(), "read", "op://v/i/f")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("you are not currently signed in"))
			Expect(err.Error()).ToNot(ContainSubstring("THE-SECRET-VALUE"))
		})
	})

	Describe("resolveSecrets", func() {
		It("should resolve every secret", func() {
			onePasswordRunner = func(_ context.Context, _ ...string) ([]byte, error) {
				return []byte("resolved\n"), nil
			}
			res, err := resolveSecrets(context.Background(), []GenericSecret{validSecret()})
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(Secrets{"tok": "resolved"}))
		})

		It("should name the culprit secret and never include the value on error", func() {
			onePasswordRunner = func(_ context.Context, _ ...string) ([]byte, error) {
				return nil, errors.New("exit status 1: item not found")
			}
			_, err := resolveSecrets(context.Background(), []GenericSecret{
				{Name: "api_token", OnePassword: &onePasswordSecret{Item: "i", Field: "f", Vault: "v"}},
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`secret "api_token"`))
			Expect(err.Error()).To(ContainSubstring("item not found"))
		})

		It("should surface a context timeout", func() {
			onePasswordRunner = func(ctx context.Context, _ ...string) ([]byte, error) {
				<-ctx.Done()
				return nil, ctx.Err()
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
			defer cancel()

			_, err := resolveSecrets(ctx, []GenericSecret{validSecret()})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`secret "tok"`))
		})
	})

	Describe("dryRunSecrets", func() {
		It("should produce self-describing placeholders", func() {
			res := dryRunSecrets([]GenericSecret{{Name: "api_token"}, {Name: "db_pass"}})
			Expect(res).To(Equal(Secrets{"api_token": "<secret:api_token>", "db_pass": "<secret:db_pass>"}))
		})
	})

	Describe("Secrets.Redact", func() {
		It("should mask every secret value in text and skip empties", func() {
			s := Secrets{"a": "value-one", "b": "value-two", "c": ""}
			out := s.Redact("a=value-one b=value-two c= plain")
			Expect(out).To(Equal("a=[REDACTED] b=[REDACTED] c= plain"))
		})

		It("should be a no-op when there are no secrets", func() {
			Expect(Secrets(nil).Redact("nothing to hide")).To(Equal("nothing to hide"))
		})
	})

	Describe("runWrapper resolution", func() {
		var b *AppBuilder

		BeforeEach(func() {
			b = &AppBuilder{ctx: context.Background(), cfg: map[string]any{}, log: NoopLogger{}}
		})

		secretCmd := func() GenericCommand {
			return GenericCommand{Name: "x", Secrets: []GenericSecret{validSecret()}}
		}

		It("should resolve secrets onto the builder before the handler runs", func() {
			onePasswordRunner = func(_ context.Context, _ ...string) ([]byte, error) {
				return []byte("resolved\n"), nil
			}
			var seen Secrets
			action := runWrapper(secretCmd(), map[string]any{}, map[string]any{}, b, func(_ *fisk.ParseContext) error {
				seen = b.Secrets()
				return nil
			})
			Expect(action(nil)).To(Succeed())
			Expect(seen).To(Equal(Secrets{"tok": "resolved"}))
		})

		It("should use placeholders and never call op under BUILDER_DRY_RUN", func() {
			called := false
			onePasswordRunner = func(_ context.Context, _ ...string) ([]byte, error) {
				called = true
				return []byte("should-not-be-used"), nil
			}
			Expect(os.Setenv("BUILDER_DRY_RUN", "1")).To(Succeed())
			DeferCleanup(func() { os.Unsetenv("BUILDER_DRY_RUN") })

			var seen Secrets
			action := runWrapper(secretCmd(), map[string]any{}, map[string]any{}, b, func(_ *fisk.ParseContext) error {
				seen = b.Secrets()
				return nil
			})
			Expect(action(nil)).To(Succeed())
			Expect(called).To(BeFalse())
			Expect(seen).To(Equal(Secrets{"tok": "<secret:tok>"}))
		})

		It("should reset stale secrets for a command that declares none", func() {
			b.secrets = Secrets{"stale": "value"}
			action := runWrapper(GenericCommand{Name: "y"}, map[string]any{}, map[string]any{}, b, func(_ *fisk.ParseContext) error {
				return nil
			})
			Expect(action(nil)).To(Succeed())
			Expect(b.Secrets()).To(BeNil())
		})

		It("should abort and surface resolution errors before the handler", func() {
			onePasswordRunner = func(_ context.Context, _ ...string) ([]byte, error) {
				return nil, errors.New("not signed in")
			}
			handlerRan := false
			action := runWrapper(secretCmd(), map[string]any{}, map[string]any{}, b, func(_ *fisk.ParseContext) error {
				handlerRan = true
				return nil
			})
			err := action(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`secret "tok"`))
			Expect(handlerRan).To(BeFalse())
		})
	})
})
