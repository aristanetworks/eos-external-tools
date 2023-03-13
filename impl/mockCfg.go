// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/viper"

	"lemurbldr/manifest"
	"lemurbldr/util"
)

// TemplateParams struct exported for template usage
type TemplateParams struct {
	DefaultCommonCfg map[string]string
	RemoteRepo       []manifest.Repo
	Includes         []string
}

func loadCfgTemplate() (string, error) {
	buf, readError := os.ReadFile(viper.GetString("MockTemplate"))
	if readError != nil {
		return "", readError
	}
	return string(buf), readError
}

func setupCfgParams(arch string, target manifest.Target, pkg string, srcRepo string) (TemplateParams, error) {
	var templateParams TemplateParams
	workingBasePath := viper.GetString("WorkingDir")
	destPath := filepath.Join(workingBasePath, pkg)
	basePath := viper.GetString("SrcDir")
	srcPath := filepath.Join(basePath, srcRepo)
	templateParams.DefaultCommonCfg = map[string]string{
		"target_arch": arch,
		"root":        fmt.Sprintf("%s-{{ target_arch }}", pkg)}
	for _, manifestRepo := range target.Repo {
		templateParams.RemoteRepo = append(templateParams.RemoteRepo, manifestRepo)
	}
	for _, includeTpl := range target.Include {
		includeSrcPath := filepath.Join(srcPath, includeTpl)
		err := util.CopyFile("impl.mockCfg", includeSrcPath, destPath)
		if err != nil {
			return templateParams, err
		}
		includePath := filepath.Join(destPath, includeTpl)
		templateParams.Includes = append(templateParams.Includes, includePath)
	}
	return templateParams, nil
}

func generateCfg(arch string, target manifest.Target, pkg string, srcRepo string) (string, error) {
	templateParams, error := setupCfgParams(arch, target, pkg, srcRepo)
	if error != nil {
		return "", error
	}
	var buf bytes.Buffer
	templateString, readError := loadCfgTemplate()
	if readError != nil {
		return "", readError
	}
	t, templateCreateErr := template.New("").Parse(templateString)
	if templateCreateErr != nil {
		return "", templateCreateErr
	}
	templateExecErr := t.Execute(&buf, templateParams)
	return buf.String(), templateExecErr
}

func dumpToConfigFile(arch string, pkg string, cfg string) error {

	cfgFile, createErr := os.Create(getMockCfgPath(pkg, arch))
	if createErr != nil {
		return createErr
	}

	defer cfgFile.Close()

	_, writeErr := cfgFile.WriteString(cfg)

	if writeErr != nil {
		return writeErr
	}

	return nil
}

func mockCfgGenerate(arch string, pkg string, pkgManifest manifest.Package, srcRepo string) error {

	for _, pkgTarget := range pkgManifest.Target {
		thisTargetName := pkgTarget.Name
		if arch != thisTargetName {
			continue
		}

		mockCfgString, err := generateCfg(arch, pkgTarget, pkg, srcRepo)
		if err != nil {
			return err
		}
		return dumpToConfigFile(arch, pkg, mockCfgString)
	}

	return nil
}
