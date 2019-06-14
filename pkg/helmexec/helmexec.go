package helmexec

// Interface for executing helm commands
type Interface interface {
	SetExtraArgs(args ...string)
	SetHelmBinary(bin string)

	AddRepo(name, repository, certfile, keyfile, username, password string) error
	UpdateRepo() error
	BuildDeps(chart string) error
	UpdateDeps(chart string) error
	TemplateRelease(chart string, flags ...string) error
	Fetch(chart string, flags ...string) error
	Lint(chart string, flags ...string) error
}

type DependencyUpdater interface {
	UpdateDeps(chart string) error
}
