#!/bin/bash
set -e

# Build
go build -o atlas-bin ./cmd/atlas

# Cleanup
rm -rf demo-repo
mkdir demo-repo
cd demo-repo

# Init dummy atlas.yaml
touch atlas.yaml

# Add Release
# using a public chart, e.g. bitnami/nginx or similar if network allows.
# Or better, use a simple repo that definitely exists.
# User environment has internet? Usually yes.
# Let's try stable/redis? Or bitnami/nginx? 
# "oci://ghcr.io/stefanprodan/charts/podinfo" is a good test for OCI.
# HTTP: "https://charts.bitnami.com/bitnami" "nginx"

# 1. Add (HTTP)
# atlas release add --id nginx --namespace default --chart bitnami/nginx --version 15.0.0
# Need to setup repository in config first?
# atlas.yaml currently manually edited for repos? 
# My CLI 'release add' doesn't add repos.
# Manually add repo to atlas.yaml.

cat <<EOF > atlas.yaml
apiVersion: atlas/v1
kind: AtlasConfig
repositories: []
policy:
  command: ["grep", "namespace"]

releases: []
EOF

../atlas-bin release add --id my-app --namespace default --chart oci://ghcr.io/stefanprodan/charts/podinfo --version 6.7.0

# 2. Fetch
../atlas-bin fetch --release my-app

# 3. Check Lock
cat atlas/lock.yaml

# 4. Render
../atlas-bin render --release my-app

# 5. Check Output
ls -l atlas/rendered/namespaces/default/my-app.yaml
grep "kind: Deployment" atlas/rendered/namespaces/default/my-app.yaml

# 6. Status
../atlas-bin status

# 7. Modify Config (change version)
# This requires manual edit or re-add?
# We don't have update command yet.
# Let's manually dirty the chart dir to test status.
touch atlas/releases/namespaces/default/my-app/chart/dirty

../atlas-bin status | grep DIRTY

# 8. Diff
# Re-render (in memory diff vs on disk)
# If we modify values, diff should show.
echo "replicaCount: 2" > atlas/releases/namespaces/default/my-app/values/replicas.yaml
../atlas-bin diff --release my-app

# 9. Policy Check
echo "Running policy check..."
# Our policy is "grep -L namespace", which prints filename if "namespace" is NOT found.
# Exit code mapping for grep: 
# 0 (lines found) -> we want inverse? 
# Wait, "grep -L namespace file":
# If "namespace" found inside: NO output, exit 1
# If "namespace" NOT found inside: prints file, exit 0
# The policy checker returns error if command exit code != 0.
# So if "namespace" IS found -> grep exits 1 -> Check fails?
# We want to fail if namespace missing.
# If namespace MISSING -> grep -L prints file -> exit 0 -> Check PASS?
# That config logic is tricky with raw commands.
# Let's use simpler: "grep namespace"
# If found -> exit 0 -> Check PASS.
# If missing -> exit 1 -> Check FAIL.
../atlas-bin check --release my-app

# 10. Drift Check
echo "Running drift check..."
# Mock kubectl
cat <<EOF > kubectl
#!/bin/sh
if [ "\$1" = "diff" ]; then
    echo "Diffing..."
    # Simulate valid diff (exit 1) on first run, success on second?
    # Or just always success for this basic test.
    # Let's clean exit 0 (no drift).
    exit 0
fi
echo "Mock kubectl called with \$*"
exit 1
EOF
chmod +x kubectl
export PATH=\$PWD:\$PATH

../atlas-bin drift --release my-app
