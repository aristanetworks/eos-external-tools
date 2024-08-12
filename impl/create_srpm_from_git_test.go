// Copyright (c) 2023 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"code.arista.io/eos/tools/eext/util"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

type TestDataType struct {
	gitSpec       *gitSpec
	expectedValue string
}

// We are currently using a tarball of the libpcap repo, and extracting it in a temp folder.
// This ensures that we mock 'cloneGitRepo' and steps after are tested.
// If we migrate to a remote repo, we can use this function to update the url.
func getSrcURL() string {
	url := "https://artifactory.infra.corp.arista.io/artifactory/eext-sources/eext-testData/libpcap.tar"
	return url
}

func downloadTarball(url, targetDir string) (string, error) {
	tarBallFilePath := filepath.Join(targetDir, "libpcap.tar")
	out, err := os.Create(tarBallFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %s", err)
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned: %s", resp.Status)
	}
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write to file: %s", err)
	}

	return tarBallFilePath, nil
}

// Gets the .tar file from test repo, and untars it into the required git repo.
func cloneRepoFromUrl(url, targetDir string) (string, error) {
	tarBallFilePath, err := downloadTarball(url, targetDir)
	if err != nil {
		return "", fmt.Errorf("failed to download tarball from %s: %s", url, err)
	}

	err = util.RunSystemCmdInDir(targetDir, "tar", "-xvf", tarBallFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to extract tarball %s: %s", tarBallFilePath, err)
	}

	clonedDir := filepath.Join(targetDir, "libpcap")
	fmt.Println(clonedDir)

	// suppress git error 128 (dubious ownership)
	user, err := util.CheckOutput("whoami", []string{}...)
	if err != nil {
		return "", fmt.Errorf("failed to get current user %s", err)
	}
	suppressCmdArgs := []string{"chown", "-R", strings.TrimSpace(user), clonedDir}
	err = util.RunSystemCmd("sudo", suppressCmdArgs...)
	if err != nil {
		return "", fmt.Errorf("failed to suppress git warning %s", err)
	}

	return clonedDir, nil
}

// A mock function for cloneGitRepo().
// Since we do not use remote git repo, this function downloads a tarball from a test repo,
// and expands it to be used as though we have cloned the repo from git.
func cloneGitDir() (string, error) {
	srcURL := getSrcURL()
	tempDir, err := os.MkdirTemp("", "upstream-git-test")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %s", err)
	}

	clonedDir, err := cloneRepoFromUrl(srcURL, tempDir)
	if err != nil {
		return "", fmt.Errorf("failed to clone repo dir from source %s at %s: %s", srcURL, tempDir, err)
	}

	return clonedDir, nil
}

func populateTestData(cloneDir string, revisionList, expectedList []string) []*TestDataType {
	// Not used in any tests currently, since we mock cloneGitRepo.
	// Will be usefull for testing gitSpec.typeOfGitRevisionFromRemote.
	srcURL := getSrcURL()

	var dataList []*TestDataType
	for i, revision := range revisionList {
		gitSpec := &gitSpec{
			SrcUrl:    srcURL,
			Revision:  revision,
			ClonedDir: cloneDir,
		}
		dataType := &TestDataType{
			gitSpec:       gitSpec,
			expectedValue: expectedList[i],
		}
		dataList = append(dataList, dataType)
	}

	return dataList
}

func populateTestDataForRevision(cloneDir string) []*TestDataType {
	revisionList := []string{"libpcap-1.10.4", "95691eb", "59747a7e74506bd2fbf6cc668e1d66b68ac6eb6d"}
	expectedList := []string{"TAG", "COMMIT", "COMMIT"}

	testData := populateTestData(cloneDir, revisionList, expectedList)

	// Required for testing typeOfGitRevisionFromRemote()
	// Keep disabled until we start using remote test data.
	/*for i, data := range testData {
		if i%2 == 0 {
			data.gitSpec.ClonedDir = ""
		}
	}*/

	return testData
}

func populateTestDataForGitSignature(cloneDir string) []*TestDataType {
	// Yet to verify commit signatures,
	// since not many commits signed with public keys are available.
	revisionList := []string{"libpcap-1.10.1"}
	expectedList := []string{""}

	return populateTestData(cloneDir, revisionList, expectedList)
}

func TestRevisionType(t *testing.T) {
	cloneDir, err := cloneGitDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(cloneDir)

	// To reuse populateTestData(), we convert the obtained GitRevisionType to string
	resolveGitRevisionTypeToString := []string{"UNDEFINED", "COMMIT", "TAG"}
	testDataList := populateTestDataForRevision(cloneDir)
	for _, data := range testDataList {
		gitSpec := data.gitSpec
		expectedType := data.expectedValue

		typeLocalRepo, err := gitSpec.typeOfGitRevision()
		if err != nil {
			t.Fatal(err)
		}

		// Test requires call to remote git repo.
		// Enable when we use remote repo for test.
		/*typeRemoteRepo, err := gitSpec.typeOfGitRevisionFromRemote()
		if err != nil {
			t.Fatal(err)
		}*/

		require.Equal(t, expectedType, resolveGitRevisionTypeToString[typeLocalRepo])
		//require.Equal(t, expectedType, resolveGitRevisionTypeToString[typeRemoteRepo])
	}
	t.Log("Test typeOfGitRevision PASSED")
}

func TestRpmNameFromSpecFile(t *testing.T) {
	viper.Set("SrcDir", "testData/")
	defer viper.Reset()
	pkg := "libpcap"
	repo := "upstream-git-repo-1"
	expectedRpmName := "libpcap-1.10.1"

	gotRpmName, err := getRpmNameFromSpecFile(repo, pkg, false)
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, expectedRpmName, gotRpmName)
	t.Log("Test rpmNameFromSpecFile PASSED")
}

func TestVerifyGitSignature(t *testing.T) {
	cloneDir, err := cloneGitDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(cloneDir)

	viper.Set("PkiPath", "../pki")
	defer viper.Reset()
	pubKeyPath := filepath.Join(getDetachedSigDir(), "tcpdump/tcpdumpPubKey.pem")
	testData := populateTestDataForGitSignature(cloneDir)
	for _, data := range testData {
		gitSpec := data.gitSpec

		err := verifyGitSignature(pubKeyPath, *gitSpec, "")
		if err != nil {
			t.Fatal(err)
		}
	}
	t.Log("Test verifyGitRepoSignature PASSED")
}

func TestGitArchive(t *testing.T) {
	clonedDir, err := cloneGitDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(clonedDir)

	testWorkingDir, mkdirErr := os.MkdirTemp("", "upstream-git")
	if mkdirErr != nil {
		t.Fatal(mkdirErr)
	}
	defer os.RemoveAll(testWorkingDir)

	viper.Set("SrcDir", "testData/")
	defer viper.Reset()
	pkg := "libpcap"
	repo := "upstream-git-repo-1"
	revision := "libpcap-1.10.1"

	archiveFile, err := generateArchiveFile(testWorkingDir, clonedDir, revision, repo, pkg, false, "")
	if err != nil {
		t.Fatal(err)
	}

	archivePath := filepath.Join(testWorkingDir, archiveFile)
	err = util.CheckPath(archivePath, false, false)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Test gitArchive PASSED")
}
