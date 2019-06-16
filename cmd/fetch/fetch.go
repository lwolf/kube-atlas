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

package fetch

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/lwolf/kube-atlas/pkg/helmexec"
	"github.com/lwolf/kube-atlas/pkg/state"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	chartName    string
	chartVersion string
	fetchAll     bool
	devel        bool
)

var fetchUsage = `fetch command fetches helm chart and stores it 
in the pkgName/chart directory. It uses chart and version specified in the
kube-atlas.yaml, but could be overridden using args. 

Examples:

	# fetch charts for all releases present in the config 
	kube-atlas fetch --all

	# fetch chart for the release name prometheus from the config
	kube-atlas fetch prometheus

	# override values from the config (if present) for specific release or
	# create a new directory structure 
	kube-atlas fetch prometheus --chart stable/prometheus --version 8.12.2

`

// upgradeCmd represents the upgrade command
var CmdFetch = &cobra.Command{
	Use:     "fetch",
	Example: "\tkube-atlas fetch prometheus\n\tkube-atlas fetch prometheus --chart=stable/prometheus-operator",
	Short:   "Fetch/upgrade helm chart",
	Long:    fetchUsage,
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var s state.ClusterSpec
		err := viper.Unmarshal(&s)
		if err != nil {
			log.Fatal().Err(err).Msg("unable to unmarshal config")
			return
		}
		var releases []state.ReleaseSpec
		if fetchAll {
			releases = s.Releases
		} else if len(args) > 0 {
			rl := s.ReleaseByName(args[0])
			if rl == nil {
				log.Fatal().Msg("failed to find release by name in the config")
				return
			}
			if rl.Dirty {
				log.Fatal().Msg("release is marked as dirty in the config, remove the flag first")
				return
			}
			log.Debug().Msgf("release information from the config %v", rl)
			if chartName != "" && rl.Chart != "" && rl.Chart != chartName {
				log.Debug().
					Str("current_chart", rl.Chart).
					Str("new_chart", chartName).
					Msg("going to download different chart, please update the config on success")
				rl.Chart = chartName
			}
			if chartVersion != "" && rl.Version != "" && rl.Version != chartVersion {
				log.Debug().
					Str("current_version", rl.Version).
					Str("new_version", chartVersion).
					Msg("going to download different version of chart, please update the config  on success")
				rl.Version = chartVersion
			}
			releases = append(releases, *rl)
		} else {
			log.Fatal().Msg("either --all or release name is required")
		}
		for _, release := range releases {
			chartPath, err := release.GetChartPath(&s.Defaults)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to construct chart directory for package")
			}

			// make sure that directory structure exists
			err = release.InitDirs(&s.Defaults)
			if err != nil {
				log.Error().Err(err).Msg("failed to populate directories for package")
			}
			fetchFlags := []string{}
			if chartVersion != "" {
				fetchFlags = append(fetchFlags, "--version", release.Version)
			}
			if release.Devel {
				fetchFlags = append(fetchFlags, "--devel")
			}
			log.Debug().Msgf("going to fetch chart from %s", release.Chart)
			destTmp, err := ioutil.TempDir("", "helm-")
			if err != nil {
				log.Error().Err(err)
			}
			defer os.RemoveAll(destTmp)

			helm := helmexec.New(&log.Logger)
			fetchFlags = append(fetchFlags, "--untar", "--untardir", destTmp)
			log.Info().Strs("fetchFlags", fetchFlags).Msg("fetch flags:")
			if err := helm.Fetch(release.Chart, fetchFlags...); err != nil {
				log.Fatal().Err(err).Msg("failed to fetch chart")
			}
			files, err := ioutil.ReadDir(destTmp)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to read tmp directory with the chart")
			}
			if len(files) != 1 {
				log.Fatal().Msg("failed to find chart directory after fetching...panic?!")
			}
			chartTmp := filepath.Join(destTmp, files[0].Name())

			log.Info().Str("pkgChart", chartPath).Msgf("removing chart sources")
			if err := os.RemoveAll(chartPath); err != nil {
				log.Error().Err(err)
			}
			log.Debug().Str("src", chartTmp).Str("dest", chartPath).Msg("moving chart to destination")
			if err := os.Rename(chartTmp, chartPath); err != nil {
				log.Error().Err(err)
			}
		}
	},
}

func init() {
	CmdFetch.Flags().StringVar(&chartName, "chart", "", "Name of the helm chart to fetch into package, e.g. stable/prometheus")
	CmdFetch.Flags().StringVar(&chartVersion, "version", "", "Version of the helm chart to fetch into package, e.g. 8.11.4")
	CmdFetch.Flags().BoolVar(&fetchAll, "all", false, "Fetch all releases listed in the config")
	CmdFetch.Flags().BoolVar(&devel, "devel", false, "Fetch development versions of the chart")
}
