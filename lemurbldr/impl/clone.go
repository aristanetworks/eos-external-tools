// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"fmt"
	"os"

	"lemurbldr/util"
)

// Clone git clones the repository pointed by repoURL to a new directory
// named repo under SrcDir.
// force indicates it is okay to overwrite an existing directory
func Clone(repoURL string, repo string, force bool) error {
	if err := CheckEnv(); err != nil {
		return err
	}
	repoSrcDir := getRepoSrcDir(repo)

	if util.CheckPath(repoSrcDir, false, false) == nil && !force {
		return fmt.Errorf("impl.Clone: %s already exists, use --force to overwrite", repoSrcDir)
	}

	if rmErr := os.RemoveAll(repoSrcDir); rmErr != nil {
		return rmErr
	}

	var cloneErr error
	if util.GlobalVar.Quiet {
		cloneErr = util.RunSystemCmd("git", "clone", "--quiet", repoURL, repoSrcDir)
	} else {
		cloneErr = util.RunSystemCmd("git", "clone", repoURL, repoSrcDir)
	}

	if cloneErr != nil {
		return fmt.Errorf("impl.Clone: Cloning %s to %s errored out with %s",
			repoURL, repoSrcDir, cloneErr)
	}

	return nil
}
