package state

import (
	"bytes"
	"os"
	"path/filepath"
	"text/template"

	securejoin "github.com/cyphar/filepath-securejoin"
	"github.com/spf13/viper"
)

const (
	RenderModeSingle           = "single"
	RenderModeMulti            = "multi"
	RenderModeCustom           = "custom"
	DefaultChartDir            = "chart"
	DefaultManifestsDir        = "manifests"
	DefaultValuesDir           = "values"
	DefaultPatchesDir          = "patches"
	DefaultReleaseDir          = "releases"
	DefaultKubeVersion         = "1.14.1-0"
	DefaultRenderMode          = RenderModeSingle
	DefaultReleasePathTemplate = "{{.ReleasesPath}}/{{.ClusterName}}/{{.ReleaseNamespace}}/{{.ReleaseName}}"
)

type DefaultConfig struct {
	ClusterName         string `yaml:"clusterName"`
	ChartPath           string `yaml:"chartPath"`
	ManifestsPath       string `yaml:"manifestsPath"`
	ValuesPath          string `yaml:"valuesPath"`
	PatchesPath         string `yaml:"patchesPath"`
	SourcePath          string `yaml:"sourcePath"`
	ReleasePath         string `yaml:"releasePath"`
	KubeVersion         string `yaml:"kubeVersion"`
	RenderMode          string `yaml:"renderMode"`
	ReleasePathTemplate string `yaml:"releasePathTemplate"`
}

func (dc *DefaultConfig) GetReleasePath() string {
	if dc.ReleasePath != "" {
		return dc.ReleasePath
	}
	return DefaultReleaseDir
}
func (dc *DefaultConfig) GetRenderMode() string {
	if dc.RenderMode != "" {
		return dc.RenderMode
	}
	return DefaultRenderMode
}

func (dc *DefaultConfig) GetReleasePathTemplate() string {
	if dc.ReleasePathTemplate != "" {
		return dc.ReleasePathTemplate
	}
	return DefaultReleasePathTemplate
}

func (dc *DefaultConfig) GetChartPath() string {
	if dc.ChartPath != "" {
		return dc.ChartPath
	}
	return DefaultChartDir
}
func (dc *DefaultConfig) GetManifestsPath() string {
	if dc.ManifestsPath != "" {
		return dc.ManifestsPath
	}
	return DefaultManifestsDir
}

func (dc *DefaultConfig) GetValuesPath() string {
	if dc.ValuesPath != "" {
		return dc.ValuesPath
	}
	return DefaultValuesDir
}
func (dc *DefaultConfig) GetPatchesPath() string {
	if dc.PatchesPath != "" {
		return dc.PatchesPath
	}
	return DefaultPatchesDir
}

func (dc *DefaultConfig) GetKubeVersion() string {
	if dc.KubeVersion != "" {
		return dc.KubeVersion
	}
	return DefaultKubeVersion
}

type ClusterSpec struct {
	Defaults DefaultConfig `yaml:"defaults"`

	Repositories []RepositorySpec `yaml:"repositories"`
	Releases     []ReleaseSpec    `yaml:"releases"`
}

func LoadSpec() (*ClusterSpec, error) {
	var state ClusterSpec
	err := viper.Unmarshal(&state)
	if err != nil {
		return nil, err
	}
	err = ValidateConfig()
	if err != nil {
		return nil, err
	}
	return &state, nil
}

func ValidateConfig() error {
	return nil
}

func (cs *ClusterSpec) ReleaseByName(name string) *ReleaseSpec {
	for _, r := range cs.Releases {
		if r.Name == name {
			return &r
		}
	}
	return nil
}

func (cs *ClusterSpec) CreateSourceDirectories() error {
	for _, r := range cs.Releases {
		err := r.InitDirs(&cs.Defaults)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cs *ClusterSpec) CreateReleaseDirectories() error {
	clusterPath := cs.Defaults.GetReleasePath()
	err := os.MkdirAll(clusterPath, 0755)
	if err != nil {
		return err
	}
	for _, r := range cs.Releases {
		dstPath, err := r.GetReleasePath(&cs.Defaults)
		if err != nil {
			return err
		}
		err = os.MkdirAll(dstPath, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

// RepositorySpec defines values for a helm charts repo
type RepositorySpec struct {
	Name     string `yaml:"name"`
	URL      string `yaml:"url"`
	CertFile string `yaml:"certFile"`
	KeyFile  string `yaml:"keyFile"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// ReleaseSpec defines the structure of a release
type ReleaseSpec struct {
	Version     string `yaml:"version"`
	KubeVersion string `yaml:"kubeVersion"`
	// Name is the name of this release
	Name  string `yaml:"name"`
	Chart string `yaml:"chart"`
	// Devel, when set to true, use development versions, too. Equivalent to version '>0.0.0-0'
	Devel       bool     `yaml:"devel"`
	Dirty       bool     `yaml:"dirty"`
	Namespace   string   `yaml:"namespace"`
	ReleasePath string   `yaml:"release_path"`
	RenderMode  string   `yaml:"renderMode"`
	Values      []string `yaml:"values"`
	Manifests   []string `yaml:"manifests"`
}

type releaseTemplateVars struct {
	ReleasesPath     string
	ClusterName      string
	ReleaseNamespace string
	ReleaseName      string
}

func (r *ReleaseSpec) GetReleasePath(d *DefaultConfig) (string, error) {
	// "{{releasePath}}/{{clusterName}}/{{releaseNamespace}}/{{releaseName}}"
	var b bytes.Buffer
	vars := releaseTemplateVars{
		ReleaseName:      r.Name,
		ReleaseNamespace: r.Namespace,
		ReleasesPath:     d.ReleasePath,
		ClusterName:      d.ClusterName,
	}
	tmpl := template.Must(template.New("path").Parse(d.GetReleasePathTemplate()))
	err := tmpl.Execute(&b, vars)
	if err != nil {
		return "", err
	}
	return filepath.Clean(b.String()), nil
}

func (r *ReleaseSpec) GetKubeVersion(d *DefaultConfig) string {
	if r.KubeVersion != "" {
		return r.KubeVersion
	}
	return d.GetKubeVersion()
}

func (r *ReleaseSpec) GetRenderMode(d *DefaultConfig) string {
	if r.RenderMode != "" {
		return r.RenderMode
	}
	return d.GetRenderMode()
}

func (r *ReleaseSpec) GetPkgPath(d *DefaultConfig) (string, error) {
	return securejoin.SecureJoin(d.SourcePath, r.Name)
}

func (r *ReleaseSpec) GetChartPath(d *DefaultConfig) (string, error) {
	pkgPath, err := r.GetPkgPath(d)
	if err != nil {
		return "", nil
	}
	return securejoin.SecureJoin(pkgPath, d.GetChartPath())
}
func (r *ReleaseSpec) GetManifestsPath(d *DefaultConfig) (string, error) {
	pkgPath, err := r.GetPkgPath(d)
	if err != nil {
		return "", nil
	}
	return securejoin.SecureJoin(pkgPath, d.GetManifestsPath())
}
func (r *ReleaseSpec) GetValuesPath(d *DefaultConfig) (string, error) {
	pkgPath, err := r.GetPkgPath(d)
	if err != nil {
		return "", nil
	}
	return securejoin.SecureJoin(pkgPath, d.GetValuesPath())
}
func (r *ReleaseSpec) GetPatchesPath(d *DefaultConfig) (string, error) {
	pkgPath, err := r.GetPkgPath(d)
	if err != nil {
		return "", nil
	}
	return securejoin.SecureJoin(pkgPath, d.GetPatchesPath())
}

func (r *ReleaseSpec) InitDirs(d *DefaultConfig) error {
	pkgPath, err := r.GetPkgPath(d)
	if err != nil {
		return err
	}
	err = os.MkdirAll(pkgPath, 0755)
	if err != nil {
		return err
	}
	for _, sub := range []string{d.GetChartPath(), d.GetManifestsPath(), d.GetPatchesPath(), d.GetValuesPath()} {
		subPath, err := securejoin.SecureJoin(pkgPath, sub)
		if err != nil {
			return err
		}
		err = os.MkdirAll(subPath, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}
