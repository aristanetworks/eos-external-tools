// Copyright (c) 2023 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/spf13/viper"

	"golang.org/x/sys/unix"
)

var parsedMockCfgTemplate *template.Template

// CheckEnv validates environment setup and returns an error if any issue is found.
func CheckEnv() error {
	checks := map[string]struct {
		path          string
		isDir         bool
		writable      bool
		parseTemplate bool
	}{
		"SrcDir":          {viper.GetString("SrcDir"), true, false, false},
		"WorkingDir":      {viper.GetString("WorkingDir"), true, false, false},
		"DestDir":         {viper.GetString("DestDir"), true, true, false},
		"MockCfgTemplate": {viper.GetString("MockCfgTemplate"), false, false, true},
		"DnfConfigFile":   {viper.GetString("DnfConfigFile"), false, false, false},
		"SrcConfigFile":   {viper.GetString("SrcConfigFile"), false, false, false},
		"PkiPath":         {viper.GetString("PkiPath"), true, false, false},
	}

	var errors []string

	for name, check := range checks {
		if check.path == "" {
			continue
		}
		info, err := os.Stat(check.path)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", name, err))
			continue
		}
		if check.isDir && !info.IsDir() {
			errors = append(errors, fmt.Sprintf("%s is not a directory: %s", name, check.path))
		}
		if check.writable && unix.Access(check.path, unix.W_OK) != nil {
			errors = append(errors, fmt.Sprintf("%s is not writable: %s", name, check.path))
		}
		if check.parseTemplate && parsedMockCfgTemplate == nil {
			parsedMockCfgTemplate, err = template.ParseFiles(check.path)
			if err != nil {
				errors = append(errors, fmt.Sprintf("%s parsing error: %v", name, err))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("Environment check failed:\n%s", strings.Join(errors, "\n"))
	}

	if checks["SrcDir"].path == "" {
		if _, err := os.Stat("./eext.yaml"); err != nil {
			return fmt.Errorf("No eext.yaml found. SrcDir is unspecified, so sources are expected in the current working directory.")
		}
	}

	return nil
}
