// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"code.arista.io/eos/tools/eext/executor"
)

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

	rootCmd.PersistentFlags().String("config", "", "config file (default is eext-viper.yaml in /etc or $HOME/.config)")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Quiet terminal output (default is false)")
	rootCmd.PersistentFlags().BoolP("dry-run", "d", false, "Instead of running the commands, print what would be run")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile, err := rootCmd.Flags().GetString("config"); err != nil {
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

// Select appropriate strategy for running commands based on the dry-run flag
func selectExecutor() executor.Executor {
	// In the executors ever grow to become expensive on unweildy to be created
	// multiple times in one program session, sync.Once could be used
	if dryRun, _ := rootCmd.PersistentFlags().GetBool("dry-run"); dryRun {
		return &executor.DryRunExecutor{}
	} else {
		suppress, _ := rootCmd.PersistentFlags().GetBool("quiet")
		return &executor.OsExecutor{Suppress: suppress}
	}
}
