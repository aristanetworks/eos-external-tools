// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"golang.org/x/exp/maps"

	"code.arista.io/eos/tools/eext/dnfconfig"
	"code.arista.io/eos/tools/eext/manifest"
	"code.arista.io/eos/tools/eext/util"
)

// MockCfgTemplateData is used to execute the mock config template
type MockCfgTemplateData struct {
	DefaultCommonCfg map[string]string
	Macros           map[string]string
	Repo             []*dnfconfig.DnfRepoParams
	Includes         []string
}

// Common config used by both mockBuilder and mockCfgBuilder
type builderCommon struct {
	pkg               string
	repo              string
	isPkgSubdirInRepo bool
	arch              string
	rpmReleaseMacro   string
	eextSignature     string
	buildSpec         *manifest.Build
	dnfConfig         *dnfconfig.DnfConfig
	errPrefix         util.ErrPrefix
	dependencyList    []string
	enableNetwork     bool
}

type mockCfgBuilder struct {
	*builderCommon
	templateData *MockCfgTemplateData
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
		"resultdir":   getMockResultsDir(pkg, arch),
	}

	cfgBldr.templateData.Macros = make(map[string]string)
	if cfgBldr.rpmReleaseMacro != "" {
		cfgBldr.templateData.Macros["eext_release"] = cfgBldr.rpmReleaseMacro
	}

	if cfgBldr.eextSignature != "" {
		var eextsigTag string
		eextsigTag = fmt.Sprintf("eextsig=%s", cfgBldr.eextSignature)
		cfgBldr.templateData.Macros["distribution"] = eextsigTag
	}

	for _, repoBundleSpecifiedInManifest := range cfgBldr.buildSpec.RepoBundle {
		bundleName := repoBundleSpecifiedInManifest.Name
		bundleVersionOverride := repoBundleSpecifiedInManifest.VersionOverride
		bundleRepoOverrides := repoBundleSpecifiedInManifest.DnfRepoParamsOverride

		bundleConfig, found := cfgBldr.dnfConfig.DnfRepoBundleConfig[bundleName]
		if !found {
			return fmt.Errorf("%sUnknown repo-bundle name %s",
				cfgBldr.errPrefix, bundleName)
		}

		for overrideRepo, _ := range bundleRepoOverrides {
			_, isValidRepo := bundleConfig.DnfRepoConfig[overrideRepo]
			if !isValidRepo {
				return fmt.Errorf("%sBad repo-override %s specified in manifest",
					cfgBldr.errPrefix, overrideRepo)
			}
		}

		// Iterate in sorted order to make config file generation non-random
		reposInBundle := maps.Keys(bundleConfig.DnfRepoConfig)
		sort.Strings(reposInBundle)
		for _, repoName := range reposInBundle {
			repoParams, err := bundleConfig.GetDnfRepoParams(
				repoName,
				arch,
				bundleVersionOverride,
				bundleRepoOverrides,
				cfgBldr.errPrefix)
			if err != nil {
				return err
			}

			cfgBldr.templateData.Repo = append(cfgBldr.templateData.Repo, repoParams)
		}
	}

	if len(cfgBldr.dependencyList) != 0 {
		localRepo := &dnfconfig.DnfRepoParams{
			Name:     "local-deps",
			BaseURL:  "file://" + getMockDepsDir(cfgBldr.pkg, arch),
			Enabled:  true,
			GpgCheck: false,
			Priority: dnfconfig.RepoHighPriority,
		}
		cfgBldr.templateData.Repo = append(cfgBldr.templateData.Repo, localRepo)
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
		return fmt.Errorf("%sError '%s' executing template with %v",
			cfgBldr.errPrefix, templateExecError,
			*cfgBldr.templateData)
	}

	return nil
}
