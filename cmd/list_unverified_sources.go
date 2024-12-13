// Copyright (c) 2025 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package cmd

import (
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
	listUnverifiedSourcesCmd.MarkFlagRequired("package")
	rootCmd.AddCommand(listUnverifiedSourcesCmd)
}
