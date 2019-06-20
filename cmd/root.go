// Copyright © 2019 Sergey Nuzhdin ipaq.lw@gmail.com
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
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/lwolf/kube-atlas/cmd/add"
	"github.com/lwolf/kube-atlas/cmd/bootstrap"
	"github.com/lwolf/kube-atlas/cmd/fetch"
	"github.com/lwolf/kube-atlas/cmd/render"
)

var globalUsage = `kube-atlas is an opinionated way to manage Kubernetes manifests
in a GitOps way. 

To begin working with kube-atlas, run the 'kube-atlas init' command:

	$ kube-atlas init

This will create a kube-atlas.yaml file in your current directory..

Common actions from this point include:

- kube-atlas add:        add entry to your cluster state, will create required directories
- kube-atlas fetch:      download new version of chart to your local directory 
- kube-atlas render:     render entire cluster state to the release directory `

var (
	cfgFile     string
	sourcePath  string
	releasePath string
	Version     string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "kube-atlas",
	Short: "The Kubernetes cluster state manager",
	Long:  globalUsage,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// 	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.Error().Err(err).Msg("failed to run")
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
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "file", "f", "kube-atlas.yaml", "path to the config file")
	RootCmd.PersistentFlags().StringVar(&sourcePath, "source-path", "apps", "source directory with charts and manifests")
	RootCmd.PersistentFlags().StringVar(&releasePath, "release-path", "releases", "release directory for rendered output")
	_ = viper.BindPFlag("defaults.sourcePath", RootCmd.PersistentFlags().Lookup("source-path"))
	_ = viper.BindPFlag("defaults.releasePath", RootCmd.PersistentFlags().Lookup("release-path"))

	RootCmd.Version = Version
	RootCmd.AddCommand(fetch.CmdFetch)
	RootCmd.AddCommand(add.CmdAdd)
	RootCmd.AddCommand(render.CmdRender)
	RootCmd.AddCommand(bootstrap.CmdInit)
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
		log.Debug().Str("config", viper.ConfigFileUsed()).Msg("Config file loaded")
		// viper.Debug()
	}
}
