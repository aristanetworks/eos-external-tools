package impl

import (
	"os"
	"path/filepath"
	"testing"

	"code.arista.io/eos/tools/eext/executor"
	"github.com/stretchr/testify/require"
)

func TestIsSigfileApplicable(t *testing.T) {
	testCases := []struct {
		tarballPath   string
		signaturePath string
		out1          bool
		out2          bool
	}{
		{"foo.tar.gz", "foo.tar.gz.sig", true, false},
		{"foo.tar.gz", "foo.tar.sig", true, true},
		{"foobar.tar.gz", "signature", false, false},
		{"foo.tar.gz", "bar.tar.gz.sig", false, false},
	}
	for _, tc := range testCases {
		res1, res2 := isSigfileApplicable(tc.tarballPath, tc.signaturePath)
		if res1 != tc.out1 || res2 != tc.out2 {
			t.Errorf("isSigfileApplicable for (%s, %s) -> (%t, %t); expected (%t, %t)",
				tc.tarballPath, tc.signaturePath, res1, res2, tc.out1, tc.out2)
		}
	}
}

func TestMatchTarballSignature(t *testing.T) {
	t.Log("Test tarball Signature Match")
	curPath, _ := os.Getwd()
	workingDir := filepath.Join(curPath, "testData/tarballSig/checkTarball")
	tarballSigPath := filepath.Join(workingDir, "libpcap-1.10.4.tar.gz.sig")
	tarballPath := filepath.Join(workingDir, "libpcap-1.10.4.tar.gz")
	intermediateTarball, err := matchTarballSignCmprsn(
		&executor.OsExecutor{},
		tarballPath,
		tarballSigPath,
		workingDir,
		"TestmatchTarballSignature : ",
	)
	os.Remove(intermediateTarball)
	require.NoError(t, err)
}

func getSHA256hash(folderName string, sha256InManifest string) error {
	sourceFile := "mrtparse-2.0.1.tar.gz"
	cwd, _ := os.Getwd()
	downloadDir := filepath.Join(cwd, folderName)
	srcFilePath := filepath.Join(downloadDir, sourceFile)
	checkSHA256HashErr := checkSHA256Hash(srcFilePath, sha256InManifest, "")
	return checkSHA256HashErr
}

func TestUpstreamSourcesBadSHA256Hash(t *testing.T) {
	sha256InManifest := "ac4456cda847db6a757fjhagdfjaghdfk3ee3b59e0d7ca224266aafa326163aec"
	err := getSHA256hash("testData/upstream-hash-check-bad", sha256InManifest)
	require.Error(t, err)
}

func TestUpstreamSourcesGoodSHA256Hash(t *testing.T) {
	sha256InManifest := "ac4456cda847db6a757f0c27cb09ad9b3ee3b59e0d7ca224266aafa326163aec"
	err := getSHA256hash("testData/upstream-hash-check-good", sha256InManifest)
	require.NoError(t, err)
}
