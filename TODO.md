## bugs:
[x] directory for clusterName doesn't get created
[ ] we should fail if we can't create any of the directories
[x] render specific chart does not work
[x] namespace is not set during render
[ ] check every file for existence before copying 
[ ] add option to concatenate all the rendered manifests
[ ] iterate over manifests, if it's directory
    [ ] wildcard copy everything under manifests
[ ] subcharts a not being copied to the release folder
## features
[ ] init kube-atlas.yaml from helmfile
[ ] distinguish local/remote charts, don't try to fetch local
[ ] fetch --all to download all charts


-------
## future
[ ] ability to support multiple cluster/versions/releases
[ ] ability to set release name
[ ] interactive init
[ ] research and add support for json patch/merge
    * https://github.com/pivotal-cf/yaml-patch
    * https://github.com/cppforlife/go-patch
    * https://github.com/evanphx/json-patch
