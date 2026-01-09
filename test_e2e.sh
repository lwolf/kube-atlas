#!/bin/bash
set -e

# Setup
TEST_DIR="test_env"
rm -rf $TEST_DIR
mkdir $TEST_DIR
cd $TEST_DIR

ATLAS_BIN="../bin/atlas"

echo "Initializing atlas.yaml..."
cat <<EOF > atlas.yaml
apiVersion: atlas/v1
kind: AtlasConfig
releases: []
policy:
  command: ["grep", "namespace"]
EOF

echo "Adding repo (HTTP bitnami)..."
$ATLAS_BIN repo add --name bitnami --url https://charts.bitnami.com/bitnami

echo "Adding release (HTTP nginx)..."
$ATLAS_BIN release add --id nginx --namespace default --chart bitnami/nginx --version 15.0.0

echo "Adding release (OCI podinfo)..."
$ATLAS_BIN release add --id podinfo --namespace default --chart oci://ghcr.io/stefanprodan/charts/podinfo --version 6.7.0

echo "Fetching charts..."
$ATLAS_BIN fetch

echo "Verifying fetch..."
if [ ! -d "atlas/releases/namespaces/default/nginx/chart" ]; then
    echo "Failed to fetch nginx"
    exit 1
fi
if [ ! -d "atlas/releases/namespaces/default/podinfo/chart" ]; then
    echo "Failed to fetch podinfo"
    exit 1
fi
if [ ! -f "atlas/lock.yaml" ]; then
    echo "Lock file missing"
    exit 1
fi

echo "Rendering..."
$ATLAS_BIN render

echo "Verifying render..."
if [ ! -f "atlas/rendered/namespaces/default/nginx.yaml" ]; then
    echo "Failed to render nginx"
    exit 1
fi
if [ ! -f "atlas/rendered/namespaces/default/podinfo.yaml" ]; then
    echo "Failed to render podinfo"
    exit 1
fi

echo "Checking status..."
$ATLAS_BIN status

echo "Checking policies..."
$ATLAS_BIN check

echo "Checking drift (mock)..."
# Mock kubectl
cat <<EOF > kubectl
#!/bin/sh
if [ "\$1" = "diff" ]; then
    echo "Mock kubectl: diffing..."
    exit 0
fi
echo "Mock kubectl called with \$*"
exit 1
EOF
chmod +x ./kubectl
export PATH="$(pwd):$PATH"
echo "Current PATH: $PATH"

$ATLAS_BIN drift

echo "CLI Workflow Test PASSED"
