// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package cmd

import (
	"code.arista.io/eos/tools/eext/impl"
	"github.com/spf13/cobra"
)

// listUnverifiedSourcescmd represents the list-unverified-sources command
var listUnverifiedSourcescmd = &cobra.Command{
	Use:   "list-unverified-sources",
	Short: "list unverified upstream sources",
	Long: `Checks for the upstream sources within package which don't have a valid signature check i.e, skip-check flag is true
			and generates content hash for the upstream sources.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, _ := cmd.Flags().GetString("repo")
		pkg, _ := cmd.Flags().GetString("package")
		err := impl.ListUnverifiedSources(repo, pkg)
		return err
	},
}

func init() {
	listUnverifiedSourcescmd.Flags().StringP("repo", "r", "", "Repository name (OPTIONAL)")
	listUnverifiedSourcescmd.Flags().StringP("package", "p", "", "specify package name (OPTIONAL)")
	rootCmd.AddCommand(listUnverifiedSourcescmd)
}
