package executor

import (
	"testing"
)

func TestDryRunWithMixedCalls(t *testing.T) {
	drex := DryRunExecutor{}
	drex.Exec("true")
	drex.Exec("with some spaces")
	drex.Exec("cat", "space-y file")
	drex.ExecInDir("/tmp", "pwd")
	drex.ExecInDir("/tmp", "cat", "difficult 'quoted' arg")
	drex.ExecInDir("", "true")
	drex.Output("cat", "file")
	shellExpected := `#!/usr/bin/env sh

true
'with some spaces'
cat 'space-y file'
(cd '/tmp' && pwd)
(cd '/tmp' && cat 'difficult \'quoted\' arg')
(cd '$PWD' && true)
cat file`
	shellActual := drex.GenerateShellScript()
	if shellActual != shellExpected {
		t.Fatalf("GenerateShellScript generated unexpected output. Expected:\n%s\n\nGot:\n%s\n",
			shellExpected, shellActual)
	}
	descExpected := `Would execute: true
Would execute: 'with some spaces'
Would execute: cat 'space-y file'
In the directory '/tmp', would execute: pwd
In the directory '/tmp', would execute: cat 'difficult \'quoted\' arg'
In the directory '', would execute: true
Would execute: cat file`
	descActual := drex.GenerateDescription()
	if descActual != descExpected {
		t.Fatalf("GenerateDescription generated unexpected output. Expected:\n%s\n\nGot:\n%s\n",
			descExpected, descActual)
	}
}
