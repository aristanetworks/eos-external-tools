// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package cmd

import (
	"github.com/spf13/cobra"

	"extbldr/impl"
)

// createSrpmCmd represents the createSrpm command
var createSrpmCmd = &cobra.Command{
	Use:   "createSrpm -p <package> [-s <subpackage>]",
	Short: "Build modified SRPM",
	Long: `A new SRPM is built based on the manifest, spec file and sources specified.
The sources are expected to be already available in <SrcDir>/<package>.
The manifest might specify only a single subpackage per package in the general case.
In situations where multiple SRPMs need to be built in dependency order, the manifest might specify multple subpackages. The [ -s <subpackage> ] can also be used.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		pkg, _ := cmd.Flags().GetString("package")
		subpkg, _ := cmd.Flags().GetString("subpackage")
		err := impl.CreateSrpm(pkg, subpkg)
		return err
	},
}

func init() {
	rootCmd.AddCommand(createSrpmCmd)
}
