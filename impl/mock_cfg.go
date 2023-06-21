// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"fmt"
	"os"
	"path/filepath"

	"code.arista.io/eos/tools/eext/manifest"
	"code.arista.io/eos/tools/eext/repoconfig"
	"code.arista.io/eos/tools/eext/util"
)

// RepoData holds dnf repo name and baseurl for mock.cfg generation
type RepoData struct {
	Name    string
	BaseURL string
}

// MockCfgTemplateData is used to execute the mock config template
type MockCfgTemplateData struct {
	DefaultCommonCfg map[string]string
	Repo             []RepoData
	Includes         []string
}

type mockCfgBuilder struct {
	pkg               string
	repo              string
	isPkgSubdirInRepo bool
	arch              string
	buildSpec         *manifest.Build
	dnfRepoConfig     *repoconfig.DnfReposConfig
	errPrefix         util.ErrPrefix
	templateData      *MockCfgTemplateData
}

// populateTemplateData sets up the MockCfgTemplateData instance named templateData
// in cfgBldr for executing the template.
func (cfgBldr *mockCfgBuilder) populateTemplateData() error {

	pkg := cfgBldr.pkg
	arch := cfgBldr.arch

	cfgBldr.templateData = &MockCfgTemplateData{}
	cfgBldr.templateData.DefaultCommonCfg = map[string]string{
		"target_arch": arch,
		"root":        getMockChrootDirName(pkg, arch),
	}

	for _, repoSpecifiedInManifest := range cfgBldr.buildSpec.Repo {
		baseURL, err := cfgBldr.dnfRepoConfig.BaseURL(
			repoSpecifiedInManifest.Name,
			arch,
			repoSpecifiedInManifest.Version,
			repoSpecifiedInManifest.UseBaseArch,
			cfgBldr.errPrefix)

		if err != nil {
			return fmt.Errorf("%sError deriving baseURL: %s",
				cfgBldr.errPrefix, err)
		}
		repoData := RepoData{
			Name:    repoSpecifiedInManifest.Name,
			BaseURL: baseURL,
		}
		cfgBldr.templateData.Repo = append(cfgBldr.templateData.Repo, repoData)
	}

	mockCfgDir := getMockCfgDir(pkg, arch)
	if err := util.MaybeCreateDirWithParents(mockCfgDir, cfgBldr.errPrefix); err != nil {
		return err
	}

	// Includes in mock configuration will specify absolute paths.
	// It is expected that includes are copied over to the
	// same directory as the mock configuration file.
	for _, includeFile := range cfgBldr.buildSpec.Include {
		absoluteIncludeFilePathForMockCfg := filepath.Join(mockCfgDir, includeFile)
		cfgBldr.templateData.Includes = append(cfgBldr.templateData.Includes,
			absoluteIncludeFilePathForMockCfg)
	}
	return nil
}

// Create mock configuration directory
// Copy over any include files from source repo to mock configuration directory.
func (cfgBldr *mockCfgBuilder) prep() error {
	arch := cfgBldr.arch
	mockCfgDir := getMockCfgDir(cfgBldr.pkg, arch)
	if err := util.MaybeCreateDirWithParents(mockCfgDir, cfgBldr.errPrefix); err != nil {
		return err
	}

	for _, includeFile := range cfgBldr.buildSpec.Include {
		pkgDirInRepo := getPkgDirInRepo(cfgBldr.repo, cfgBldr.pkg, cfgBldr.isPkgSubdirInRepo)
		includeFilePath := filepath.Join(pkgDirInRepo, includeFile)
		if err := util.CheckPath(includeFilePath, false, false); err != nil {
			return fmt.Errorf("%sinclude file %s not found in repo",
				cfgBldr.errPrefix, pkgDirInRepo)
		}
		if err := util.CopyToDestDir(
			includeFilePath, mockCfgDir, cfgBldr.errPrefix); err != nil {
			return err
		}
	}
	return nil
}

// This executes the mock configuration template with the templateData
// seup previously and writes it to a file.
func (cfgBldr *mockCfgBuilder) createMockCfgFile() error {
	arch := cfgBldr.arch
	mockCfgPath := getMockCfgPath(cfgBldr.pkg, arch)
	mockCfgFileHandle, createErr := os.Create(mockCfgPath)
	if createErr != nil {
		return fmt.Errorf("%sError '%s' creating/truncating empty file %s for mock configuration",
			cfgBldr.errPrefix, createErr, mockCfgPath)
	}
	defer mockCfgFileHandle.Close()

	// parsedMockCfgTemplate is already expected to be setup
	if parsedMockCfgTemplate == nil {
		panic("parsedMockCfgTemplate is nil")
	}
	templateExecError := parsedMockCfgTemplate.Execute(mockCfgFileHandle, cfgBldr.templateData)
	if templateExecError != nil {
		return fmt.Errorf("%sError '%s' executing template with %s",
			cfgBldr.errPrefix, templateExecError, cfgBldr.templateData)
	}

	return nil
}
