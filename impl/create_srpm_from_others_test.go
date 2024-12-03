package impl

import (
	"os"
	"path/filepath"
	"testing"

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
		tarballPath,
		tarballSigPath,
		workingDir,
		"TestmatchTarballSignature : ",
	)
	os.Remove(intermediateTarball)
	require.NoError(t, err)
}
