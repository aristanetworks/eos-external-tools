// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"log"
)

// Build calls CreateSrpm and Mock in sequence
func Build(repo string, pkg string, arch string,
	extraCreateSrpmArgs CreateSrpmExtraCmdlineArgs,
	extraMockArgs MockExtraCmdlineArgs) error {
	if err := CreateSrpm(repo, pkg, extraCreateSrpmArgs); err != nil {
		return err
	}

	if err := Mock(repo, pkg, arch, extraMockArgs); err != nil {
		return err
	}
	log.Println("SUCCESS: Build")
	return nil
}
