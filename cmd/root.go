// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"code.arista.io/eos/tools/eext/util"
)

var cfgFile string
var repoName string
var pkgName string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:          "eext",
	SilenceUsage: true,
	Short:        "Build external packages for EOS",
	Long: `Modified external packages for EOS Abuild can be specified using a git repository.
The repository would have a manifest which specifies the upstream SRPM/tarball,
any Arista specific patches and the modified spec file. The patches and the spec file are also
stored in the repository.
The upstream SRPM can be checked into the repository or could be stored in another repository,
with the link specified in the mnaifest.

This tool builds the Arista modified SRPM and RPMs from this repository.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is eext-viper.yaml in /etc or $HOME/.config)")
	rootCmd.PersistentFlags().BoolVarP(&(util.GlobalVar.Quiet), "quiet", "q", false, "Quiet terminal output (default is false)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigType("yaml")
		viper.SetConfigName("eext-viper")
		viper.AddConfigPath("$HOME/.config")
		viper.AddConfigPath("/etc/eext")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
