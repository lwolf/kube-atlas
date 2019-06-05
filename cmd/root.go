// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"github.com/rs/zerolog"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var globalUsage = `kube-atlas is an opinionated way to manage Kubernetes manifests
in a GitOps way. 

To begin working with kube-atlas, run the 'kube-atlas init' command:

	$ kube-atlas init

This will create a kube-atlas.yaml file in your current directory..

Common actions from this point include:

- kube-atals add:        add entry to your cluster state, will create required directories and entry to kube-atlas.yaml
- kube-atals upgrade:    download new version of chart to your local directory 
- kube-atals render:     render entire cluster state to the release directory `

var (
	cfgFile     string
	sourcePath  string
	releasePath string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kube-atlas",
	Short: "The Kubernetes cluster state manager",
	Long:  globalUsage,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// 	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Error().Err(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	// TODO: setup logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is kube-atlas.yaml)")
	rootCmd.PersistentFlags().StringVar(&sourcePath, "source-path", "", "source directory with charts and manifests")
	rootCmd.PersistentFlags().StringVar(&releasePath, "release-path", "", "release directory for rendered output")
	_ = viper.BindPFlag("source_path", rootCmd.PersistentFlags().Lookup("source-path"))
	_ = viper.BindPFlag("release_path", rootCmd.PersistentFlags().Lookup("release-path"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	}
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Info().Str("config", viper.ConfigFileUsed()).Msg("Config file loaded")
		viper.Debug()
	}
}
