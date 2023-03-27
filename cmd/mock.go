// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package cmd

import (
	"github.com/spf13/cobra"

	"code.arista.io/eos/tools/eext/impl"
	"code.arista.io/eos/tools/eext/util"
)

var noCheck bool
var onlyCreateCfg bool

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
		extraArgs := impl.MockExtraCmdlineArgs{
			NoCheck:       noCheck,
			OnlyCreateCfg: onlyCreateCfg}
		if err := util.MaybeSetDefaultArch(); err != nil {
			return err
		}
		if err := impl.Mock(repo, pkg, util.GlobalVar.Arch, extraArgs); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	mockCmd.Flags().StringVarP(&repoName, "repo", "r", "", "Repository name (OPTIONAL)")
	mockCmd.Flags().StringVarP(&pkgName, "package", "p", "", "package name (OPTIONAL)")
	mockCmd.Flags().StringVarP(&util.GlobalVar.Arch, "target", "t", "", "target architecture for the RPM (OPTIONAL)")
	mockCmd.Flags().BoolVar(&onlyCreateCfg, "only-create-cfg", false, "Just create mock configuration, don't run mock (OPTIONAL)")
	mockCmd.Flags().BoolVar(&noCheck, "nocheck", false, "Pass --nocheck to rpmbuild (OPTIONAL)")
	rootCmd.AddCommand(mockCmd)
}
