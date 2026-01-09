package main

import (
    "context"
    "flag"
    "fmt"
    "os"
    
    "github.com/charmbracelet/bubbletea"
    "github.com/lwolf/kube-atlas/internal/app"
    "github.com/lwolf/kube-atlas/internal/ui"
)

var version = "0.1.0"

func main() {
    // Define subcommands
    versionCmd := flag.NewFlagSet("version", flag.ExitOnError)
    
    repoAddCmd := flag.NewFlagSet("add", flag.ExitOnError)

    releaseAddCmd := flag.NewFlagSet("add", flag.ExitOnError)
    fetchCmd := flag.NewFlagSet("fetch", flag.ExitOnError)
    renderCmd := flag.NewFlagSet("render", flag.ExitOnError)
    statusCmd := flag.NewFlagSet("status", flag.ExitOnError)

    diffCmd := flag.NewFlagSet("diff", flag.ExitOnError)
    checkCmd := flag.NewFlagSet("check", flag.ExitOnError)
    driftCmd := flag.NewFlagSet("drift", flag.ExitOnError)

    // Setup Usage handlers
    setupUsage(versionCmd, "version", "Print the current version", "atlas version")
    setupUsage(repoAddCmd, "repo add", "Add a new Helm repository", "atlas repo add --name bitnami --url https://charts.bitnami.com/bitnami")
    setupUsage(releaseAddCmd, "release add", "Add a new release to the configuration", "atlas release add --id nginx --namespace default --chart bitnami/nginx --version 15.0.0")
    setupUsage(fetchCmd, "fetch", "Fetch charts for configured releases and update lock file", "atlas fetch\natlas fetch --release nginx")
    setupUsage(renderCmd, "render", "Render Helm charts and Kustomize overlays into final manifests", "atlas render\natlas render --release nginx")
    setupUsage(statusCmd, "status", "Show the status of configured releases and their local charts/renders", "atlas status")
    setupUsage(diffCmd, "diff", "Show the diff between the current configuration and the last rendered manifest", "atlas diff --release nginx")
    setupUsage(checkCmd, "check", "Run policy checks against rendered manifests", "atlas check\natlas check --release nginx")
    setupUsage(driftCmd, "drift", "Detect drift between rendered manifests and the live cluster state", "atlas drift\natlas drift --release nginx")

    // repo add flags
    repoName := repoAddCmd.String("name", "", "Repository Name")
    repoURL := repoAddCmd.String("url", "", "Repository URL")

    // release add flags
    id := releaseAddCmd.String("id", "", "Release ID")
    namespace := releaseAddCmd.String("namespace", "", "Target namespace")
    chart := releaseAddCmd.String("chart", "", "Chart reference (repo/chart)")
    chartVersion := releaseAddCmd.String("version", "", "Chart version")

    // Common flags for operations
    fetchRelease := fetchCmd.String("release", "", "Specific release ID")
    renderRelease := renderCmd.String("release", "", "Specific release ID")

    diffRelease := diffCmd.String("release", "", "Specific release ID")
    checkRelease := checkCmd.String("release", "", "Specific release ID")
    driftCmdRelease := driftCmd.String("release", "", "Specific release ID")

    wd, err := os.Getwd()
    if err != nil {
        fmt.Fprintf(os.Stderr, "failed to get working directory: %v\n", err)
        os.Exit(1)
    }

    application := app.New()

    if len(os.Args) < 2 {
        // TUI Mode
        m, err := ui.NewModel(context.Background(), wd, application)
        if err != nil {
            fmt.Fprintf(os.Stderr, "failed to initialize TUI: %v\n", err)
            os.Exit(1)
        }
        
        p := tea.NewProgram(m)
        if _, err := p.Run(); err != nil {
            fmt.Fprintf(os.Stderr, "error running TUI: %v\n", err)
            os.Exit(1)
        }
        os.Exit(0)
    }

    // CLI Mode
    // ... existing ...

    switch os.Args[1] {
    case "version":
        versionCmd.Parse(os.Args[2:])
        fmt.Println(version)

    case "repo":
        if len(os.Args) < 3 {
             fmt.Println("expected 'add' subcommand")
             os.Exit(1)
        }
        switch os.Args[2] {
        case "add":
            repoAddCmd.Parse(os.Args[3:])
            if *repoName == "" || *repoURL == "" {
                fmt.Println("missing required flags: --name, --url")
                repoAddCmd.PrintDefaults()
                os.Exit(1)
            }
            err = application.AddRepository(app.AddRepositoryOptions{
                RepoRoot: wd,
                Name:     *repoName,
                URL:      *repoURL,
            })
            if err != nil {
                fmt.Fprintf(os.Stderr, "error adding repository: %v\n", err)
                os.Exit(1)
            }
        default:
             fmt.Printf("unknown repo subcommand: %s\n", os.Args[2])
             os.Exit(1)
        }
    
    case "release":
        if len(os.Args) < 3 {
            fmt.Println("expected 'add' subcommand")
            os.Exit(1)
        }
        switch os.Args[2] {
        case "add":
            releaseAddCmd.Parse(os.Args[3:])
            if *id == "" || *namespace == "" || *chart == "" || *chartVersion == "" {
                fmt.Println("missing required flags: --id, --namespace, --chart, --version")
                releaseAddCmd.PrintDefaults()
                os.Exit(1)
            }
            
            err = application.AddRelease(app.AddReleaseOptions{
                ID:        *id,
                Namespace: *namespace,
                Chart:     *chart,
                Version:   *chartVersion,
                RepoRoot:  wd,
            })
            if err != nil {
                fmt.Fprintf(os.Stderr, "error adding release: %v\n", err)
                os.Exit(1)
            }
        default:
            fmt.Printf("unknown release subcommand: %s\n", os.Args[2])
            os.Exit(1)
        }

    case "fetch":
        fetchCmd.Parse(os.Args[2:])
        err = application.Fetch(app.FetchOptions{
            RepoRoot:  wd,
            ReleaseID: *fetchRelease,
        })
        if err != nil {
            fmt.Fprintf(os.Stderr, "fetch failed: %v\n", err)
            os.Exit(1)
        }

    case "render":
        renderCmd.Parse(os.Args[2:])
        err = application.Render(app.RenderOptions{
            RepoRoot:  wd,
            ReleaseID: *renderRelease,
        })
        if err != nil {
            fmt.Fprintf(os.Stderr, "render failed: %v\n", err)
            os.Exit(1)
        }

    case "status":
        statusCmd.Parse(os.Args[2:])
        err = application.Status(app.StatusOptions{
            RepoRoot: wd,
        })
        if err != nil {
            fmt.Fprintf(os.Stderr, "status failed: %v\n", err)
            os.Exit(1)
        }
        
    case "diff":
        diffCmd.Parse(os.Args[2:])
        err = application.Diff(app.DiffOptions{
            RepoRoot:  wd,
            ReleaseID: *diffRelease,
        })
        if err != nil {
            fmt.Fprintf(os.Stderr, "diff failed: %v\n", err)
            os.Exit(1)
        }

    case "check":
        checkCmd.Parse(os.Args[2:])
        err = application.Check(context.Background(), app.CheckOptions{
            RepoRoot:  wd,
            ReleaseID: *checkRelease,
        })
        if err != nil {
            fmt.Fprintf(os.Stderr, "check failed: %v\n", err)
            os.Exit(1)
        }
        
    case "drift":
        driftCmd.Parse(os.Args[2:])
        err = application.Drift(context.Background(), app.DriftOptions{
            RepoRoot:  wd,
            ReleaseID: *driftCmdRelease,
        })
        if err != nil {
            fmt.Fprintf(os.Stderr, "drift logic failed: %v\n", err)
            os.Exit(1)
        }

    case "help":
        if len(os.Args) > 2 {
            switch os.Args[2] {
            case "repo":
                if len(os.Args) > 3 && os.Args[3] == "add" {
                    repoAddCmd.Usage()
                } else {
                    fmt.Println("usage: atlas repo add --help")
                }
            case "release":
                if len(os.Args) > 3 && os.Args[3] == "add" {
                    releaseAddCmd.Usage()
                } else {
                    fmt.Println("usage: atlas release add --help")
                }
            case "fetch":
                fetchCmd.Usage()
            case "render":
                renderCmd.Usage()
            case "status":
                statusCmd.Usage()
            case "diff":
                diffCmd.Usage()
            case "check":
                checkCmd.Usage()
            case "drift":
                driftCmd.Usage()
            case "version":
                versionCmd.Usage()
            default:
                printUsage()
            }
        } else {
            printUsage()
        }

    default:
        printUsage()
        os.Exit(1)
    }
}

func printUsage() {
    fmt.Println("Atlas - GitOps-native Helm chart management")
    fmt.Println()
    fmt.Println("USAGE:")
    fmt.Println("  atlas <command> [<args>]")
    fmt.Println()
    fmt.Println("COMMANDS:")
    fmt.Println("  repo add      Manage Helm repositories")
    fmt.Println("  release add   Manage application releases")
    fmt.Println("  fetch         Download charts and update lock file")
    fmt.Println("  render        Render final Kubernetes manifests")
    fmt.Println("  status        Check local state of releases")
    fmt.Println("  diff          Compare config vs local render")
    fmt.Println("  check         Run policy checks")
    fmt.Println("  drift         Detect drift vs live cluster")
    fmt.Println("  version       Print version information")
    fmt.Println()
    fmt.Println("Use 'atlas <command> --help' for more information on a specific command.")
}

func setupUsage(fs *flag.FlagSet, name, desc, examples string) {
    fs.Usage = func() {
        fmt.Printf("Command: atlas %s\n", name)
        fmt.Printf("Description: %s\n", desc)
        fmt.Println("\nFLAGS:")
        fs.PrintDefaults()
        if examples != "" {
            fmt.Println("\nEXAMPLES:")
            fmt.Println(examples)
        }
        fmt.Println()
    }
}
