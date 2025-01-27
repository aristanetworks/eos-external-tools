package executor

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// An executor that dispatches to os.Exec
type OsExecutor struct {
	Suppress bool // whether to suppress the subcmd output
}

func (ex *OsExecutor) ExecInDir(dir string, name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	if ex.Suppress {
		cmd.Stdout = io.Discard
	} else {
		cmd.Stdout = os.Stdout
	}
	return cmd.Run()

}
func (ex *OsExecutor) Exec(name string, arg ...string) error {
	return ex.ExecInDir("", name, arg...)
}

func (ex *OsExecutor) Output(name string, arg ...string) (string, error) {
	output, err := exec.Command(name, arg...).Output()
	if err != nil {
		escaped_args := shellEscape(append([]string{name}, arg...))
		if exitErr, ok := err.(*exec.ExitError); ok {
			return string(output),
				fmt.Errorf("running `%s` exited with exit-code %d\nstderr:\n%s",
					escaped_args, exitErr.ExitCode(), exitErr.Stderr)
		}
		return string(output),
			fmt.Errorf("running `%s` failed with '%w'", escaped_args, err)
	}
	return string(output), nil
}

// Join strings in a way that preserves the shell token boundries. For instance
// the args ["cat", "a file"] when simply joined with strings.Join would result
// in a string "cat a file" which has a different meaning to the original. This
// function is a simple shell escaping join.
func shellEscape(args []string) string {
	var processedArgs []string
	for _, arg := range args {
		escaped := strings.ReplaceAll(arg, "'", "\\'")
		if strings.Contains(arg, " ") {
			processedArgs = append(processedArgs, "'"+escaped+"'")
		} else {
			processedArgs = append(processedArgs, escaped)
		}
	}
	return strings.Join(processedArgs, " ")
}
