package executor

import (
	"os/exec"
	"testing"
)

func TestMockedExecutorNoResponsesPrepared(t *testing.T) {
	// the defer below is to contain the panic and make the log clean
	defer func() { _ = recover() }()

	mex := MockedExecutor{}
	mex.Exec("true")
	// should panic before reaching this point
	t.Errorf("Did not panic on empty responses!")
}

func TestMockedExecutorExec(t *testing.T) {
	mex := MockedExecutor{
		Responses: []Response{{0, "", nil}},
	}
	err := mex.Exec("true")
	if err != nil {
		t.Fatalf("mocked Exec returned an error: %s", err)
	}
	if !mex.HasCall(RecordedCall{"", "true", nil}) {
		t.Fatal("MockedExecutor did not record the call!")

	}
}

func TestMockedExecutorExecFailing(t *testing.T) {
	mex := MockedExecutor{
		Responses: []Response{{1, "", &exec.ExitError{}}},
	}
	err := mex.Exec("false")
	if err == nil {
		t.Fatal("mocked Exec returned no error, but it should have!")
	}
	if !mex.HasCall(RecordedCall{"", "false", nil}) {
		t.Fatal("MockedExecutor did not record the call!")

	}
}

func TestMockedExecutorOutput(t *testing.T) {
	mex := MockedExecutor{
		Responses: []Response{{0, "Hello", nil}},
	}
	expected := "Hello"
	out, err := mex.Output("echo", "Hello")
	if err != nil {
		t.Fatalf("mocked Output returned an error: %s", err)
	}
	if out != "Hello" {
		t.Fatalf("Output returned unexpected output. Expected:\n%s\nGot:\n%s\n",
			expected, out)
	}
	if !mex.HasCall(RecordedCall{"", "echo", []string{"Hello"}}) {
		t.Fatal("MockedExecutor did not record the call!")

	}
}

func TestMockedExecutorExecInDir(t *testing.T) {
	mex := MockedExecutor{
		Responses: []Response{{0, "", nil}},
	}
	err := mex.ExecInDir("/tmp", "true")
	if err != nil {
		t.Fatalf("mocked Exec returned an error: %s", err)
	}
	if !mex.HasCall(RecordedCall{"/tmp", "true", nil}) {
		t.Fatal("MockedExecutor did not record the call!")
	}
}

func TestMockedExecutorMixedCalls(t *testing.T) {
	mex := MockedExecutor{
		Responses: []Response{
			{0, "", nil},
			{0, "", nil},
			{0, "", nil},
			{0, "", nil},
			{0, "", nil},
			{0, "", nil},
			{0, "", nil},
		},
	}
	mex.Exec("true")
	mex.Exec("with some spaces")
	mex.Exec("cat", "space-y file")
	mex.ExecInDir("/tmp", "pwd")
	mex.ExecInDir("/tmp", "cat", "difficult 'quoted' arg")
	mex.ExecInDir("", "false")
	mex.Output("cat", "file")
	expectedCalls := []RecordedCall{
		{"", "true", nil},
		{"", "with some spaces", nil},
		{"", "cat", []string{"space-y file"}},
		{"/tmp", "pwd", nil},
		{"/tmp", "cat", []string{"difficult 'quoted' arg"}},
		{"", "false", nil},
		{"", "cat", []string{"file"}},
	}
	for _, call := range expectedCalls {
		if !mex.HasCall(call) {
			t.Fatalf("MockedExecutor did not record the call! Missing call: %v", call)
		}
	}
}
