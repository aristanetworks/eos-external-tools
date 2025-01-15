// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package cmd

import (
	"github.com/spf13/cobra"

	"code.arista.io/eos/tools/eext/impl"
)

// createSrpmCmd represents the createSrpm command
var createSrpmCmd = &cobra.Command{
	Use:   "create-srpm",
	Short: "Build modified SRPM",
	Long: `A new SRPM is built based on the manifest, spec file and sources specified.
The sources are expected to be already available in <SrcDir>/<repo> if --repo <repo> is specified,
otherwise sources are expected to be in current working directory.
The results are made available in <DestDir>/SRPMS/<package>
The manifest might specify only a single package per repo in the general case.
In situations where multiple SRPMs need to be built in dependency order, the manifest might specify multple packages. The [ -p <package> ] can also be used to just build a specific package.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, _ := cmd.Flags().GetString("repo")
		pkg, _ := cmd.Flags().GetString("package")
		doBuildPrep, _ := cmd.Flags().GetBool("do-build-prep")
		extraArgs := impl.CreateSrpmExtraCmdlineArgs{
			DoBuildPrep: doBuildPrep,
		}
		err := impl.CreateSrpm(repo, pkg, extraArgs, selectExecutor())
		return err
	},
}

func init() {
	createSrpmCmd.Flags().StringP("repo", "r", "", "Repository name (OPTIONAL)")
	createSrpmCmd.Flags().StringP("package", "p", "", "package name (OPTIONAL)")
	createSrpmCmd.Flags().Bool("skip-build-prep", false, "DEPRECATED. No-op")
	createSrpmCmd.Flags().MarkHidden("skip-build-prep")
	createSrpmCmd.Flags().Bool("do-build-prep", false, "Runs build-prep on the created SRPM to make sure patches apply cleanly (OPTIONAL)")
	rootCmd.AddCommand(createSrpmCmd)
}
