// Copyright (c) 2023, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/choria-io/fisk"
	"github.com/expr-lang/expr"
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

// SurveyValidator is a validator for github.com/AlecAivazis/survey
func SurveyValidator(validation string, required bool) func(any) error {
	return func(v any) error {
		val, ok := v.(string)
		if !ok {
			return fmt.Errorf("unsupported validation type")
		}

		if !required && len(val) == 0 {
			return nil
		}

		ok, err := Validate(v, validation)
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

	opts := []expr.Option{
		expr.Env(env), expr.AsBool(),
	}
	opts = append(opts, ShellSafeValidator()...)
	opts = append(opts, IPv4Validator()...)
	opts = append(opts, IPv6Validator()...)
	opts = append(opts, IPvValidator()...)
	opts = append(opts, IntValidator()...)
	opts = append(opts, FloatValidator()...)

	program, err := expr.Compile(validation, opts...)
	if err != nil {
		return false, err
	}

	output, err := expr.Run(program, env)
	if err != nil {
		return false, err
	}

	return output.(bool), nil
}

func FloatValidator() []expr.Option {
	f := func(params ...any) (any, error) {
		_, err := strconv.ParseFloat(params[0].(string), 64)
		return err == nil, nil
	}

	return []expr.Option{
		expr.Function("isFloat", f, new(func(string) (bool, error))),
		expr.Function("is_float", f, new(func(string) (bool, error))),
	}
}

func IntValidator() []expr.Option {
	f := func(params ...any) (any, error) {
		_, err := strconv.Atoi(params[0].(string))
		return err == nil, nil
	}

	return []expr.Option{
		expr.Function("isInt", f, new(func(string) (bool, error))),
		expr.Function("is_int", f, new(func(string) (bool, error))),
	}
}

func IPvValidator() []expr.Option {
	f := func(params ...any) (any, error) {
		val := params[0].(string)
		ip := net.ParseIP(val)

		if ip == nil {
			return false, fmt.Errorf("%s is not an IP address", val)
		}

		return true, nil
	}

	return []expr.Option{
		expr.Function("isIP", f, new(func(string) (bool, error))),
		expr.Function("is_ip", f, new(func(string) (bool, error))),
	}
}

func IPv4Validator() []expr.Option {
	f := func(params ...any) (any, error) {
		val := params[0].(string)
		ip := net.ParseIP(val).To4()

		if ip == nil {
			return false, fmt.Errorf("%s is not an IPv4 address", val)
		}

		return true, nil
	}

	return []expr.Option{
		expr.Function("isIPv4", f, new(func(string) (bool, error))),
		expr.Function("is_ipv4", f, new(func(string) (bool, error))),
	}
}

func IPv6Validator() []expr.Option {
	f := func(params ...any) (any, error) {
		val := params[0].(string)
		ip := net.ParseIP(val)

		if ip == nil {
			return false, fmt.Errorf("%s is not an IPv6 address", val)
		}

		if ip.To4() != nil {
			return false, fmt.Errorf("%s is not an IPv6 address", val)
		}

		return true, nil
	}
	return []expr.Option{
		expr.Function("isIPv6", f, new(func(string) (bool, error))),
		expr.Function("is_ipv6", f, new(func(string) (bool, error))),
	}
}

func ShellSafeValidator() []expr.Option {
	f := func(params ...any) (any, error) {
		val := strings.TrimSpace(params[0].(string))
		badchars := []string{"`", "$", ";", "|", "&&", ">", "<"}

		for _, c := range badchars {
			if strings.Contains(val, c) {
				return false, fmt.Errorf("may not contain '%s'", c)
			}
		}

		return true, nil
	}

	return []expr.Option{
		expr.Function("isShellSafe", f, new(func(string) (bool, error))),
		expr.Function("is_shellsafe", f, new(func(string) (bool, error))),
	}
}
