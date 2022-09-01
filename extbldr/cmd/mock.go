// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package cmd

import (
	"github.com/spf13/cobra"

	"extbldr/impl"
)

var arch string

var mockCmd = &cobra.Command{
	Use:   "mock -t <Arch> -p <Package>",
	Short: "Use mock to build the RPMS from the SRPMS built previously.",
	Long:  `Use mock to build The RPMS for the specified architecture from the specified SRPM package in <SrcDir>/<package>/rpmbuild/SRPMS and output placed inside <WorkingDir>/<package>/RPMS.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		pkg, _ := cmd.Flags().GetString("package")
		subpkg, _ := cmd.Flags().GetString("subpackage")
		err := impl.Mock(arch, pkg, subpkg)
		return err
	},
}

func init() {
	mockCmd.Flags().StringVarP(&arch, "target", "t", "", "target architecture for the RPM")
	rootCmd.AddCommand(mockCmd)
}
