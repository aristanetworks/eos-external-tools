// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package main

import (
	"log"
	"os"

	"lemurbldr/cmd"
)

func main() {
	log.SetOutput(os.Stdout)
	cmd.SetViperDefaults()
	cmd.Execute()
}
