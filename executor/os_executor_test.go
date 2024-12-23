package executor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTrueRunsClean(t *testing.T) {
	ex := OsExecutor{}
	err := ex.Exec("true")
	if err != nil {
		t.Fatalf("`true` returned error: %v", err)
	}
}

func TestFalseReturnsError(t *testing.T) {
	ex := OsExecutor{}
	err := ex.Exec("false")
	if err == nil {
		t.Fatal("`false` did not return error")
	}
}

func TestOutputEchoEchoes(t *testing.T) {
	ex := OsExecutor{}
	// the `-n` in echo makes it not print the newline at the end
	// making the errors easier on the eyes when printed
	if out, err := ex.Output("echo", "-n", "Hello!"); err != nil {
		t.Fatalf("`echo Hello!` returned error: %v", err)
	} else {
		if out != "Hello!" {
			t.Fatalf("`echo Hello!` returned '%s' instead of Hello!", out)
		}
	}
}

func TestOutputPreservesStdErrOnFail(t *testing.T) {
	ex := OsExecutor{}
	// The bash invocation below spawns a nested shell session, with the last arg
	// being what to run in that sub-shell. This way the whole contents of this
	// one arg becomes the "text" to run inside the shell, enabling magic like
	// command chaining and redirecting to stderr
	_, err := ex.Output("bash", "-c", "echo err_msg >&2; false")
	if err != nil {
		if !strings.Contains(err.Error(), "err_msg") {
			t.Fatal("the error message was not forwarded")
		}
	} else {
		t.Fatal("Output() should have returned an error")
	}
}

func tempDirOrFatal(dir string, pattern string, t *testing.T) string {
	dir, err := os.MkdirTemp(dir, pattern)
	if err != nil {
		t.Fatal(err)
	}
	return dir
}
func TestExecInDirRunsThere(t *testing.T) {
	dir := tempDirOrFatal("", "OsExecutorTest", t)
	defer os.RemoveAll(dir)
	// we cannot be in (as in: have PWD pointing at) the newly created temp dir
	// so no need to check if startingCwd != dir
	startingCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Cannot get the current working directory. Error: %s", err)
	}
	ex := OsExecutor{}

	bash_line := "test `pwd` = " + dir
	if err := ex.ExecInDir(dir, "bash", "-c", bash_line); err != nil {
		t.Fatal("The test failed1", dir)
	}
	finalCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Cannot get the current working directory. Error: %s", err)
	}
	if startingCwd != finalCwd {
		t.Fatal("The working directory is different after running Exec")
	}

}
func TestExecInDirRunsThereCheckWithFile(t *testing.T) {
	dir := tempDirOrFatal("", "OsExecutorTest", t)
	defer os.RemoveAll(dir)
	_, err := os.Create(filepath.Join(dir, "payload"))
	if err != nil {
		t.Fatal(err)
	}
	ex := OsExecutor{}
	if err := ex.ExecInDir(dir, "ls", "payload"); err != nil {
		t.Fatalf("ls did not find the file that should be there. Error: %s", err)
	}
}

type shellEscapeTest struct {
	args     []string
	expected string
}

var shellEscapeTests = []shellEscapeTest{
	// Empty slice of args, empty string should be returned
	{[]string{}, ""},
	// Joins nothing - one thing, nothing should be changed
	{[]string{"true"}, "true"},
	// Joins nothing but escapes because of the space
	{[]string{"foo bar"}, "'foo bar'"},
	// Joins two things with no special escaping
	{[]string{"cat", "and_the_dog"}, "cat and_the_dog"},
	// Joins two things with one that has a space inside
	{[]string{"cat", "and the_dog"}, "cat 'and the_dog'"},
	// Joins three args with no spaces
	{[]string{"foo", "bar", "baz"}, "foo bar baz"},
	// Joins three args with spaces
	{[]string{"f o", "b r", "b z"}, "'f o' 'b r' 'b z'"},
	// Escapes an explicit quote mark
	{[]string{"b'r"}, "b\\'r"},
	// Escapes has multiple quote marks
	{[]string{"foo'bar'baz"}, "foo\\'bar\\'baz"},
	// Escapes spaces and quote marks
	{[]string{"foo 'bar' baz"}, "'foo \\'bar\\' baz'"},
}

func TestShellEscape(t *testing.T) {
	for _, tt := range shellEscapeTests {
		actual := shellEscape(tt.args)
		if actual != tt.expected {
			t.Fatalf("Expected '%s', got '%s'", tt.expected, actual)
		}
	}
}
