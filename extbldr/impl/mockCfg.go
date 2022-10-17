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

	"extbldr/manifest"
	"extbldr/util"
)

// TemplateParams struct exported for template usage
type TemplateParams struct {
	DefaultCommonCfg map[string]string
	RemoteRepo       []manifest.Repo
	Includes         []string
}

const tmpl = `
config_opts['chroot_setup_cmd'] = "install bash bzip2 coreutils cpio diffutils redhat-release findutils gawk glibc-minimal-langpack grep gzip info patch redhat-rpm-config rpm-build sed shadow-utils tar unzip util-linux which xz"
config_opts['package_manager'] = "dnf"
config_opts['releasever'] = "9"
{{range $key,$val := .DefaultCommonCfg}}
config_opts['{{$key}}'] = "{{$val}}"
{{end}}

config_opts['dnf.conf'] = """
[main]
assumeyes=1
best=1
debuglevel=2
gpgcheck=0
install_weak_deps=0
keepcache=1
logfile=/var/log/yum.log
mdpolicy=group:primary
metadata_expire=0
module_platform_id=platform:el9
obsoletes=1
protected_packages=
reposdir=/dev/null
retries=20
syslog_device=
syslog_ident=mock


{{range .RemoteRepo}}
[{{.Name}}]
name = {{.Name}}
baseurl = {{.BaseURL}}
enabled = 1
{{end}}
"""
{{range .Includes}}
include("{{.}}")
{{end}}
`

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
	t, templateCreateErr := template.New("").Parse(tmpl)
	if templateCreateErr != nil {
		return "", templateCreateErr
	}
	templateExecErr := t.Execute(&buf, templateParams)
	return buf.String(), templateExecErr
}

// getCfgFilePath Returns the config file location under the workingDir/Package.
// File name of format mock_<pkg name>_<architecture>.cfg
func getCfgFilePath(arch string, pkg string) string {
	basePath := viper.GetString("WorkingDir")
	srcPath := filepath.Join(basePath, pkg)
	cfgFileName := fmt.Sprintf("mock_%s_%s.cfg", pkg, arch)
	cfgFilePath := filepath.Join(srcPath, cfgFileName)
	return cfgFilePath

}

func dumpToConfigFile(arch string, pkg string, cfg string) error {

	cfgFile, createErr := os.Create(getCfgFilePath(arch, pkg))
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
