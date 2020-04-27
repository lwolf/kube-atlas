package exec_kustomize

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog"

	"github.com/lwolf/kube-atlas/pkg/exec"
)

const (
	command = "kustomize"
)

type execer struct {
	binary string
	runner exec.Runner
	logger *zerolog.Logger
	extra  []string
}

// New for running helm commands
func NewExecKustomize(logger *zerolog.Logger) *execer {
	return &execer{
		binary: command,
		logger: logger,
		runner: &exec.ShellRunner{
			Logger: logger,
		},
	}
}

func (e *execer) SetExtraArgs(args ...string) {
	e.extra = args
}

func (e *execer) SetBinary(bin string) {
	e.binary = bin
}

func (e *execer) exec(args []string, env map[string]string) ([]byte, error) {
	cmdargs := args
	if len(e.extra) > 0 {
		cmdargs = append(cmdargs, e.extra...)
	}
	cmd := fmt.Sprintf("exec: %s %s", e.binary, strings.Join(cmdargs, " "))
	e.logger.Debug().Msg(cmd)
	bytes, err := e.runner.Execute(e.binary, cmdargs, env)
	e.logger.Debug().Msgf("%s: %s", cmd, bytes)
	return bytes, err
}

func (e *execer) Build(chart string, flags ...string) error {
	out, err := e.exec(append([]string{"build", chart}, flags...), map[string]string{})
	e.write(out)
	return err
}

func (e *execer) info(out []byte) {
	if len(out) > 0 {
		e.logger.Info().Msgf("%s", out)
	}
}

func (e *execer) write(out []byte) {
	if len(out) > 0 {
		e.logger.Debug().Msgf("%s\n", out)
	}
}
