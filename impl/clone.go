// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"fmt"
	"log"
	"os"

	"code.arista.io/eos/tools/eext/util"
)

// Clone git clones the repository pointed by repoURL to a new directory
// named repo under SrcDir.
// force indicates it is okay to overwrite an existing directory
func Clone(repoURL string, repo string, force bool) error {
	if err := CheckEnv(); err != nil {
		return err
	}
	repoDir := getRepoDir(repo)

	if util.CheckPath(repoDir, false, false) == nil && !force {
		return fmt.Errorf("impl.Clone: %s already exists, use --force to overwrite", repoDir)
	}

	if rmErr := os.RemoveAll(repoDir); rmErr != nil {
		return rmErr
	}

	var cloneErr error
	if util.GlobalVar.Quiet {
		cloneErr = util.RunSystemCmd("git", "clone", "--quiet", repoURL, repoDir)
	} else {
		cloneErr = util.RunSystemCmd("git", "clone", repoURL, repoDir)
	}

	if cloneErr != nil {
		return fmt.Errorf("impl.Clone: Cloning %s to %s errored out with %s",
			repoURL, repoDir, cloneErr)
	}

	log.Println("SUCCESS: clone")
	return nil
}
