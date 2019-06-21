package helmexec

import (
	"fmt"
	"github.com/rs/zerolog"
	"strings"
	"sync"
)

const (
	command = "helm"
)

type execer struct {
	helmBinary      string
	runner          Runner
	logger          *zerolog.Logger
	kubeContext     string
	extra           []string
	decryptionMutex sync.Mutex
}

// New for running helm commands
func New(logger *zerolog.Logger) *execer {
	return &execer{
		helmBinary: command,
		logger:     logger,
		runner: &ShellRunner{
			logger: logger,
		},
	}
}

func (helm *execer) SetExtraArgs(args ...string) {
	helm.extra = args
}

func (helm *execer) SetHelmBinary(bin string) {
	helm.helmBinary = bin
}

func (helm *execer) AddRepo(name, repository, certfile, keyfile, username, password string) error {
	var args []string
	args = append(args, "repo", "add", name, repository)
	if certfile != "" && keyfile != "" {
		args = append(args, "--cert-file", certfile, "--key-file", keyfile)
	}
	if username != "" && password != "" {
		args = append(args, "--username", username, "--password", password)
	}
	helm.logger.Info().Msgf("Adding repo %v %v", name, repository)
	out, err := helm.exec(args, map[string]string{})
	helm.info(out)
	return err
}

func (helm *execer) UpdateRepo() error {
	helm.logger.Info().Msg("Updating repo")
	out, err := helm.exec([]string{"repo", "update"}, map[string]string{})
	helm.info(out)
	return err
}

func (helm *execer) UpdateDeps(chart string) error {
	helm.logger.Info().Msgf("Updating dependency %v", chart)
	out, err := helm.exec([]string{"dependency", "update", chart}, map[string]string{})
	helm.info(out)
	return err
}

func (helm *execer) BuildDeps(chart string) error {
	helm.logger.Info().Msgf("Building dependency %v", chart)
	out, err := helm.exec([]string{"dependency", "build", chart}, map[string]string{})
	helm.write(out)
	return err
}

func (helm *execer) TemplateRelease(chart string, flags ...string) error {
	out, err := helm.exec(append([]string{"template", chart}, flags...), map[string]string{})
	helm.write(out)
	return err
}

func (helm *execer) Lint(chart string, flags ...string) error {
	helm.logger.Info().Msgf("Linting %v", chart)
	out, err := helm.exec(append([]string{"lint", chart}, flags...), map[string]string{})
	helm.write(out)
	return err
}

func (helm *execer) Fetch(chart string, flags ...string) error {
	helm.logger.Info().Msgf("Fetching %v", chart)
	out, err := helm.exec(append([]string{"fetch", chart}, flags...), map[string]string{})
	helm.info(out)
	return err
}

func (helm *execer) exec(args []string, env map[string]string) ([]byte, error) {
	cmdargs := args
	if len(helm.extra) > 0 {
		cmdargs = append(cmdargs, helm.extra...)
	}
	if helm.kubeContext != "" {
		cmdargs = append(cmdargs, "--kube-context", helm.kubeContext)
	}
	cmd := fmt.Sprintf("exec: %s %s", helm.helmBinary, strings.Join(cmdargs, " "))
	helm.logger.Debug().Msg(cmd)
	bytes, err := helm.runner.Execute(helm.helmBinary, cmdargs, env)
	helm.logger.Debug().Msgf("%s: %s", cmd, bytes)
	return bytes, err
}

func (helm *execer) info(out []byte) {
	if len(out) > 0 {
		helm.logger.Info().Msgf("%s", out)
	}
}

func (helm *execer) write(out []byte) {
	if len(out) > 0 {
		helm.logger.Debug().Msgf("%s\n", out)
	}
}
