// Copyright (c) 2023 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"code.arista.io/eos/tools/eext/manifest"
	"code.arista.io/eos/tools/eext/srcconfig"
	"code.arista.io/eos/tools/eext/util"
)

func (bldr *srpmBuilder) getUpstreamSourceForOthers(upstreamSrcFromManifest manifest.UpstreamSrc,
	downloadDir string) (*upstreamSrcSpec, error) {

	repo := bldr.repo
	pkg := bldr.pkgSpec.Name
	isPkgSubdirInRepo := bldr.pkgSpec.Subdir

	srcParams, err := srcconfig.GetSrcParams(
		pkg,
		upstreamSrcFromManifest.FullURL,
		upstreamSrcFromManifest.SourceBundle.Name,
		upstreamSrcFromManifest.Signature.DetachedSignature.FullURL,
		upstreamSrcFromManifest.SourceBundle.SrcRepoParamsOverride,
		upstreamSrcFromManifest.Signature.DetachedSignature.OnUncompressed,
		bldr.srcConfig,
		bldr.errPrefix)
	if err != nil {
		return nil, fmt.Errorf("%sUnable to get source params for %s",
			err, upstreamSrcFromManifest.SourceBundle.Name)
	}

	var downloadErr error
	upstreamSrc := upstreamSrcSpec{}

	upstreamSrcType := bldr.pkgSpec.Type
	bldr.log("downloading %s", srcParams.SrcURL)
	// Download source
	if upstreamSrc.sourceFile, downloadErr = download(
		srcParams.SrcURL,
		downloadDir,
		repo, pkg, isPkgSubdirInRepo,
		bldr.errPrefix); downloadErr != nil {
		return nil, downloadErr
	}
	bldr.log("downloaded")

	upstreamSrc.skipSigCheck = upstreamSrcFromManifest.Signature.SkipCheck
	pubKey := upstreamSrcFromManifest.Signature.DetachedSignature.PubKey

	if upstreamSrcType == "tarball" && !upstreamSrc.skipSigCheck {
		if srcParams.SignatureURL == "" || pubKey == "" {
			return nil, fmt.Errorf("%sNo detached-signature/public-key specified for upstream-sources entry %s",
				bldr.errPrefix, srcParams.SrcURL)
		}
		if upstreamSrc.sigFile, downloadErr = download(
			srcParams.SignatureURL,
			downloadDir,
			repo, pkg, isPkgSubdirInRepo,
			bldr.errPrefix); downloadErr != nil {
			return nil, downloadErr
		}

		pubKeyPath := filepath.Join(getDetachedSigDir(), pubKey)
		if pathErr := util.CheckPath(pubKeyPath, false, false); pathErr != nil {
			return nil, fmt.Errorf("%sCannot find public-key at path %s",
				bldr.errPrefix, pubKeyPath)
		}
		upstreamSrc.pubKeyPath = pubKeyPath
	} else if upstreamSrcType == "srpm" || upstreamSrcType == "unmodified-srpm" {
		// We don't expect SRPMs to have detached signature or
		// to be validated with a public-key specified in manifest.
		if srcParams.SignatureURL != "" {
			return nil, fmt.Errorf("%sUnexpected detached-sig specified for SRPM",
				bldr.errPrefix)
		}
		if pubKey != "" {
			return nil, fmt.Errorf("%sUnexpected public-key specified for SRPM",
				bldr.errPrefix)
		}
	}

	return &upstreamSrc, nil
}

// verifyRpmSignature verifies that the RPM specified at rpmPath
// is signed with a valid key in the key ring and that the signatures
// are valid.
func verifyRpmSignature(rpmPath string, errPrefix util.ErrPrefix) error {
	output, err := util.CheckOutput("rpm", "-K", rpmPath)
	if err != nil {
		return fmt.Errorf("%s:%s", errPrefix, err)
	}
	if !strings.Contains(output, "digests signatures OK") {
		return fmt.Errorf("%sSignature check of %s failed. rpm -K output:\n%s",
			errPrefix, rpmPath, output)
	}
	return nil
}

// checkValidSignature verifies that tarball anf signature
// correspond to same package
func checkValidSignature(tarballPath, tarballSigPath string) (
	bool, bool) {
	lastDotIndex := strings.LastIndex(tarballSigPath, ".")
	if lastDotIndex == -1 || !strings.HasPrefix(
		tarballPath, tarballSigPath[:lastDotIndex]) {
		return false, false
	}
	decompress := strings.Count(tarballPath[lastDotIndex:], ".")
	dcmprsnReqd := (decompress > 0)
	return true, dcmprsnReqd
}

// uncompressTarball decompresses the compression one layer at a time
// to match the tarball with its valid signature
func uncompressTarball(tarballPath string, downloadDir string) (string, error) {
	if err := util.RunSystemCmd(
		"7za", "x",
		"-y", tarballPath,
		"-o"+downloadDir); err != nil {
		return "", err
	}
	lastDotIndex := strings.LastIndex(tarballPath, ".")
	return tarballPath[:lastDotIndex], nil
}

// matchTarballSignCmprsn evaluvates and finds correct compressed/uncompressed tarball
// that matches with the sign file.
func matchTarballSignCmprsn(tarballPath string, tarballSigPath string,
	downloadDir string, errPrefix util.ErrPrefix) (string, error) {
	ok, dcmprsnReqd := checkValidSignature(tarballPath, tarballSigPath)
	if !ok {
		return "", fmt.Errorf("%sError while matching tarball and signature",
			errPrefix)
	}
	if dcmprsnReqd {
		newTarballPath, err := uncompressTarball(tarballPath, downloadDir)
		if err != nil {
			return "", fmt.Errorf("%sError '%s' while decompressing trarball",
				errPrefix, err)
		}
		return newTarballPath, nil
	}
	return "", nil
}

// VerifyTarballSignature verifies that the detached signature of the tarball
// is valid.
func verifyTarballSignature(
	tarballPath string, tarballSigPath string, pubKeyPath string,
	errPrefix util.ErrPrefix) error {
	tmpDir, mkdtErr := os.MkdirTemp("", "eext-keyring")
	if mkdtErr != nil {
		return fmt.Errorf("%sError '%s'creating temp dir for keyring",
			errPrefix, mkdtErr)
	}
	defer os.RemoveAll(tmpDir)

	keyRingPath := filepath.Join(tmpDir, "eext.gpg")
	baseArgs := []string{
		"--homedir", tmpDir,
		"--no-default-keyring", "--keyring", keyRingPath}
	gpgCmd := "gpg"

	// Create keyring
	createKeyRingCmdArgs := append(baseArgs, "--fingerprint")
	if err := util.RunSystemCmd(gpgCmd, createKeyRingCmdArgs...); err != nil {
		return fmt.Errorf("%sError '%s'creating keyring",
			errPrefix, err)
	}

	// Import public key
	importKeyCmdArgs := append(baseArgs, "--import", pubKeyPath)
	if err := util.RunSystemCmd(gpgCmd, importKeyCmdArgs...); err != nil {
		return fmt.Errorf("%sError '%s' importing public-key %s",
			errPrefix, err, pubKeyPath)
	}

	verifySigArgs := append(baseArgs, "--verify", tarballSigPath, tarballPath)
	if output, err := util.CheckOutput(gpgCmd, verifySigArgs...); err != nil {
		return fmt.Errorf("%sError verifying signature %s for tarball %s with pubkey %s."+
			"\ngpg --verify err: %sstdout:%s",
			errPrefix, tarballSigPath, tarballPath, pubKeyPath, err, output)
	}

	return nil
}