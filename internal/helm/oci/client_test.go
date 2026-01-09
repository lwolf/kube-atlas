package oci

import (
    "testing"
)

func TestClient_Pull_Skipped(t *testing.T) {
    // We skip actual OCI pull in unit tests to avoid network/auth flakes.
    // In a real scenario we might mock the registry or use a local one.
    t.Skip("Skipping OCI pull test (network required)")
}
