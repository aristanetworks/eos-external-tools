// Copyright (c) 2023 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package srcconfig

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"code.arista.io/eos/tools/eext/util"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// SrcRepoParamsOverride spec
// Override default specs for source repo bundle in package manifest
type SrcRepoParamsOverride struct {
	VersionOverride   string `yaml:"version"`
	SrcSuffixOverride string `yaml:"src-suffix"`
	SigSuffixOverride string `yaml:"sig-suffix"`
}

// SrcBundle spec
// To generate the source url for srpm
type SrcBundle struct {
	URLFormat         string            `yaml:"url-format"`
	DefaultSrcSuffix  string            `yaml:"default-src-suffix"`
	DefaultSigSuffix  string            `yaml:"default-sig-suffix"`
	VersionLabels     map[string]string `yaml:"version-labels"`
	HasDetachedSig    bool              `yaml:"has-detached-sig"`
	urlFormatTemplate *template.Template
}

// SrcConfig spec
// Mapping for all source repo bundles
type SrcConfig struct {
	SrcBundle map[string]*SrcBundle `yaml:"source-bundle"`
}

// SrcParams spec
// Updated source and signature url paths
type SrcParams struct {
	SrcURL       string
	SignatureURL string
}

func (s *SrcBundle) getTranslatedVersion(versionOverride string, errPrefix util.ErrPrefix) (string, error) {
	var version string
	if versionOverride == "" {
		version = "default"
	} else {
		version = versionOverride
	}
	translatedVersion, isVersionLabel := s.VersionLabels[version]
	if !isVersionLabel {
		translatedVersion = version
	}

	if translatedVersion == "default" {
		return "", fmt.Errorf("%s No defaults specified for source-bundle, please specify a version override",
			errPrefix)
	}

	return translatedVersion, nil
}

func (s *SrcBundle) getFormattedSigURL(srcFormattedURL, sigSuffixOverride string, onUncompressed bool) string {
	var sigFormattedURL string
	if s.HasDetachedSig {
		var sigSuffix string
		if sigSuffixOverride == "" {
			sigSuffix = s.DefaultSigSuffix
		} else {
			sigSuffix = sigSuffixOverride
		}

		var sigURLWithoutSuffix string
		if onUncompressed {
			sigURLWithoutSuffix = strings.TrimSuffix(srcFormattedURL, filepath.Ext(srcFormattedURL))
		} else {
			sigURLWithoutSuffix = srcFormattedURL
		}
		sigFormattedURL = sigURLWithoutSuffix + sigSuffix
	} else {
		sigFormattedURL = ""
	}
	return sigFormattedURL
}

func (s *SrcBundle) getSrcRepoParams(
	pkgName string,
	srcOverrideParams SrcRepoParamsOverride,
	onUncompressed bool,
	errPrefix util.ErrPrefix) (
	*SrcParams, error) {

	translatedVersion, err := s.getTranslatedVersion(srcOverrideParams.VersionOverride, errPrefix)
	if err != nil {
		return nil, err
	}

	var srcSuffix string
	if srcOverrideParams.SrcSuffixOverride == "" {
		srcSuffix = s.DefaultSrcSuffix
	} else {
		srcSuffix = srcOverrideParams.SrcSuffixOverride
	}
	urlData := struct {
		Host       string
		PathPrefix string
		PkgName    string
		Version    string
		Suffix     string
	}{
		Host:       viper.GetString("SrcRepoHost"),
		PathPrefix: viper.GetString("SrcRepoPathPrefix"),
		PkgName:    pkgName,
		Version:    translatedVersion,
		Suffix:     srcSuffix,
	}

	var urlBuf bytes.Buffer
	if err := s.urlFormatTemplate.Execute(&urlBuf, urlData); err != nil {
		return nil, fmt.Errorf("%sError executing template %s with data %v",
			errPrefix, s.URLFormat, urlData)
	}
	srcFormattedURL := urlBuf.String()

	sigFormattedURL := s.getFormattedSigURL(srcFormattedURL, srcOverrideParams.SigSuffixOverride, onUncompressed)

	return &SrcParams{
		SrcURL:       srcFormattedURL,
		SignatureURL: sigFormattedURL,
	}, nil
}

func formatURLTemplate(urlFormat string, errPrefix util.ErrPrefix) (string, error) {
	var urlBuf bytes.Buffer
	urlData := struct {
		Host       string
		PathPrefix string
	}{
		Host:       viper.GetString("SrcRepoHost"),
		PathPrefix: viper.GetString("SrcRepoPathPrefix"),
	}

	t, err := template.New("urlTemplate").Parse(urlFormat)
	if err != nil {
		return "", fmt.Errorf("%sError parsing source URL %s", errPrefix, urlFormat)
	}

	if err := t.Execute(&urlBuf, urlData); err != nil {
		return "", fmt.Errorf("%sError executing template %s with data %v",
			errPrefix, urlFormat, urlData)
	}

	return urlBuf.String(), nil
}

func getSrcParamsWithoutBundle(srcFullURL, sigFullURL string, errPrefix util.ErrPrefix) (*SrcParams, error) {
	srcURL, err := formatURLTemplate(srcFullURL, errPrefix)
	if err != nil {
		return nil, fmt.Errorf("%sUnable to format provided source url %s", errPrefix, srcFullURL)
	}
	sigURL, err := formatURLTemplate(sigFullURL, errPrefix)
	if err != nil {
		return nil, fmt.Errorf("%sUnable to format provided signature url %s", errPrefix, sigFullURL)
	}
	return &SrcParams{
		SrcURL:       srcURL,
		SignatureURL: sigURL,
	}, nil
}

func GetSrcParams(
	pkgName string,
	srcFullURL string,
	bundleName string,
	sigFullURL string,
	srcOverrideParams SrcRepoParamsOverride,
	onUncompressed bool,
	srcConfig *SrcConfig,
	errPrefix util.ErrPrefix) (
	*SrcParams, error) {

	var srcParams *SrcParams
	var srcParamsErr error
	if bundleName != "" {
		reqSrcBundle, ok := srcConfig.SrcBundle[bundleName]
		if !ok {
			return nil, fmt.Errorf("%sSource bundle %s not found in manifest", errPrefix, bundleName)
		}
		srcParams, srcParamsErr = reqSrcBundle.getSrcRepoParams(
			pkgName,
			srcOverrideParams,
			onUncompressed,
			errPrefix)
	} else {
		srcParams, srcParamsErr = getSrcParamsWithoutBundle(srcFullURL, sigFullURL, errPrefix)
	}

	if srcParamsErr != nil {
		return nil, srcParamsErr
	}
	return srcParams, nil
}

func LoadSrcConfig() (*SrcConfig, error) {
	cfgPath := viper.GetString("SrcConfigFile")
	_, statErr := os.Stat(cfgPath)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return nil, fmt.Errorf("srcconfig.LoadSrcConfig: %s doesn't exist",
				cfgPath)
		}
		return nil, fmt.Errorf("srcconfig.LoadSrcConfig: os.Stat on %s returned %s",
			cfgPath, statErr)
	}

	yamlContents, readErr := os.ReadFile(cfgPath)
	if readErr != nil {
		return nil, fmt.Errorf("srcconfig.LoadSrcConfig: os.ReadFile on %s returned %s",
			cfgPath, readErr)
	}

	var config SrcConfig
	if parseErr := yaml.UnmarshalStrict(yamlContents, &config); parseErr != nil {
		return nil, fmt.Errorf("srcconfig.LoadSrcConfig: Error parsing yaml file %s: %s",
			cfgPath, parseErr)
	}

	for bundleName, srcBundle := range config.SrcBundle {
		templateName := "srcRepoBundle_" + bundleName
		t, parseErr := template.New(templateName).Parse(srcBundle.URLFormat)
		if parseErr != nil {
			return nil, fmt.Errorf(
				"srcconfig.LoadSrcConfig: Error parsing source %s for src repo-bundle %s",
				srcBundle.URLFormat, bundleName)
		}
		srcBundle.urlFormatTemplate = t
	}
	return &config, nil
}
