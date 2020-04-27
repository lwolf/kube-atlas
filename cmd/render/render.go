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
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/lwolf/kube-atlas/pkg/fileutil"
	"github.com/lwolf/kube-atlas/pkg/helmexec"
	"github.com/lwolf/kube-atlas/pkg/state"
)

var (
	renderAll bool
	dryRun    bool
)

type releaseType string

const (
	releaseTypeHelm      releaseType = "helm"
	releaseTypeKustomize releaseType = "kustomize"
	releaseTypeRaw       releaseType = "raw"
	releaseTypeNone      releaseType = "none"
)

func releaseContentType(release *state.ReleaseSpec, s *state.ClusterSpec) releaseType {
	rlog := log.With().Str("release", release.Name).Logger()
	chartPath, err := release.GetChartPath(&s.Defaults)
	if err != nil {
		rlog.Error().Err(err).Msg("failed to get chart directory")
		return releaseTypeNone
	}
	var fls []os.FileInfo
	fls, err = ioutil.ReadDir(chartPath)
	if err != nil {
		rlog.Error().Err(err).Msg("failed to get chart directory content")
		return releaseTypeNone
	}
	if len(fls) == 0 {
		return releaseTypeNone
	}
	var yamlsFound bool
	for _, f := range fls {
		if f.Name() == "Chart.yaml" {
			return releaseTypeHelm
		} else if f.Name() == "kustomization.yaml" {
			return releaseTypeKustomize
		} else if filepath.Ext(f.Name()) == ".yaml" {
			yamlsFound = true
		}
	}
	if yamlsFound {
		return releaseTypeRaw
	}
	return releaseTypeNone
}

func renderHelmChart(release *state.ReleaseSpec, s *state.ClusterSpec) error {
	rlog := log.With().Str("release", release.Name).Logger()
	renderTmp, err := ioutil.TempDir("", "helm-release-")
	if err != nil {
		rlog.Fatal().Err(err).Msg("failed to create temp directory")
	}
	defer func() {
		err := os.RemoveAll(renderTmp)
		if err != nil {
			rlog.Error().Err(err).Msg("failed remove temp directory")
		}
	}()

	helm := helmexec.New(&log.Logger)
	args := []string{
		"--output-dir", renderTmp,
		"--name", release.Name,
		"--kube-version", release.GetKubeVersion(&s.Defaults),
	}
	if release.Namespace != "" {
		args = append(args, "--namespace", release.Namespace)
	}
	configPath, err := release.GetValuesPath(&s.Defaults)
	if err != nil {
		rlog.Error().Err(err).Msg("failed to get values directory")
	}
	for _, configFile := range release.Values {
		fullPath := filepath.Join(configPath, configFile)
		isDir, err := fileutil.IsDir(fullPath)
		if err != nil {
			rlog.Error().Err(err).Msg("failed to check path")
			continue
		}
		if isDir {
			rlog.Error().Err(err).Msg("only files are supported at the moment, skipping the directory")
			continue
		}
		if fileutil.Exists(fullPath) {
			args = append(args, "--values", fullPath)
		} else {
			rlog.Error().Str("file", configFile).Msg("values file does not exists, skipping")
		}
	}
	chartPath, err := release.GetChartPath(&s.Defaults)
	if err != nil {
		rlog.Error().Err(err).Msg("failed to get chart directory")
		return err
	}
	if err := helm.TemplateRelease(chartPath, args...); err != nil {
		return err
	}
	dstPath, err := release.GetReleasePath(&s.Defaults)
	if err != nil {
		return err
	}
	err = os.MkdirAll(dstPath, 0755)
	if err != nil {
		return err
	}
	rlog.Debug().Str("path", dstPath).Msg("destination path for the rendered chart")
	err = os.RemoveAll(dstPath)
	if err != nil {
		rlog.Fatal().Err(err)
	}
	err = os.MkdirAll(dstPath, 0755)
	if err != nil {
		rlog.Fatal().Err(err)
	}
	var fds []os.FileInfo
	// there should be only a single directory after helm template in the temp

	// resultDir := filepath.Join(destTmp, r.Name)
	if fds, err = ioutil.ReadDir(renderTmp); err != nil {
		rlog.Fatal().Err(err).Msg("failed to read directory content")
	}
	chartTmpPath := filepath.Join(renderTmp, fds[0].Name())
	if fds, err = ioutil.ReadDir(chartTmpPath); err != nil {
		rlog.Fatal().Err(err).Msg("failed to read directory content")
	}
	renderMode := release.GetRenderMode(&s.Defaults)
	switch renderMode {
	case state.RenderModeSingle:
		var resultYaml bytes.Buffer
		for _, fd := range fds {
			err = concatYamls(filepath.Join(chartTmpPath, fd.Name()), &resultYaml)
		}
		chartFile := filepath.Join(dstPath, fmt.Sprintf("%s.%s", release.Name, "yaml"))
		rlog.Debug().Str("chartFile", chartFile).Msg("chart file result")
		in, err := os.Create(chartFile)
		if err != nil {
			rlog.Fatal().Err(err).Msg("failed to create destination yaml for chart")
		}
		defer in.Close()
		_, err = in.Write(resultYaml.Bytes())
		if err != nil {
			rlog.Fatal().Err(err).Msg("failed to write concatenated yaml of chart")
		}
	case state.RenderModeMulti:
		for _, fd := range fds {
			srcfp := filepath.Join(chartTmpPath, fd.Name())
			dstfp := filepath.Join(dstPath, fd.Name())
			rlog.Debug().Msgf("copy from %s to %s", srcfp, dstfp)
			if fd.IsDir() {
				err = fileutil.CopyDir(srcfp, dstfp, "")
				if err != nil {
					rlog.Error().Err(err).Msg("error copying dir")
				}
			} else {
				err = fileutil.CopyFile(srcfp, dstfp)
				if err != nil {
					rlog.Error().Err(err).Msg("error copying file")
				}
			}
		}
	default:
		rlog.Error().Str("renderMode", renderMode).Msg("Not Implemented Error")
	}
	return nil
}

func concatYamls(src string, buf *bytes.Buffer) error {
	var fi os.FileInfo
	var err error
	if fi, err = os.Stat(src); err != nil {
		return err
	}
	if fi.IsDir() {
		var fds []os.FileInfo
		if fds, err = ioutil.ReadDir(src); err != nil {
			return err
		}
		for _, fd := range fds {
			err = concatYamls(path.Join(src, fd.Name()), buf)
			if err != nil {
				return err
			}
		}
	} else {
		var srcfd *os.File
		if srcfd, err = os.Open(src); err != nil {
			return err
		}
		defer srcfd.Close()
		if _, err = io.Copy(buf, srcfd); err != nil {
			return err
		}
		// XXX: is there a better way to write new line
		if _, err = fmt.Fprintln(buf, ""); err != nil {
			return err
		}

	}
	return err
}

func renderKustomize(release *state.ReleaseSpec, s *state.ClusterSpec) error {
	return fmt.Errorf("not implemented error")
}

func renderRaw(release *state.ReleaseSpec, s *state.ClusterSpec) error {
	return fmt.Errorf("not implemented error")
}

func copyManifests(release *state.ReleaseSpec, s *state.ClusterSpec) error {
	rlog := log.With().Str("release", release.Name).Logger()
	manifestsPath, err := release.GetManifestsPath(&s.Defaults)
	if err != nil {
		return err
	}
	dstPath, err := release.GetReleasePath(&s.Defaults)
	if err != nil {
		return err
	}
	// by default include all the manifests in the folder
	// any value set in manifests key overrides it
	manifests := release.Manifests
	if len(release.Manifests) == 0 {
		log.Debug().
			Str("release", release.Name).
			Msg("no whitelisted manifests found, including all")
		var fds []os.FileInfo
		if fds, err = ioutil.ReadDir(manifestsPath); err != nil {
			return err
		}
		for _, f := range fds {
			log.Debug().Msgf("including `%s`", f.Name())
			manifests = append(manifests, f.Name())
		}
	}
	for _, m := range manifests {
		m = filepath.Clean(m)
		mlog := rlog.With().Str("manifest", m).Logger()
		p := filepath.Join(manifestsPath, m)
		if !fileutil.Exists(p) {
			mlog.Info().Str("path", p).Msg("file does not exist")
			continue
		}
		isDir, err := fileutil.IsDir(p)
		if err != nil {
			continue
		}
		if isDir {
			err = fileutil.CopyDir(p, dstPath, m)
			if err != nil {
				mlog.Error().Err(err).Msg("failed to copy directory")
			}
		} else {
			manifestDestPath := filepath.Join(dstPath, fmt.Sprintf("manifest-%s", m))
			rlog.Debug().Str("source", p).Str("dst", manifestDestPath).Msg("going to copy raw manifests")
			err = fileutil.CopyFile(p, manifestDestPath)
			if err != nil {
				mlog.Error().Err(err).Msg("failed to copy")
			}
		}
	}
	return nil
}

// renderCmd represents the render command
var CmdRender = &cobra.Command{
	Use:   "render",
	Short: "Render or template the release",
	Long: `Render command applies required templating to the 
release and writes the resulting yamls to the corresponding
 directory.

* For helm chart it will do "helm template"
* For kustomize it will do kustomization
* For raw yamls it will just copy it to the destination 

After that it will also copy manifests listed in the spec.`,
	Run: func(cmd *cobra.Command, args []string) {
		s, err := state.LoadSpec()
		if err != nil {
			log.Fatal().Err(err).Msg("unable to unmarshal config")
		}
		err = s.CreateReleaseDirectories()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create destination directories")
		}
		var releases []state.ReleaseSpec
		if renderAll {
			releases = s.Releases
		} else if len(args) > 0 {
			if len(args) > 0 {
				for _, r := range args {
					rl := s.ReleaseByName(r)
					if rl != nil {
						releases = append(releases, *rl)
					}
				}
				if len(releases) == 0 {
					log.Fatal().Strs("names", args).Msg("no releases with these names found in the config")
				}
			}
		} else {
			log.Fatal().Msg("either --all or release name is required")
		}
		err = s.CreateReleaseDirectories()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create target directory structure")
		}
		clusterPath := s.Defaults.GetReleasePath()
		err = os.MkdirAll(clusterPath, 0755)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create release directory")
		}
		for _, r := range releases {
			rlog := log.With().Str("release", r.Name).Logger()
			// validate that chart directory exists and not empty
			_, err := r.GetChartPath(&s.Defaults)
			if err != nil {
				rlog.Error().Err(err).Msg("failed to get chart directory")
				continue
			}
			// process chart directory
			switch releaseContentType(&r, s) {
			case releaseTypeHelm:
				err = renderHelmChart(&r, s)
				if err != nil {
					log.Error().Err(err).Msg("failed to render helm chart")
					break
				}
				rlog.Info().Msg("completed rendering helm chart")
			case releaseTypeKustomize:
				err = renderKustomize(&r, s)
				if err != nil {
					log.Error().Err(err).Msg("failed to apply kustomization")
					break
				}
				rlog.Info().Msg("completed kustomization")
			case releaseTypeRaw:
				err = renderRaw(&r, s)
				if err != nil {
					log.Error().Err(err).Msg("failed to copy raw manifests")
					break
				}
			case releaseTypeNone:
			default:
				rlog.Warn().Msg("unknown release chart folder content, skipping")
			}
			// process manifests directory
			err = copyManifests(&r, s)
			if err != nil {
				log.Error().Err(err).Msg("failed to copy manifests")
				continue
			}
			rlog.Info().Msg("manifests were copied")
		}
	},
}

func init() {
	CmdRender.Flags().BoolVar(&renderAll, "all", false, "Render all the releases listed in the config")
	CmdRender.Flags().BoolVar(&dryRun, "dry-run", false, "Render to stdout")
}
