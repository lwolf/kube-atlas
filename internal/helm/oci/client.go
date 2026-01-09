package oci

import (
    "fmt"
    "os"

    "helm.sh/helm/v3/pkg/action"
    "helm.sh/helm/v3/pkg/cli"
)

// Client handles OCI operations.
type Client struct {
    settings *cli.EnvSettings
}

func NewClient() *Client {
    return &Client{
        settings: cli.New(),
    }
}

// Pull downloads a chart from an OCI registry to the destination directory.
// ref should be full OCI ref (e.g., oci://ghcr.io/foo/bar)
// version is required.
func (c *Client) Pull(ref string, version string, destDir string) error {
    // We use Helm's action.Pull
    // This requires some setup of the action configuration,
    // although for Pull specifically it might be simpler as it's client-side.
    
    // We can use the generic 'Pull' action from Helm SDK
    cfg := new(action.Configuration)
    // We don't strictly need a kubeconfig for pulling, so init with no-op
    if err := cfg.Init(nil, c.settings.Namespace(), os.Getenv("HELM_DRIVER"), func(format string, v ...interface{}) {}); err != nil {
        return fmt.Errorf("failed to init helm config: %w", err)
    }

    pull := action.NewPullWithOpts(action.WithConfig(cfg))
    pull.Settings = c.settings
    pull.Version = version
    pull.DestDir = destDir
    pull.Untar = true
    
    // Registry loggin is handled by Helm if configured in env, 
    // or we can explicitly set it on the action if needed.
    
    // Execute pull
    // Note: Helm's Pull returns the path to the downloaded file
    _, err := pull.Run(ref)
    if err != nil {
        return fmt.Errorf("failed to pull chart %s:%s: %w", ref, version, err)
    }

    return nil
}
