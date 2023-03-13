// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"lemurbldr/impl"
)

var checkenvCmd = &cobra.Command{
	Use:   "checkenv",
	Short: "Checks the environment to see if a build can be done!",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		if err = impl.CheckEnv(); err == nil {
			fmt.Println("Environment looks OK!")
		}
		return err
	},
}

func init() {
	rootCmd.AddCommand(checkenvCmd)
}
