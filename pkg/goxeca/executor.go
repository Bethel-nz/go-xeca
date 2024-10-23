package goxeca

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type Executor struct{}

func NewExecutor() *Executor {
	return &Executor{}
}

func (e *Executor) Execute(cmd *exec.Cmd) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	output := strings.TrimSpace(stdout.String())

	if err != nil {
		output += strings.TrimSpace(stderr.String())

		if stderr.Len() > 0 {

			return output, fmt.Errorf("command execution failed: %v, Stderr: %s", err, stderr.String())
		}
		return output, fmt.Errorf("command execution failed: %v", err)
	}

	return output, nil
}
