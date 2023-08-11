// Copyright (c) 2023 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"code.arista.io/eos/tools/eext/dnfconfig"
	"code.arista.io/eos/tools/eext/manifest"
	"code.arista.io/eos/tools/eext/testutil"
	"code.arista.io/eos/tools/eext/util"
)

func TestMockConfig(t *testing.T) {
	t.Log("Create temporary working directory")
	testWorkingDir, err := os.MkdirTemp("", "mock-cfg-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testWorkingDir)

	srcDir := filepath.Join(testWorkingDir, "src")
	workDir := filepath.Join(testWorkingDir, "work")
	destDir := filepath.Join(testWorkingDir, "dest")

	for _, subdir := range []string{srcDir, workDir, destDir} {
		os.Mkdir(subdir, 0775)
	}

	t.Log("Copy testData/manifest to src directory")
	pkg := "pkg1"
	testutil.SetupManifest(t, srcDir, pkg, "manifest.yaml")

	testutil.SetupViperConfig(
		srcDir,
		workDir,
		destDir,
		"https://foo.org",         // repoHost
		"testData/dnfconfig.yaml", //dnfConfigFile
	)
	defer viper.Reset()

	t.Log("Loading manifest")
	manifestObj, err := manifest.LoadManifest(pkg)
	require.NoError(t, err)
	require.NotNil(t, manifestObj)

	t.Log("Load dnfconfig.yaml")
	dnfConfig, loadErr := dnfconfig.LoadDnfConfig()
	require.NoError(t, loadErr)
	require.NotNil(t, dnfConfig)

	cfgBldr := mockCfgBuilder{
		builderCommon: &builderCommon{
			pkg:             pkg,
			arch:            "x86_64",
			rpmReleaseMacro: "my-release",
			eextSignature:   "my-signature",
			buildSpec:       &manifestObj.Package[0].Build,
			dnfConfig:       dnfConfig,
		},
	}

	envErr := CheckEnv()
	require.NoError(t, envErr)

	populateErr := cfgBldr.populateTemplateData()
	require.NoError(t, populateErr)

	prepErr := cfgBldr.prep()
	require.NoError(t, prepErr)

	createErr := cfgBldr.createMockCfgFile()
	require.NoError(t, createErr)

	outFilePath := filepath.Join(workDir, pkg, "mock-x86_64/mock-cfg/mock.cfg")
	require.FileExists(t, outFilePath)

	var expectedMockCfgTemplate *template.Template
	var parseErr error
	if expectedMockCfgTemplate, parseErr = template.ParseFiles("testData/expected-mock.cfg"); parseErr != nil {
		panic("Failed to parse testData/expected-mock.cfg")
	}
	generatedExpectedMockCfgPath := filepath.Join(testWorkingDir, "expected-mock.cfg")
	generatedExpectedMockCfgFileHandle, createErr := os.Create(generatedExpectedMockCfgPath)
	if createErr != nil {
		panic("Failed to create empty file for generating expected mock configuration")
	}

	if templateExecError := expectedMockCfgTemplate.Execute(
		generatedExpectedMockCfgFileHandle,
		struct{ TestWorkingDir string }{
			testWorkingDir,
		}); templateExecError != nil {
		panic(fmt.Sprintf("Error %s executing expectedMockCfgTemplate template",
			templateExecError))
	}
	generatedExpectedMockCfgFileHandle.Close()

	if diffErr := util.RunSystemCmd("diff", "-u", generatedExpectedMockCfgPath, outFilePath); diffErr != nil {
		t.Errorf("Mock configuration differes from expected one: diff -u %s %s failed",
			generatedExpectedMockCfgPath, outFilePath)
	}
}
