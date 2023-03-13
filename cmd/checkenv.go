// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"lemurbldr/util"
)

var checkenvCmd = &cobra.Command{
	Use:   "checkenv",
	Short: "Checks the environment to see if a build can be done!",
	RunE: func(cmd *cobra.Command, args []string) error {
		srcDir := viper.GetString("SrcDir")
		workingDir := viper.GetString("WorkingDir")
		destDir := viper.GetString("DestDir")

		var aggError string
		failed := false
		if err := util.CheckPath(srcDir, true, false); err != nil {
			aggError += fmt.Sprintf("\ntrouble with SrcDir: %s", err)
			failed = true
		}

		if err := util.CheckPath(workingDir, true, true); err != nil {
			aggError += fmt.Sprintf("\ntrouble with WorkingDir: %s", err)
			failed = true
		}

		if err := util.CheckPath(destDir, true, true); err != nil {
			aggError += fmt.Sprintf("\ntrouble with DestDir: %s", err)
			failed = true
		}

		if failed {
			return fmt.Errorf("Environment check failed!%s", aggError)
		}
		fmt.Println("Environment looks OK!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkenvCmd)
}
