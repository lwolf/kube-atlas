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

package render

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/lwolf/kube-atlas/pkg/config"
	"github.com/lwolf/kube-atlas/pkg/fileutil"
	"github.com/lwolf/kube-atlas/pkg/helmexec"
)

// renderCmd represents the render command
var CmdRender = &cobra.Command{
	Use:   "render",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		var state config.ClusterSpec
		err := viper.Unmarshal(&state)
		if err != nil {
			log.Fatal().Err(err).Msg("unable to unmarshal config")
		}
		var releases []config.ReleaseSpec
		if len(args) > 0 {
			for _, r := range args {
				rl := state.ReleaseByName(r)
				if rl != nil {
					releases = append(releases, *rl)
				}
			}
			if len(releases) == 0 {
				log.Fatal().Strs("names", args).Msg("no releases with these names found in the config")
			}
		} else {
			releases = state.Releases
		}

		clusterPath := state.Defaults.GetReleasePath()
		err = os.MkdirAll(clusterPath, 0755)
		if err != nil {
			log.Fatal().Err(err)
		}
		for _, r := range releases {
			helm := helmexec.New(&log.Logger)
			helm.SetHelmBinary("helm")
			destTmp, err := ioutil.TempDir("", "helm-")
			if err != nil {
				log.Error().Err(err)
			}
			defer os.RemoveAll(destTmp)
			args := []string{"--output-dir", destTmp, "--name", r.Name}
			configPath, err := r.GetValuesPath(&state.Defaults)
			if err != nil {
				log.Error().Err(err).Msg("failed to get values directory")
			}
			for _, configFile := range r.Values {
				fullPath := filepath.Join(configPath, configFile)
				log.Info().Str("file", fullPath).Msg("checking config values")
				if fileutil.Exists(fullPath) {
					args = append(args, "--values", fullPath)
				} else {
					log.Error().Str("file", configFile).Msg("values file does not exists, skipping")
				}
			}
			chartPath, err := r.GetChartPath(&state.Defaults)
			if err != nil {
				log.Error().Err(err).Msg("failed to get chart directory")
			}
			if err := helm.TemplateRelease(chartPath, args...); err != nil {
				log.Fatal().Err(err)
			}
			dstPath, err := r.GetReleasePath(&state.Defaults)
			if err != nil {
				log.Error().Err(err).Msg("failed to get destination directory")
			}
			log.Info().Str("path", dstPath).Msg("destination path for the rendered chart")
			err = os.RemoveAll(dstPath)
			if err != nil {
				log.Fatal().Err(err)
			}
			err = os.Rename(filepath.Join(destTmp, r.Name, "templates"), dstPath)
			log.Info().
				Str("path", dstPath).
				Str("source", filepath.Join(destTmp, r.Name, "templates")).
				Msg("moving rendered templates to the destination")
			if err != nil {
				log.Fatal().Err(err)
			}
			manifestsPath, err := r.GetManifestsPath(&state.Defaults)
			if err != nil {
				log.Fatal().Err(err)
			}
			for _, m := range r.Manifests {
				manifest := filepath.Join(manifestsPath, m)
				manifestDestPath := filepath.Join(dstPath, fmt.Sprintf("manifest-%s", m))
				log.Info().Str("source", manifest).Str("dst", manifestDestPath).Msg("going to copy raw manifests")
				err = fileutil.CopyFile(manifest, manifestDestPath)
				if err != nil {
					log.Error().Err(err)
				}
			}
		}
	},
}

func init() {}
