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
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	chartName string
)

var (
	addUsage = `Add command creates a predefined directory structure
for the new package and optionally fetches helm chart.

To initialize new package run:

	$ kube-atlas add prometheus

or to also fetch the chart
	$ kube-atlas add prometheus --chart=stable/prometheus-operator
`
)

func initializePkgDirs(pkgName string) error {
	// TODO: finetune permissions on new directories
	sourcePath := viper.GetString("source_path")
	path := filepath.Join(sourcePath, pkgName)
	log.Info().Str("path", path).Msg("creating pkg directory")
	err := os.MkdirAll(path, 0774)
	if err != nil {
		return err
	}
	for _, sub := range []string{"manifests", "patches", "values", "chart"} {
		subPath := filepath.Join(path, sub)
		log.Info().Str("path", subPath).Msg("creating pkg directory")
		err := os.MkdirAll(subPath, 0774)
		if err != nil {
			return err
		}
	}
	return nil
}

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:     "add <name>",
	Example: "\tkube-atlas add prometheus\n\tkube-atlas add prometheus --chart=stable/prometheus-operator",
	Short:   "Creates directory structure for the new package",
	Long:    addUsage,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := initializePkgDirs(args[0])
		if err != nil {
			log.Fatal().Err(err)
		}
		if chartName != "" {
			log.Info().Msgf("going to fetch chart from %s", chartName)
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.Flags().StringVar(&chartName, "chart", "", "Name of the helm chart to fetch into package, e.g. stable/prometheus")
}
