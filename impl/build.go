// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"log"

	"code.arista.io/eos/tools/eext/executor"
)

// Build calls CreateSrpm and Mock in sequence
func Build(repo string, pkg string, arch string,
	extraCreateSrpmArgs CreateSrpmExtraCmdlineArgs,
	extraMockArgs MockExtraCmdlineArgs, executor executor.Executor) error {
	if err := CreateSrpm(repo, pkg, extraCreateSrpmArgs, executor); err != nil {
		return err
	}

	if err := Mock(repo, pkg, arch, extraMockArgs, executor); err != nil {
		return err
	}
	log.Println("SUCCESS: Build")
	return nil
}
