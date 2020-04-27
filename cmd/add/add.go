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
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/lwolf/kube-atlas/pkg/state"
)

var (
	chartName    string
	chartVersion string
	namespace    string
)

var (
	addUsage = `Add command creates a predefined directory structure
for the new package and optionally fetches helm chart.

To initialize new package run:

	$ kube-atlas add prometheus
	$ kube-atlas add prometheus --chart stable/prometheus
	$ kube-atlas add prometheus --chart stable/prometheus --version 8.8.8

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
		var s state.ClusterSpec
		err := viper.Unmarshal(&s)
		if err != nil {
			log.Fatal().Err(err).Msg("unable to unmarshal config")
		}
		if len(args) > 1 && (chartName != "" || chartVersion != "") {
			log.Fatal().Msg("Unable to use `chart` and `version' keys with multiple arguments")
		}
		for _, pkg := range args {
			r := s.ReleaseByName(pkg)
			if r == nil {
				// package does not exists in the kube-atlas.yaml
				r = &state.ReleaseSpec{Name: pkg}
				log.Info().Str("pkg", pkg).Msg("New package is being added, add record to your kube-atlas.yaml")
				if namespace == "" {
					namespace = "%namespace%"
				}
				if chartName == "" {
					chartName = "%repo/chartName%"
				}
				if chartVersion == "" {
					chartVersion = "%chartVersion%"
				}
				msg := fmt.Sprintf(`
  - name: %s
    namespace: %s
    chart: %s
    version: %s
    manifests: []
    values: []
`, pkg, namespace, chartName, chartVersion)
				log.Info().Msg(msg)
			}
			log.Info().Str("pkg", pkg).Msg("Creating/Fixing directory structure for the package")
			err = r.InitDirs(&s.Defaults)
			if err != nil {
				log.Error().Err(err).Msg("failed to create directories")
			}
		}
	},
}

func init() {
	CmdAdd.Flags().StringVar(&chartName, "chart", "", "Name of the helm chart to fetch into package, e.g. stable/prometheus")
	CmdAdd.Flags().StringVar(&chartVersion, "version", "", "Version of the helm chart to fetch into package, e.g. 8.11.4")
	CmdAdd.Flags().StringVar(&namespace, "namespace", "", "Namespace, to add to the kube-atlas file")
}
