// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"fmt"
	"os"
	"path/filepath"

	"extbldr/util"
)

// Clone git clones the repository pointed by repoURL to a new directory under
// named by pkg under dstBasePath. force indicates whether to overwrite an
// existing directory
func Clone(repoURL string, dstBasePath string, pkg string, force bool) error {
	dstPath := filepath.Join(dstBasePath, pkg)
	_, statErr := os.Stat(dstPath)

	if statErr != nil && !os.IsNotExist(statErr) {
		// Error other than ENOENT
		return fmt.Errorf("impl.Clone: os.Stat on %s returned %s", dstPath, statErr)
	} else if statErr == nil {
		// No error, dstPath already exists
		if force {
			rmErr := util.RunSystemCmd("rm", "-rf", dstPath)
			if rmErr != nil {
				return fmt.Errorf("Removing %s errored out with %s", dstPath, rmErr)
			}
		} else {
			return fmt.Errorf("impl.Clone: %s already exists, use --force to overwrite", dstPath)
		}
	}

	// Create dstBasePath if required
	creatErr := util.MaybeCreateDir("impl.clone", dstBasePath)

	if creatErr != nil {
		return creatErr
	}
	var cloneErr error
	if util.GlobalVar.Quiet {
		cloneErr = util.RunSystemCmd("git", "clone", "--quiet", repoURL, dstPath)
	} else {
		cloneErr = util.RunSystemCmd("git", "clone", repoURL, dstPath)
	}

	if cloneErr != nil {
		return fmt.Errorf("impl.Clone: Cloning %s to %s errored out with %s",
			repoURL, dstPath, cloneErr)
	}

	return nil
}
