// Copyright (c) 2023 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package srcconfig

// SrcRepoParamsOverride spec
// Override default specs for source repo bundle in package manifest
type SrcRepoParamsOverride struct {
	VersionOverride   string `yaml:"version"`
	SrcSuffixOverride string `yaml:"src-suffix"`
	SigSuffixOverride string `yaml:"sig-suffix"`
}
