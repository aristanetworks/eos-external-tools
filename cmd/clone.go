// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"lemurbldr/impl"
	"lemurbldr/util"
)

var force bool

// cloneCmd represents the clone command
var cloneCmd = &cobra.Command{
	Use:   "clone -r <repo> <URL>",
	Short: "git clone the repository for the modified external package",
	Long: `The git repository specified by the URL is cloned to a local directory.
The local directory is <BASE_PATH>/<repo>.
<BASE_PATH> is specified by the SrcDir configuration or the LEMURBLDR_SRCDIR env var.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("Requires exactly one argument (URL)")
		}
		arg := args[0]
		err := util.RunSystemCmd("git", "ls-remote", arg)
		if err != nil {
			return fmt.Errorf("Invalid URL for git repo: %s. git ls-remote errored out with %s", arg, err)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		pkg, _ := cmd.Flags().GetString("repo")
		basePath := viper.GetString("SrcDir")
		err := impl.Clone(args[0], basePath, pkg, force)
		return err
	},
}

func init() {
	cloneCmd.Flags().BoolVarP(&force, "force", "f", false, "Clone again if the local directory already exists")
	cloneCmd.Flags().StringVarP(&repoName, "repo", "r", "", "Repository name (REQUIRED)")
	cloneCmd.Flags().StringVarP(&pkgName, "package", "p", "", "package name (OPTIONAL)")
	cloneCmd.MarkFlagRequired("repo")
	rootCmd.AddCommand(cloneCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cloneCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cloneCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
