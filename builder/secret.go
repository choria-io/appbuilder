// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"context"
	"errors"
	"fmt"
	"regexp"
)

// ErrInvalidSecret indicates a secret is not well-formed or has no usable provider
var ErrInvalidSecret = errors.New("invalid secret")

// secretNamePattern restricts secret names to template-safe identifiers so that
// {{ .Secrets.<name> }} dot-access always works; non-identifier names are rejected.
var secretNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// GenericSecret is a value resolved at command-invocation time from an external store and
// exposed to templates as {{ .Secrets.<name> }}. Exactly one provider sub-key must be set,
// dispatch mirrors transformerForQuery.
type GenericSecret struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	OnePassword *onePasswordSecret `json:"one_password,omitempty"`
}

// secretProvider resolves a single secret value from an external store
type secretProvider interface {
	// Resolve fetches the secret value, it must never return the value inside an error
	Resolve(ctx context.Context) (string, error)
	// Validate ensures the provider configuration is well-formed
	Validate() error
}

// providerForSecret returns the configured provider based on which sub-key is present
func providerForSecret(s GenericSecret) (secretProvider, error) {
	switch {
	case s.OnePassword != nil:
		return s.OnePassword, nil

	default:
		return nil, fmt.Errorf("%w: %q has no provider configured", ErrInvalidSecret, s.Name)
	}
}

// Validate ensures the secret has a template-safe name and a valid provider
func (s GenericSecret) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidSecret)
	}

	if !secretNamePattern.MatchString(s.Name) {
		return fmt.Errorf(`%w: name %q must match %s, reference non-identifier names with {{ index .Secrets "%s" }}`, ErrInvalidSecret, s.Name, secretNamePattern.String(), s.Name)
	}

	provider, err := providerForSecret(s)
	if err != nil {
		return err
	}

	return provider.Validate()
}

// resolveSecrets resolves every secret, aborting on the first failure. The returned error names
// the culprit secret and never contains the secret value.
func resolveSecrets(ctx context.Context, secrets []GenericSecret) (Secrets, error) {
	res := Secrets{}

	for _, s := range secrets {
		provider, err := providerForSecret(s)
		if err != nil {
			return nil, err
		}

		v, err := provider.Resolve(ctx)
		if err != nil {
			return nil, fmt.Errorf("could not resolve secret %q: %w", s.Name, err)
		}

		res[s.Name] = v
	}

	return res, nil
}

// dryRunSecrets returns self-describing placeholders so BUILDER_DRY_RUN never contacts a store
func dryRunSecrets(secrets []GenericSecret) Secrets {
	res := Secrets{}

	for _, s := range secrets {
		res[s.Name] = fmt.Sprintf("<secret:%s>", s.Name)
	}

	return res
}
