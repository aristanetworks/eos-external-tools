package executor

type Executor interface {
	// Execute a command in the current working directory.
	Exec(name string, arg ...string) error

	// Execute a command in a directory specified by `dir`.
	// The current working directory is not changed after the call.
	ExecInDir(dir string, name string, arg ...string) error

	// Execute a command and capture its standard output.
	// If the command returns 0, the output is returned.
	// Should the command return a non-zero code, the standard error is embedded
	// in the error object's error message.
	Output(name string, arg ...string) (string, error)
}
