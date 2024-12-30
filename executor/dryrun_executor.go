package executor

import (
	"fmt"
	"strings"
)

// An executor that only notes all the commands that would normally be executed.
type DryRunExecutor struct {

	// Note that instead of having multiple ledgers updated with each subsequent
	// invocation, we could store an abstract invocation object, and then have a
	// set of "export" functions, like "ExportShell" that would generate
	// appropriate shell script, but with limited number of exporters those
	// extra abstractions feel unnecessary.

	// human friendly description of what an invocation would do
	invocations []string

	// a script that would be equivalent to the invocations requested
	shellScript []string
}

func (ex *DryRunExecutor) Exec(name string, arg ...string) error {
	escapedInvocation := shellEscape(append([]string{name}, arg...))
	message := fmt.Sprintf("Would execute: %s", escapedInvocation)
	fmt.Println(message)
	ex.invocations = append(ex.invocations, message)
	ex.shellScript = append(ex.shellScript, escapedInvocation)
	return nil
}

func (ex *DryRunExecutor) ExecInDir(dir string, name string, arg ...string) error {
	escapedInvocation := shellEscape(append([]string{name}, arg...))
	message := fmt.Sprintf(
		"In the directory '%s', would execute: %s", dir, escapedInvocation)

	ex.invocations = append(ex.invocations, message)

	// empty `dir` means run in the same directory, but if we just simply
	// interpolate arg for `cd` with an empty string, the result will change the
	// directory to the homedir. Let's prevent that, and replace it with PWD
	if dir == "" {
		dir = "$PWD"
	}
	full_invocation := fmt.Sprintf("(cd '%s' && %s)", dir, escapedInvocation)
	ex.shellScript = append(ex.shellScript, full_invocation)
	return nil
}

func (ex *DryRunExecutor) Output(name string, arg ...string) (string, error) {
	ex.Exec(name, arg...)
	return "", nil
}

func (ex *DryRunExecutor) GenerateShellScript() string {
	preamble := "#!/usr/bin/env sh\n"
	return strings.Join(append([]string{preamble}, ex.shellScript...), "\n")
}

func (ex *DryRunExecutor) GenerateDescription() string {
	return strings.Join(ex.invocations, "\n")
}
