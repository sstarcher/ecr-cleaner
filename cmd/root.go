/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/sstarcher/ecr-cleaner/cleaner"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string
var region string
var days int
var dryRun bool
var removeSemver bool
var debug bool
var force bool
var repo string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ecr-cleaner",
	Short: "Cleanup images from AWS ECR",
	Long:  `An opinionated tool to cleanup ECR images`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.SetLevel(log.InfoLevel)
		if debug {
			log.SetLevel(log.DebugLevel)
		}

		cleaner, err := cleaner.New(&region)
		if err != nil {
			return err
		}

		return cleaner.Prune(time.Hour*24*time.Duration(days), !removeSemver, dryRun, force, repo)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ecr-cleaner.yaml)")

	rootCmd.Flags().StringVar(&region, "region", "", "ECR Region")
	rootCmd.Flags().IntVar(&days, "days", 90, "How old an image may be in days before removal (defaults to 90)")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "simulate the deletion")
	rootCmd.Flags().BoolVarP(&removeSemver, "no-semver", "n", false, "disables protection of semantic versioned tags")
	rootCmd.Flags().BoolVar(&debug, "debug", false, "enables debug logging")
	rootCmd.Flags().BoolVar(&force, "force", false, "force will remove images even if not images would remain")
	rootCmd.Flags().StringVar(&repo, "repo", "", "specifies to run against a single repository")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.SetEnvPrefix("ECR_CLEANER")
		// Search config in home directory with name ".ecr-cleaner" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".ecr-cleaner")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
