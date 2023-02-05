// Copyright (c) 2022, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"bytes"
	"os"
)

func fileExist(path string) bool {
	if path == "" {
		return false
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

func isJSON(data []byte) bool {
	d := bytes.TrimSpace(data)

	if bytes.HasPrefix(d, []byte("{")) && bytes.HasSuffix(d, []byte("}")) {
		return true
	}

	if bytes.HasPrefix(d, []byte("[")) && bytes.HasSuffix(d, []byte("]")) {
		return true
	}

	return false
}
