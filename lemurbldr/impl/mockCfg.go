// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"fmt"
	"os"
	"path/filepath"

	"lemurbldr/manifest"
	"lemurbldr/util"
)

// MockCfgTemplateData is used to execute the mock config template
type MockCfgTemplateData struct {
	DefaultCommonCfg map[string]string
	Repo             []manifest.Repo
	Includes         []string
}

// This sets up the MockCfgTemplateData instance
// for executing the template.
// It also copies over any include files to the relevant directory.
func setupTemplateData(
	repo string, pkg string, arch string,
	targetSpec manifest.Target,
	errPrefix util.ErrPrefix) (
	MockCfgTemplateData, error) {

	var templateData MockCfgTemplateData

	templateData.DefaultCommonCfg = map[string]string{
		"target_arch": arch,
		"root":        getMockChrootDirName(pkg, arch),
	}

	for _, repoSpecifiedInManifest := range targetSpec.Repo {
		templateData.Repo = append(templateData.Repo, repoSpecifiedInManifest)
	}

	mockCfgDir := getMockCfgDir(pkg, arch)
	for _, includeFile := range targetSpec.Include {
		// Any templates mentioned here should be copied to the
		// same directory as the mock configuration,
		// and they'll be included with absolute paths in the generated
		// mock configuration.
		includeFileSrcPath := filepath.Join(getRepoSrcDir(repo), includeFile)
		if util.CheckPath(includeFileSrcPath, false, false) != nil {
			return MockCfgTemplateData{},
				fmt.Errorf("%sCannot find the include file specified in manifest %s", errPrefix, includeFileSrcPath)
		}

		if err := util.CopyFile(includeFileSrcPath, mockCfgDir, errPrefix); err != nil {
			return MockCfgTemplateData{}, err
		}

		absoluteIncludeFilePathForMockCfg := filepath.Join(mockCfgDir, includeFile)
		templateData.Includes = append(templateData.Includes, absoluteIncludeFilePathForMockCfg)
	}
	return templateData, nil
}

func createMockCfgFile(repo string, pkgSpec manifest.Package, arch string) error {
	pkg := pkgSpec.Name

	errPrefix := util.ErrPrefix(fmt.Sprintf("impl.createMockCfgFile(%s): ", pkg))

	targetValid := false
	var targetSpec manifest.Target
	for _, targetSpec = range pkgSpec.Target {
		targetValid = true
		break
	}
	if !targetValid {
		return fmt.Errorf("%sTarget %s not found in manifest", errPrefix, arch)
	}

	templateData, stErr := setupTemplateData(repo, pkg, arch, targetSpec, errPrefix)
	if stErr != nil {
		return stErr
	}

	mockCfgPath := getMockCfgPath(pkg, arch)
	mockCfgFileHandle, createErr := os.Create(mockCfgPath)
	if createErr != nil {
		return fmt.Errorf("%sError '%s' creating/truncating empty file %s for mock configuration",
			errPrefix, createErr, mockCfgPath)
	}
	defer mockCfgFileHandle.Close()

	// parsedMockCfgTemplate is already expected to be setup
	templateExecError := parsedMockCfgTemplate.Execute(mockCfgFileHandle, templateData)
	if templateExecError != nil {
		return fmt.Errorf("%sError '%s' executing template with %s",
			errPrefix, templateExecError, templateData)
	}

	return nil
}
