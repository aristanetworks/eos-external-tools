package executor

import (
	"reflect"
)

// An executor that returns predefined, preprogrammed responses to call requests.
// It also records how it was called.
type MockedExecutor struct {
	// MockedExecutor returns the Response objects for each of the MockedExecutor
	// calls. The responses are consumed in the natural order, meaning [0] is the
	// first one to be returned.
	Responses []Response

	// records of what has been called on the executor
	Calls []RecordedCall
}

// RecordedCall stores information about what call was made
type RecordedCall struct {
	Dir  string
	Prog string
	Args []string
}

func NewRecordedCall(dir string, prog string, args []string) RecordedCall {
	return RecordedCall{Dir: dir, Prog: prog, Args: args}
}

// Response objects stores the mocked responses that are meant to be yielded by
// the executor
type Response struct {
	ReturnCode int
	Output     string
	Err        error
}

func NewResponse(code int, out string, err error) Response {
	return Response{ReturnCode: code, Output: out, Err: err}
}

// Mocked Exec that instead of running the command, it stores the information
// about the call, and returns predefined return code and an error object.
func (ex *MockedExecutor) Exec(name string, arg ...string) error {
	ex.Calls = append(ex.Calls, RecordedCall{"", name, arg})
	return ex.popResponse().Err
}

// Mocked ExecInDir that instead of running the command, it stores the information
// about the call, and returns predefined return code and an error object.
func (ex *MockedExecutor) ExecInDir(dir string, name string, arg ...string) error {
	ex.Calls = append(ex.Calls, RecordedCall{dir, name, arg})
	return ex.popResponse().Err
}

// Mocked Output that instead of running the command, it stores the information
// about the call, and returns predefined return code, the mocked Output that
// the command would have printed on the standard output, and an error object.
func (ex *MockedExecutor) Output(name string, arg ...string) (string, error) {
	ex.Calls = append(ex.Calls, RecordedCall{"", name, arg})
	response := ex.popResponse()
	return response.Output, response.Err
}

// Pops from the front one stored response and returns it
func (ex *MockedExecutor) popResponse() Response {
	if len(ex.Responses) == 0 {
		panic("Misused MockedExecutor! No responses left to yield")
	}
	popped := ex.Responses[0]
	ex.Responses = ex.Responses[1:]

	return popped
}

// Check if a call with exact params was made
func (ex *MockedExecutor) HasCall(call RecordedCall) bool {
	for _, recordedCall := range ex.Calls {
		if reflect.DeepEqual(call, recordedCall) {
			return true
		}
	}
	return false
}

// Check whether the calls made with the executor match exactly (and are in
// the same order)
func (ex *MockedExecutor) HasExactCalls(calls []RecordedCall) bool {
	return reflect.DeepEqual(calls, ex.Calls)
}
