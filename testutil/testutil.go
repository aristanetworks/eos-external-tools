// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package testutil

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const ExpectedBuildHost string = "eext-buildhost"

// $ date -d @1710460800
// Fri Mar 15 00:00:00 UTC 2024
const MrtParseChangeLogTs string = "1710460800"

// $ date -d @1628467200
// Mon Aug  9 00:00:00 UTC 2021
const DebugeditChangeLogTs string = "1628467200"

var r, w, rescueStdout *(os.File)

// SetupManifest used to setup a test manifest from testdata for manifest functionality testing
func SetupManifest(t *testing.T, baseDir string, pkg string, sampleFile string) {
	pkgDir := filepath.Join(baseDir, pkg)
	os.RemoveAll(pkgDir)
	os.Mkdir(pkgDir, 0775)

	sampleManifestPath := filepath.Join("testData", sampleFile)
	_, statErr := os.Stat(sampleManifestPath)
	if statErr != nil {
		t.Fatal(statErr)
	}

	targetPath, absErr := filepath.Abs(sampleManifestPath)
	if absErr != nil {
		t.Fatal(absErr)
	}
	linkPath := filepath.Join(pkgDir, "eext.yaml")
	symlinkErr := os.Symlink(targetPath, linkPath)
	if symlinkErr != nil {
		t.Fatal(symlinkErr)
	}
}

// SetupDummyRpm used to setup a test upstream srpm
func SetupDummyRpm(t *testing.T,
	targetDir, pkg, arch, upstreamVersion, upstreamRelease, specFileReleaseLine string,
	buildRequires, requires []string,
	isSource bool) {

	specFileTemplateStr := `
Summary: Dummy package
Name: {{.Name}}
Version: {{.UpstreamVersion}}
{{ .SpecFileReleaseLine }}
BuildArch: {{.Arch}}
License: {{.Name}}

{{ range .Requires }}
Requires: {{.}}
{{ end -}}

{{ range .BuildRequires }}
BuildRequires: {{.}}
{{ end -}}


%description
{{.Name}}

%prep
true

%build
true

%install
true

%files
`
	tmpl, tmplErr := template.New("specTemplate").Parse(specFileTemplateStr)
	if tmplErr != nil {
		t.Fatal(tmplErr)
	}

	workdir, workdirErr := os.MkdirTemp("", "rpmbuild")
	if workdirErr != nil {
		t.Fatal(workdirErr)
	}
	// defer os.RemoveAll(workdir)

	resultsDir := filepath.Join(workdir, "results")
	resultsSpecsDir := filepath.Join(workdir, "results", "SPECS")
	for _, dir := range []string{resultsDir, resultsSpecsDir} {
		if err := os.Mkdir(dir, 0775); err != nil {
			t.Fatal(err)
		}
	}

	specFilePath := filepath.Join(resultsSpecsDir, pkg+".spec")
	specFileHandle, createErr := os.Create(specFilePath)
	if createErr != nil {
		t.Fatal(createErr)
	}

	data := struct {
		Name                string
		Arch                string
		UpstreamVersion     string
		SpecFileReleaseLine string
		Requires            []string
		BuildRequires       []string
	}{pkg, arch, upstreamVersion, specFileReleaseLine,
		requires, buildRequires}

	if tmplExecErr := tmpl.Execute(specFileHandle, data); tmplExecErr != nil {
		t.Fatal(tmplExecErr)
	}
	specFileHandle.Close()

	rpmbuildCmdOptions := []string{
		"--define", fmt.Sprintf("_topdir %s", resultsDir),
		"-ba",
		specFilePath}
	rpmbuildCmd := exec.Command("rpmbuild", rpmbuildCmdOptions...)
	rpmbuildCmd.Stderr = os.Stderr
	rpmbuildCmd.Stdout = os.Stdout

	t.Logf("Running rpmbuild %s to setup a dummy rpm", rpmbuildCmdOptions)
	if rbErr := rpmbuildCmd.Run(); rbErr != nil {
		t.Fatal(fmt.Errorf("rpmbuild failed with %s", rbErr))
	}
	t.Logf("Upstream srpm setup")
	var builtRpmSubdir string
	var builtRpmExtension string
	if isSource {
		builtRpmSubdir = "SRPMS"
		builtRpmExtension = "src.rpm"
	} else {
		builtRpmSubdir = filepath.Join("RPMS", arch)
		builtRpmExtension = fmt.Sprintf("%s.rpm", arch)
	}
	builtRpmFilename := fmt.Sprintf("%s-%s-%s.%s", pkg, upstreamVersion, upstreamRelease, builtRpmExtension)
	builtRpmPath := filepath.Join(resultsDir, builtRpmSubdir, builtRpmFilename)
	targetPath := filepath.Join(targetDir, builtRpmFilename)
	linkErr := os.Link(builtRpmPath, targetPath)
	if linkErr != nil {
		t.Fatal(linkErr)
	}
}

// RunCmd runs the command in cobra cmd and returns error
func RunCmd(t *testing.T, rootCmd *cobra.Command, args []string, quiet bool, expectSuccess bool) error {
	if quiet {
		args = append(args, "--quiet")
		setupQuiet()
	}

	rootCmd.SetArgs(args)
	t.Logf("Running cmd with args: %v\n", args)
	cmdErr := rootCmd.Execute()

	if expectSuccess {
		t.Log("Expecting success.")
		require.NoError(t, cmdErr)
		if quiet {
			checkAndCleanupQuiet(t)
		}
	}
	return cmdErr

}

func setupQuiet() {
	rescueStdout = os.Stdout
	r, w, _ = os.Pipe()
	os.Stdout = w
}

func checkAndCleanupQuiet(t *testing.T) {
	w.Close()
	out, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	require.Empty(t, out)
	os.Stdout = rescueStdout
}

func SetupSrcEnv(src []string) {
	envPrefix := viper.GetString("SrcEnvPrefix")
	for index, val := range src {
		varName := envPrefix + strconv.Itoa(index)
		err := os.Setenv(varName, val)
		if err != nil {
			panic("Setenv failed")
		}
	}
}

func CleanupSrcEnv(src []string) {
	envPrefix := viper.GetString("SrcEnvPrefix")
	for index, _ := range src {
		varName := envPrefix + strconv.Itoa(index)
		err := os.Unsetenv(varName)
		if err != nil {
			panic("Unsetenv failed")
		}
	}
}

// SetupViperConfig sets up the viper config for the test
func SetupViperConfig(
	srcDir string,
	workingDir string,
	destDir string,
	srpmsDir string,
	depsDir string,
	repoHost string,
	dnfConfigFile string,
	srcRepoHost string,
	srcConfigFile string,
	srcRepoPathPrefix string,
) {
	if srcDir == "" {
		viper.Set("SrcDir", "testData")
	} else {
		viper.Set("SrcDir", srcDir)
	}
	viper.Set("WorkingDir", workingDir)
	viper.Set("DestDir", destDir)
	viper.Set("DepsDir", depsDir)
	if repoHost == "" {
		viper.Set("DnfRepoHost",
			"https://artifactory.infra.corp.arista.io")
	} else {
		viper.Set("DnfRepoHost", repoHost)
	}
	if dnfConfigFile == "" {
		viper.Set("DnfConfigFile",
			"../configfiles/dnfconfig.yaml")
	} else {
		viper.Set("DnfConfigFile", dnfConfigFile)
	}
	if srcRepoHost == "" {
		viper.Set("SrcRepoHost",
			"https://artifactory.infra.corp.arista.io")
	} else {
		viper.Set("SrcRepoHost", srcRepoHost)
	}
	if srcConfigFile == "" {
		viper.Set("SrcConfigFile",
			"../configfiles/srcconfig.yaml")
	} else {
		viper.Set("SrcConfigFile", srcConfigFile)
	}
	if srcRepoPathPrefix == "" {
		viper.Set("SrcRepoPathPrefix",
			"artifactory/eext-sources")
	} else {
		viper.Set("SrcRepoPathPrefix", srcRepoPathPrefix)
	}
	viper.Set("MockCfgTemplate",
		"../configfiles/mock.cfg.template")
	viper.Set("PkiPath",
		"../pki")
	// Don't user the default of SRC_ to make sure that
	// the test works in a barney context
	viper.Set("SrcEnvPrefix",
		"XXXSRC_")
	viper.Set("SrpmsDir", srpmsDir)
}

// CheckEnv panics if the test hasn't setup the environment correctly
func CheckEnv(t *testing.T, rootCmd *cobra.Command) {
	_ = RunCmd(t, rootCmd, []string{"checkenv"}, false, true)
	t.Log("Test environment fine")
}

// SetupDummyDependency sets up an empty rpm file at the path specified
func SetupDummyDependency(t *testing.T,
	depsDir, depPkg, depPkgArch, depVersion, depRelease string) {
	// Create dep dir path
	depPkgDirWithArch := filepath.Join(depsDir, depPkgArch, depPkg)
	if err := os.MkdirAll(depPkgDirWithArch, 0755); err != nil {
		t.Fatal(err)
	}
	releaseLine := fmt.Sprintf("Release: %s", depRelease)
	SetupDummyRpm(t, depPkgDirWithArch,
		depPkg, depPkgArch,
		depVersion, depRelease, releaseLine,
		nil,   // buildRequires
		nil,   // requires
		false, // isSource
	)
}
