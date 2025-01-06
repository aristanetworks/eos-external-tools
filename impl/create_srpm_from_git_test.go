// Copyright (c) 2023 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"os/exec"
	"testing"

	"code.arista.io/eos/tools/eext/executor"
	"code.arista.io/eos/tools/eext/util"
	"github.com/spf13/viper"
	//"code.arista.io/eos/tools/eext/executor/mocked_executor"
)

func TestRpmNameFromSpecFile(t *testing.T) {
	viper.Set("SrcDir", "testData/")
	defer viper.Reset()
	pkg := "libpcap"
	repo := "upstream-git-repo-1"
	expectedRpmName := "libpcap-1.10.1"

	mex := executor.MockedExecutor{
		Responses: []executor.Response{executor.NewResponse(0, "libpcap-1.10.1", nil)},
	}
	gotRpmName, err := getRpmNameFromSpecFile(repo, pkg, false, &mex)
	if err != nil {
		t.Fatal(err)
	}

	if expectedRpmName != gotRpmName {
		t.Fatalf("TestRpmNameFromSpecFile test failed. Expected: %s, Got %s",
			expectedRpmName, gotRpmName)

	}
	t.Log("Test rpmNameFromSpecFile PASSED")
}

func TestGitArchive(t *testing.T) {
	clonedDir := "/tmp/fake-cloned-dir"
	testWorkingDir := "/tmp/fake-working-dir"
	pkg := "libpcap"
	repo := "upstream-git-repo-1"
	revision := "libpcap-1.10.1"
	parentFolder := "libpcap-1.10.1"
	mex := executor.MockedExecutor{
		Responses: []executor.Response{executor.NewResponse(0, "libpcap-1.10.1", nil)},
	}

	_, err := generateArchiveFile(testWorkingDir, clonedDir, revision, repo, pkg, parentFolder, "", &mex)
	if err != nil {
		t.Fatal(err)
	}
	expectedCall := executor.NewRecordedCall("/tmp/fake-cloned-dir", "git",
		[]string{"archive", "--prefix", "libpcap-1.10.1/", "-o", "/tmp/fake-working-dir/Source0.tar.gz", "libpcap-1.10.1"})
	//expected := "In the directory '/tmp/fake-cloned-dir', would execute: git archive --prefix libpcap-1.10.1/ -o /tmp/fake-working-dir/Source0.tar.gz libpcap-1.10.1"
	actual := mex.Calls[0]
	if !mex.HasCall(expectedCall) {
		t.Fatalf("generateArchiveFile executed an unexpected command.\nExpected:\n%s\nActual:\n%s", expectedCall, actual)
	}

	t.Log("Test gitArchive PASSED")
}

func TestVerifyGitSignatureRevIsTagPassing(t *testing.T) {
	spec := gitSpec{
		SrcUrl:    "http://example.com",
		Revision:  "A-tag",
		ClonedDir: "/tmp/verification-dir",
	}
	pubKeyPath := "/tmp/pubkey"
	mex := executor.MockedExecutor{
		Responses: []executor.Response{
			executor.NewResponse(0, "", nil),            // for gpg --fingerprint
			executor.NewResponse(0, "", nil),            // for gpg --import
			executor.NewResponse(0, spec.Revision, nil), // for git show-ref
			executor.NewResponse(0, "", nil),            // for git verify-tag
		},
	}

	verifyGitSignature(pubKeyPath, spec, util.ErrPrefix(""), &mex)
	expectedCalls := []executor.RecordedCall{
		executor.NewRecordedCall("", "gpg", []string{"--fingerprint"}),
		executor.NewRecordedCall("", "gpg", []string{"--import", pubKeyPath}),
		executor.NewRecordedCall(spec.ClonedDir, "git", []string{"show-ref", "--quiet", "--tags", spec.Revision}),
		executor.NewRecordedCall(spec.ClonedDir, "git", []string{"verify-tag", "-v", spec.Revision}),
	}
	if !mex.HasExactCalls(expectedCalls) {
		t.Fatalf("verifyGitSignature executed wrong commands.\nExpected:\n%v\nActual:\n%v", expectedCalls, mex.Calls)
	}
}

func TestVerifyGitSignatureRevIsCommitPassing(t *testing.T) {
	spec := gitSpec{
		SrcUrl:    "http://example.com",
		Revision:  "A-tag",
		ClonedDir: "/tmp/verification-dir",
	}
	pubKeyPath := "/tmp/pubkey"
	mex := executor.MockedExecutor{
		Responses: []executor.Response{
			executor.NewResponse(0, "", nil),               // for gpg --fingerprint
			executor.NewResponse(0, "", nil),               // for gpg --import
			executor.NewResponse(1, "", &exec.ExitError{}), // for git show-ref
			executor.NewResponse(0, "", nil),               // for git cat-file
			executor.NewResponse(0, "", nil),               // for git verify-commit
		},
	}

	verifyGitSignature(pubKeyPath, spec, util.ErrPrefix(""), &mex)
	expectedCalls := []executor.RecordedCall{
		executor.NewRecordedCall("", "gpg", []string{"--fingerprint"}),
		executor.NewRecordedCall("", "gpg", []string{"--import", pubKeyPath}),
		executor.NewRecordedCall(spec.ClonedDir, "git", []string{"show-ref", "--quiet", "--tags", spec.Revision}),
		executor.NewRecordedCall(spec.ClonedDir, "git", []string{"cat-file", "-e", spec.Revision}),
		executor.NewRecordedCall(spec.ClonedDir, "git", []string{"verify-commit", "-v", spec.Revision}),
	}
	if !mex.HasExactCalls(expectedCalls) {
		t.Fatalf("verifyGitSignature executed wrong commands.\nExpected:\n%v\nActual:\n%v", expectedCalls, mex.Calls)
	}
}
