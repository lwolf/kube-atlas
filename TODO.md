## bugs:
* [x] directory for clusterName doesn't get created
* [x] render specific chart does not work
* [x] namespace is not set during render
* [x] should fail if can't create any of the directories
*     [x] validate/create target directory before rendering/processing
*     [x] check every file for existence before copying 
* [x] iterate over manifests, if it's directory
*     [x] wildcard copy everything under manifests
* [x] subcharts are not being copied to the release folder
## features
* [ ] add option to concatenate all the rendered manifests
* [ ] ability to template destination directory for release 
* [ ] add --dry-run mode
* [ ] init kube-atlas.yaml from helmfile
* [ ] distinguish local/remote charts, don't try to fetch local
*     [x] add `dirty` flag as a workaround to block chart overwriting 
* [x] fetch --all to download all charts
* [ ] fetch only if versions are differ or `--force` is set
* [ ] write proper readme
* [ ] consider adding ignore list for chart, e.g. do not copy `tests` to release

-------
## future
* [ ] ability to support multiple cluster/versions/releases
* [ ] ability to set release name
* [x] ability to set desired -kube-version
     [ ] bonus: warn if `.Capabilities.KubeVersion.GitVersion` during templating
* [ ] support rules for extracting some resource types to the predefined locations
     [ ] e.g. store dashboard resources in common place
* [ ] ability to inline values in kube-atlas.yaml without requiring values.yaml file
* [ ] interactive init
* [ ] remove dependency on helm
* [ ] support injectors (linkerd, istio)
* [ ] support kustomize
* [ ] research and add support for json patch/merge
    * https://github.com/pivotal-cf/yaml-patch
    * https://github.com/cppforlife/go-patch
    * https://github.com/evanphx/json-patch
