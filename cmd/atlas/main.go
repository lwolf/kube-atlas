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

    default:
        printUsage()
        os.Exit(1)
    }
}

func printUsage() {
    fmt.Println("usage: atlas <command> [<args>]")
    fmt.Println("commands: version, repo, release, fetch, render, status, diff, check, drift")
}
