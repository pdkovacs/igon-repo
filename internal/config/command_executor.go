package config

import (
	"fmt"
	"os/exec"

	"github.com/rs/zerolog"
)

type CmdOpts struct {
	Cwd string
}

func (o CmdOpts) String() string {
	return fmt.Sprintf("{Cwd: %v}", o.Cwd)
}

type ExecCmdParams struct {
	Name string
	Args []string
	Opts *CmdOpts
}

func (e ExecCmdParams) String() string {
	option_string := "No options given"
	if e.Opts != nil {
		option_string = fmt.Sprintf("%v", *e.Opts)
	}
	return fmt.Sprintf("%v, %v, %v", e.Name, e.Args, option_string)
}

func ExecuteCommand(params ExecCmdParams, logger zerolog.Logger) (string, error) {
	execCmdLogger := logger.With().Str("prefix", "config.ExecuteCommand").Logger()
	execCmdLogger.Info().Msgf("Starting: %v...", params)

	cmd := exec.Command(params.Name, params.Args...)
	if params.Opts != nil {
		cmd.Dir = params.Opts.Cwd
	}
	out, err := cmd.Output()
	if err != nil {
		exitError, ok := err.(*exec.ExitError)
		if ok {
			return string(exitError.Stderr), exitError
		}
		return "", err
	}
	return string(out), nil
}
