// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package cmd

import (
	"github.com/spf13/cobra"

	"lemurbldr/impl"
)

var arch string

var mockCmd = &cobra.Command{
	Use:   "mock -t <Arch> -r <repo>",
	Short: "Use mock to build the RPMS from the SRPMS built previously.",
	Long:  `Use mock to build The RPMS for the specified architecture from the specified SRPM repo in <SrcDir>/<repo>/rpmbuild/SRPMS and output placed inside <WorkingDir>/<repo>/RPMS.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, _ := cmd.Flags().GetString("repo")
		pkg, _ := cmd.Flags().GetString("package")
		err := impl.Mock(arch, repo, pkg)
		return err
	},
}

func init() {
	mockCmd.Flags().StringVarP(&repoName, "repo", "r", "", "Repository name (REQUIRED)")
	mockCmd.Flags().StringVarP(&pkgName, "package", "p", "", "package name (OPTIONAL)")
	mockCmd.Flags().StringVarP(&arch, "target", "t", "", "target architecture for the RPM (REQUIRED)")
	mockCmd.MarkFlagRequired("repo")
	mockCmd.MarkFlagRequired("target")
	rootCmd.AddCommand(mockCmd)
}
