package exec_helm

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog"

	"github.com/lwolf/kube-atlas/pkg/exec"
)

const (
	command = "helm"
)

type helmExecer struct {
	binary string
	runner exec.Runner
	logger *zerolog.Logger
	extra  []string
}

// New for running helm commands
func NewExecHelm(logger *zerolog.Logger) *helmExecer {
	return &helmExecer{
		binary: command,
		logger: logger,
		runner: &exec.ShellRunner{
			Logger: logger,
		},
	}
}

func (helm *helmExecer) SetExtraArgs(args ...string) {
	helm.extra = args
}

func (helm *helmExecer) SetBinary(bin string) {
	helm.binary = bin
}
func (helm *helmExecer) AddRepo(name, repository, certfile, keyfile, username, password string) error {
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

func (helm *helmExecer) UpdateRepo() error {
	helm.logger.Info().Msg("Updating repo")
	out, err := helm.exec([]string{"repo", "update"}, map[string]string{})
	helm.info(out)
	return err
}

func (helm *helmExecer) UpdateDeps(chart string) error {
	helm.logger.Info().Msgf("Updating dependency %v", chart)
	out, err := helm.exec([]string{"dependency", "update", chart}, map[string]string{})
	helm.info(out)
	return err
}

func (helm *helmExecer) BuildDeps(chart string) error {
	helm.logger.Info().Msgf("Building dependency %v", chart)
	out, err := helm.exec([]string{"dependency", "build", chart}, map[string]string{})
	helm.write(out)
	return err
}

func (helm *helmExecer) Lint(chart string, flags ...string) error {
	helm.logger.Info().Msgf("Linting %v", chart)
	out, err := helm.exec(append([]string{"lint", chart}, flags...), map[string]string{})
	helm.write(out)
	return err
}

func (helm *helmExecer) Fetch(chart string, flags ...string) error {
	helm.logger.Info().Msgf("Fetching %v", chart)
	out, err := helm.exec(append([]string{"fetch", chart}, flags...), map[string]string{})
	helm.info(out)
	return err
}

func (helm *helmExecer) TemplateRelease(chart string, flags ...string) error {
	out, err := helm.exec(append([]string{"template", chart}, flags...), map[string]string{})
	helm.write(out)
	return err
}

func (helm *helmExecer) exec(args []string, env map[string]string) ([]byte, error) {
	cmdargs := args
	if len(helm.extra) > 0 {
		cmdargs = append(cmdargs, helm.extra...)
	}
	cmd := fmt.Sprintf("exec: %s %s", helm.binary, strings.Join(cmdargs, " "))
	helm.logger.Debug().Msg(cmd)
	bytes, err := helm.runner.Execute(helm.binary, cmdargs, env)
	helm.logger.Debug().Msgf("%s: %s", cmd, bytes)
	return bytes, err
}

func (helm *helmExecer) info(out []byte) {
	if len(out) > 0 {
		helm.logger.Info().Msgf("%s", out)
	}
}

func (helm *helmExecer) write(out []byte) {
	if len(out) > 0 {
		helm.logger.Debug().Msgf("%s\n", out)
	}
}
