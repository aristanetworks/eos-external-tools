// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package cmd

import (
	"github.com/spf13/cobra"

	"code.arista.io/eos/tools/eext/impl"
)

var mockCmd = &cobra.Command{
	Use:   "mock",
	Short: "Build RPMs from SRPM.",
	Long: `RPMS are built from the SRPM built by createSrpm. It is expected to find the corresponding SRPMS in <DestDir>/SRPMS/<package>.
	The results are made available in <DestDir>/RPMS/<package>.
	The manifest might specify only a single package(SRPM) per repo in the general case.
	In situations where multiple packages need to be built in dependency order, the manifest might specify multple packages. The [ -p <package> ] can also be used to just build a specific package.
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, _ := cmd.Flags().GetString("repo")
		pkg, _ := cmd.Flags().GetString("package")
		target, _ := cmd.Flags().GetString("target")
		onlyCreateCfg, _ := cmd.Flags().GetBool("only-create-cfg")
		noCheck, _ := cmd.Flags().GetBool("nocheck")
		extraArgs := impl.MockExtraCmdlineArgs{
			NoCheck:       noCheck,
			OnlyCreateCfg: onlyCreateCfg,
		}
		if err := impl.Mock(repo, pkg, target, extraArgs); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	mockCmd.Flags().StringP("repo", "r", "", "Repository name (OPTIONAL)")
	mockCmd.Flags().StringP("package", "p", "", "package name (OPTIONAL)")
	mockCmd.Flags().StringP("target", "t", defaultArch(), "target architecture for the rpmbuild (OPTIONAL)")
	mockCmd.Flags().Bool("only-create-cfg", false, "Just create mock configuration, don't run mock (OPTIONAL)")
	mockCmd.Flags().Bool("nocheck", false, "Pass --nocheck to rpmbuild (OPTIONAL)")
	rootCmd.AddCommand(mockCmd)
}
