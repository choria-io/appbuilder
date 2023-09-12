// Copyright (c) 2023, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"fmt"
	"net"
	"strings"

	"github.com/antonmedv/expr"
	"github.com/choria-io/fisk"
)

// FiskValidator is a fisk.OptionValidator that compatible with Validator() on arguments and flags
func FiskValidator(validation string) fisk.OptionValidator {
	return func(value string) error {
		ok, err := Validate(value, validation)
		if err != nil {
			return fmt.Errorf("validation using %q failed: %w", validation, err)
		}

		if !ok {
			return fmt.Errorf("validation using %q did not pass", validation)
		}

		return nil
	}
}

// Validate validates value using the expr expression validation
func Validate(value any, validation string) (bool, error) {
	var env any

	vs, ok := value.(string)
	if ok {
		env = map[string]any{
			"value": vs,
			"Value": vs,
		}
	} else {
		env = value
	}

	program, err := expr.Compile(validation, expr.Env(env), expr.AsBool(),
		ShellSafeValidator(),
		IPv4Validator(),
		IPv6Validator(),
		IPvValidator(),
	)
	if err != nil {
		return false, err
	}

	output, err := expr.Run(program, env)
	if err != nil {
		return false, err
	}

	return output.(bool), nil
}

func IPvValidator() expr.Option {
	return expr.Function(
		"is_ip",
		func(params ...any) (any, error) {
			val := params[0].(string)
			ip := net.ParseIP(val)

			if ip == nil {
				return false, fmt.Errorf("%s is not an IP address", val)
			}

			return true, nil
		},
		new(func(string) (bool, error)))
}

func IPv4Validator() expr.Option {
	return expr.Function(
		"is_ipv4",
		func(params ...any) (any, error) {
			val := params[0].(string)
			ip := net.ParseIP(val).To4()

			if ip == nil {
				return false, fmt.Errorf("%s is not an IPv4 address", val)
			}

			return true, nil
		},
		new(func(string) (bool, error)))
}

func IPv6Validator() expr.Option {
	return expr.Function(
		"is_ipv6",
		func(params ...any) (any, error) {
			val := params[0].(string)
			ip := net.ParseIP(val)

			if ip == nil {
				return false, fmt.Errorf("%s is not an IPv6 address", val)
			}

			if ip.To4() != nil {
				return false, fmt.Errorf("%s is not an IPv6 address", val)
			}

			return true, nil
		},
		new(func(string) (bool, error)))
}

func ShellSafeValidator() expr.Option {
	return expr.Function(
		"is_shellsafe",
		func(params ...any) (any, error) {
			val := strings.TrimSpace(params[0].(string))
			badchars := []string{"`", "$", ";", "|", "&&", ">", "<"}

			for _, c := range badchars {
				if strings.Contains(val, c) {
					return false, fmt.Errorf("may not contain '%s'", c)
				}
			}

			return true, nil
		},
		new(func(string) (bool, error)))
}
