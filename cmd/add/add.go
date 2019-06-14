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

package add

import (
	"github.com/lwolf/kube-atlas/pkg/config"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	addUsage = `Add command creates a predefined directory structure
for the new package and optionally fetches helm chart.

To initialize new package run:

	$ kube-atlas add prometheus

or to add multiple at one step
	$ kube-atlas add prometheus grafana
`
)

// addCmd represents the add command
var CmdAdd = &cobra.Command{
	Use:     "add <name>",
	Example: "\tkube-atlas add prometheus",
	Short:   "Creates directory structure for one or more new packages",
	Long:    addUsage,
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var state config.ClusterSpec
		err := viper.Unmarshal(&state)
		if err != nil {
			log.Fatal().Err(err).Msg("unable to unmarshal config")
		}
		for _, pkg := range args {
			r := state.ReleaseByName(pkg)
			if r == nil {
				// package does not exists in the kube-atlas.yaml
				r = &config.ReleaseSpec{Name: pkg}
				log.Info().Str("pkg", pkg).Msg("New package is being added, add record to your kube-atlas.yaml")
			}
			log.Info().Str("pkg", pkg).Msg("Creating/Fixing directory structure for the package")
			err = r.InitDirs(&state.Defaults)
			if err != nil {
				log.Error().Err(err).Msg("failed to create directories")
			}
		}
	},
}

func init() {}
