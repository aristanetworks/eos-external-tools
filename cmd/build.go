// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package cmd

import (
	"github.com/spf13/cobra"

	"code.arista.io/eos/tools/eext/impl"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Run create-srpm and mock in order.",
	Long: `Builds SRPMs from manifest, and then builds the RPMs.
	The results are made available in <DestDir>/SRPMS/<package> and <DestDir>/RPMS/<package>.
	The manifest might specify only a single package(SRPM) per repo in the general case.
	In situations where multiple packages need to be built in dependency order, the manifest might specify multple packages. The [ -p <package> ] can also be used to just build a specific package.
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, _ := cmd.Flags().GetString("repo")
		pkg, _ := cmd.Flags().GetString("package")
		extraCreateSrpmArgs := impl.CreateSrpmExtraCmdlineArgs{
			SkipBuildPrep: commonArgs.skipBuildPrep,
		}
		extraMockArgs := impl.MockExtraCmdlineArgs{
			NoCheck: commonArgs.noCheck}

		if err := impl.Build(repo, pkg, defaultArch(),
			extraCreateSrpmArgs, extraMockArgs); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	buildCmd.Flags().StringVarP(&repoName, "repo", "r", "", "Repository name (OPTIONAL)")
	buildCmd.Flags().StringVarP(&pkgName, "package", "p", "", "package name (OPTIONAL)")
	buildCmd.Flags().BoolVar(&commonArgs.noCheck, "nocheck", false, "Pass --nocheck to rpmbuild (OPTIONAL)")
	buildCmd.Flags().BoolVar(&commonArgs.skipBuildPrep, "skip-build-prep", false, "Skips build-prep during createSrpm for cases where build-prep requires dependencies not in container (OPTIONAL)")
	rootCmd.AddCommand(buildCmd)
}
