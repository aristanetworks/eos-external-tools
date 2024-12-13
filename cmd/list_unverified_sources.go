// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package cmd

import (
	"fmt"

	"code.arista.io/eos/tools/eext/impl"
	"github.com/spf13/cobra"
)

// listUnverifiedSourcesCmd represents the list-unverified-sources command
var listUnverifiedSourcesCmd = &cobra.Command{
	Use:   "list-unverified-sources",
	Short: "list unverified upstream sources",
	Long: `Checks for the upstream sources within package which don't 
have a valid signature check return prints the upstreamSrc
to stdout.`,
	Args: cobra.ExactArgs(0),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		pkg, _ := cmd.Flags().GetString("package")
		if pkg == "" {
			return fmt.Errorf("package not specified. Use : eext list-unverified-sources -p <package>")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, _ := cmd.Flags().GetString("repo")
		pkg, _ := cmd.Flags().GetString("package")
		err := impl.ListUnverifiedSources(repo, pkg)
		return err
	},
}

func init() {
	listUnverifiedSourcesCmd.Flags().StringP("repo", "r", "", "Repository name (OPTIONAL)")
	listUnverifiedSourcesCmd.Flags().StringP("package", "p", "", "specify package name (REQUIRED)")
	rootCmd.AddCommand(listUnverifiedSourcesCmd)
}
