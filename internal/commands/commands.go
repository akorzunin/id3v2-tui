package commands

import (
	"fmt"
	"os/exec"
)

type Executor interface {
	Run(name string, args ...string) (string, error)
}

type CommandExecutor struct{}

func (e CommandExecutor) Run(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: %w", string(output), err)
	}
	return string(output), nil
}

func NewExecutor() Executor {
	return CommandExecutor{}
}
