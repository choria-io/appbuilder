// Copyright (c) 2023, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"

	"github.com/choria-io/appbuilder/forms"
)

func main() {
	res, err := forms.ProcessFile("form.yaml", nil)
	if err != nil {
		panic(err)
	}

	j, _ := json.MarshalIndent(res, "", "   ")
	fmt.Println(string(j))
}
