// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"fmt"
	"os"
	"path/filepath"

	"extbldr/util"
)

// Git clone the repository pointed by repoUrl to a new directory under
// named by pkg under dstBasePath. force indicates whether to overwrite an
// existing directory
func Clone(repoUrl string, dstBasePath string, pkg string, force bool) error {
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
	creatErr := util.MaybeCreateDir(dstBasePath)

	if creatErr != nil {
		return fmt.Errorf("impl.Clone: Creating %s errored out with %s", dstBasePath, creatErr)
	}

	cloneErr := util.RunSystemCmd("git", "clone", repoUrl, dstPath)
	if cloneErr != nil {
		return fmt.Errorf("impl.Clone: Cloning %s to %s errored out with %s",
			repoUrl, dstPath, cloneErr)
	}

	return nil
}
