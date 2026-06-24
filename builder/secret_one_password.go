// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// ErrOnePasswordNotFound indicates the 1Password CLI (op) is not installed or not in PATH
var ErrOnePasswordNotFound = errors.New("1Password CLI (op) not found in PATH")

// onePasswordResolveTimeout bounds a single op invocation so a stuck CLI cannot hang the command
const onePasswordResolveTimeout = 30 * time.Second

// onePasswordRefChars matches the characters op permits in a secret reference component:
// letters, digits, '-', '_', '.' and whitespace. Anything else (notably '/') would corrupt
// the op://vault/item/field reference, so we reject it before invoking op. Components with
// out-of-spec characters can only be addressed by their item/vault ID instead.
var onePasswordRefChars = regexp.MustCompile(`^[A-Za-z0-9_.\- ]+$`)

// onePasswordRunner runs the op CLI and returns its stdout. It is a package var so tests can
// override it and never invoke the real op binary.
var onePasswordRunner = func(ctx context.Context, args ...string) ([]byte, error) {
	path, err := exec.LookPath("op")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrOnePasswordNotFound, err)
	}

	cmd := exec.CommandContext(ctx, path, args...)
	// No stdin so a TTY-less op fails fast rather than blocking on an interactive prompt
	cmd.Stdin = nil

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			// Surface op's own stderr (which describes not-authenticated / item-not-found etc),
			// never stdout, so the secret value can never leak into an error.
			msg := strings.TrimSpace(stderr.String())
			if msg == "" {
				return nil, err
			}

			return nil, fmt.Errorf("%w: %s", err, msg)
		}

		return nil, err
	}

	return stdout.Bytes(), nil
}

// onePasswordSecret resolves a single field from a 1Password item using `op read`
type onePasswordSecret struct {
	// Item is the item name or ID holding the secret
	Item string `json:"item"`
	// Field is the field within the item to read
	Field string `json:"field"`
	// Vault is the vault holding the item, required because op read references must include it
	Vault string `json:"vault"`
	// Account optionally selects a specific 1Password account
	Account string `json:"account"`
}

// Validate ensures item, field and vault are set and contain only characters op permits in a
// secret reference. Vault is mandatory: op read rejects any reference missing vault/item/field.
func (s *onePasswordSecret) Validate() error {
	var errs []string

	for _, c := range []struct {
		name  string
		value string
	}{
		{"item", s.Item},
		{"field", s.Field},
		{"vault", s.Vault},
	} {
		if c.value == "" {
			errs = append(errs, fmt.Sprintf("%s is required", c.name))
			continue
		}

		if !onePasswordRefChars.MatchString(c.value) {
			errs = append(errs, fmt.Sprintf("%s %q may only contain letters, digits, spaces, '-', '_' and '.'", c.name, c.value))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%w: one_password: %s", ErrInvalidSecret, strings.Join(errs, ", "))
	}

	return nil
}

// reference builds the op:// secret reference, e.g. op://vault/item/field
func (s *onePasswordSecret) reference() string {
	return fmt.Sprintf("op://%s/%s/%s", s.Vault, s.Item, s.Field)
}

// readArgs builds the op command line for resolving this secret
func (s *onePasswordSecret) readArgs() []string {
	args := []string{"read", s.reference()}
	if s.Account != "" {
		args = append(args, "--account", s.Account)
	}

	return args
}

// Resolve reads the secret value via `op read`
func (s *onePasswordSecret) Resolve(ctx context.Context) (string, error) {
	err := s.Validate()
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(ctx, onePasswordResolveTimeout)
	defer cancel()

	out, err := onePasswordRunner(ctx, s.readArgs()...)
	if err != nil {
		return "", err
	}

	// Trim only trailing newlines that op appends, never TrimSpace, so PEM blocks or values with
	// leading/internal whitespace survive intact.
	return strings.TrimRight(string(out), "\n"), nil
}
