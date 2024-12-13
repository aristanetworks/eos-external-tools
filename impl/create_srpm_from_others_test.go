package impl

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func testTarballSig(t *testing.T, folder string) {
	curPath, _ := os.Getwd()
	workingDir := filepath.Join(curPath, "testData/tarballSig", folder)
	tarballPath := map[string]string{
		"checkTarball": filepath.Join(workingDir, "linux.10.4.1.tar.gz"),
		"matchTarball": filepath.Join(workingDir, "libpcap-1.10.4.tar.gz"),
	}
	tarballSigPath := filepath.Join(workingDir, "libpcap-1.10.4.tar.gz.sig")

	switch folder {
	case "checkTarball":
		ok, _ := checkValidSignature(tarballPath[folder], tarballSigPath)
		require.Equal(t, false, ok)
	case "matchTarball":
		intermediateTarball, err := matchTarballSignCmprsn(
			tarballPath[folder],
			tarballSigPath,
			workingDir,
			"TestmatchTarballSignature : ",
		)
		os.Remove(intermediateTarball)
		require.Equal(t, nil, err)
	}
}

func TestCheckTarballSignature(t *testing.T) {
	t.Log("Test tarball Signatue Check")
	testTarballSig(t, "checkTarball")
}

func TestMatchTarballSignature(t *testing.T) {
	t.Log("Test tarball Signatue Match")
	testTarballSig(t, "matchTarball")
}

func getSHA256hash(folderName string, sha256InManifest string) error {
	sourceFile := "mrtparse-2.0.1.tar.gz"
	cwd, _ := os.Getwd()
	downloadDir := filepath.Join(cwd, folderName)
	CheckSHA256HashErr := checkSHA256Hash(downloadDir, sourceFile,
		sha256InManifest, "")
	return CheckSHA256HashErr
}

func TestUpstreamSourcesBadSHA256Hash(t *testing.T) {
	sha256InManifest := "ac4456cda847db6a757fjhagdfjaghdfk3ee3b59e0d7ca224266aafa326163aec"
	err := getSHA256hash("testData/upstream-hash-check-bad", sha256InManifest)
	require.NotEqual(t, nil, err)
}

func TestUpstreamSourcesGoodSHA256Hash(t *testing.T) {
	sha256InManifest := "ac4456cda847db6a757f0c27cb09ad9b3ee3b59e0d7ca224266aafa326163aec"
	err := getSHA256hash("testData/upstream-hash-check-good", sha256InManifest)
	require.Equal(t, nil, err)
}
