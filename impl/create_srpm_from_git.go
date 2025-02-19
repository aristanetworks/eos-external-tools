// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package impl

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"code.arista.io/eos/tools/eext/executor"
	"code.arista.io/eos/tools/eext/manifest"
	"code.arista.io/eos/tools/eext/srcconfig"
	"code.arista.io/eos/tools/eext/util"
)

type gitSpec struct {
	SrcUrl    string
	Revision  string
	ClonedDir string
}

func getRpmNameFromSpecFile(repo, pkg string, isPkgSubdirInRepo bool) (string, error) {
	pkgSpecDirInRepo := getPkgSpecDirInRepo(repo, pkg, isPkgSubdirInRepo)
	specFiles, _ := filepath.Glob(filepath.Join(pkgSpecDirInRepo, "*.spec"))
	numSpecFiles := len(specFiles)
	if numSpecFiles == 0 {
		return "", fmt.Errorf("no *.spec files found in %s", pkgSpecDirInRepo)
	}
	if numSpecFiles > 1 {
		return "", fmt.Errorf("multiple *.spec files %s found in %s", strings.Join(specFiles, ","), pkgSpecDirInRepo)
	}
	specFilePath := specFiles[0]

	cmd := []string{"-q", "--srpm", "--qf", "%{NAME}-%{VERSION}", specFilePath}
	rpmName, err := util.CheckOutput("rpmspec", cmd...)
	if err != nil {
		return "", fmt.Errorf("cannot query spec file %s for %s", specFilePath, pkg)
	}

	return rpmName, nil
}

// We aren't using 'git clone' since it is slow for large repos.
// This method is faster and only pulls necessary changes.
func cloneGitRepo(executor executor.Executor, pkg, srcURL, revision, targetDir string) (string, error) {
	// Cloning the git repo to a temporary directory
	cloneDir, err := os.MkdirTemp(targetDir, pkg)
	if err != nil {
		return "", fmt.Errorf("error while creating tempDir for %s, %s", pkg, err)
	}
	// Init the dir as a git repo

	err = executor.ExecInDir(cloneDir, "git", "init")
	if err != nil {
		return "", fmt.Errorf("git init at %s failed: %s", cloneDir, err)
	}
	// Add the srcURL as the origin for the repo
	err = executor.ExecInDir(cloneDir, "git", "remote", "add", "origin", srcURL)
	if err != nil {
		return "", fmt.Errorf("adding %s as git remote failed: %s", srcURL, err)
	}
	// Fetch repo tags, for user inputs revision as TAG
	err = executor.ExecInDir(cloneDir, "git", "fetch", "--tags")
	if err != nil {
		return "", fmt.Errorf("fetching tags failed for %s: %s", pkg, err)
	}
	// Fetch the code changes for the provided revision

	err = executor.ExecInDir(cloneDir, "git", "fetch", "origin", revision)
	if err != nil {
		return "", fmt.Errorf("fetching revision %s failed for %s: %s", revision, pkg, err)
	}
	// Pull code to repo at provided revision

	err = executor.ExecInDir(cloneDir, "git", "reset", "--hard", "FETCH_HEAD")
	if err != nil {
		return "", fmt.Errorf("fetching HEAD at %s failed: %s", revision, err)
	}

	return cloneDir, nil
}

func generateArchiveFile(executor executor.Executor, targetDir, clonedDir, revision, repo, pkg string, isPkgSubdirInRepo bool,
	errPrefix util.ErrPrefix) (string, error) {
	// User should ensure the same fileName is specified in .spec file.
	// We use Source0.tar.gz as the generated tarball path,
	// since this can be extended to support multiple sources in future.
	gitArchiveFile := "Source0.tar.gz"
	gitArchiveFilePath := filepath.Join(targetDir, gitArchiveFile)
	parentFolder, err := getRpmNameFromSpecFile(repo, pkg, isPkgSubdirInRepo)
	if err != nil {
		return "", err
	}

	// Create the tarball from the specified commit/tag revision
	archiveCmd := []string{"archive",
		"--prefix", parentFolder + "/",
		"-o", gitArchiveFilePath,
		revision,
	}

	err = executor.ExecInDir(clonedDir, "git", archiveCmd...)
	if err != nil {
		return "", fmt.Errorf("%sgit archive of %s failed: %s %v", errPrefix, pkg, err, archiveCmd)
	}

	return gitArchiveFile, nil
}

// Download the git repo, and create a tarball at the provided commit/tag.
func archiveGitRepo(executor executor.Executor, srcURL, targetDir, revision, repo, pkg string, isPkgSubdirInRepo bool,
	errPrefix util.ErrPrefix) (string, string, error) {
	cloneDir, err := cloneGitRepo(executor, pkg, srcURL, revision, targetDir)
	if err != nil {
		return "", "", fmt.Errorf("cloning git repo failed: %s", err)
	}

	gitArchiveFile, err := generateArchiveFile(executor, targetDir, cloneDir, revision, repo, pkg, isPkgSubdirInRepo, errPrefix)
	if err != nil {
		return "", "", fmt.Errorf("generating git archive failed: %s", err)
	}

	return gitArchiveFile, cloneDir, nil
}

func getGitSpecAndSrcFile(executor executor.Executor, srcUrl, revision, downloadDir, repo, pkg string,
	isPkgSubdirInRepo bool, errPrefix util.ErrPrefix) (*gitSpec, string, error) {
	spec := gitSpec{
		SrcUrl:   srcUrl,
		Revision: revision,
	}

	sourceFile, clonedDir, downloadErr := archiveGitRepo(
		executor,
		srcUrl,
		downloadDir,
		revision,
		repo, pkg, isPkgSubdirInRepo,
		errPrefix)
	if downloadErr != nil {
		return nil, "", downloadErr
	}

	spec.ClonedDir = clonedDir
	return &spec, sourceFile, nil
}

func (bldr *srpmBuilder) getUpstreamSourceForGit(upstreamSrcFromManifest manifest.UpstreamSrc,
	downloadDir string) (*upstreamSrcSpec, error) {

	repo := bldr.repo
	pkg := bldr.pkgSpec.Name
	isPkgSubdirInRepo := bldr.pkgSpec.Subdir

	srcParams, err := srcconfig.GetSrcParams(
		pkg,
		upstreamSrcFromManifest.GitBundle.Url,
		upstreamSrcFromManifest.SourceBundle.Name,
		upstreamSrcFromManifest.Signature.DetachedSignature.FullURL,
		upstreamSrcFromManifest.SourceBundle.SrcRepoParamsOverride,
		upstreamSrcFromManifest.Signature.DetachedSignature.OnUncompressed,
		bldr.srcConfig,
		bldr.errPrefix)
	if err != nil {
		return nil, fmt.Errorf("%sunable to get source params for %s",
			err, upstreamSrcFromManifest.SourceBundle.Name)
	}

	upstreamSrc := upstreamSrcSpec{}

	bldr.log("creating tarball for %s from repo %s", pkg, srcParams.SrcURL)
	srcUrl := srcParams.SrcURL
	revision := upstreamSrcFromManifest.GitBundle.Revision
	spec, sourceFile, err := getGitSpecAndSrcFile(bldr.executor, srcUrl, revision, downloadDir,
		repo, pkg, isPkgSubdirInRepo, bldr.errPrefix)
	if err != nil {
		return nil, err
	}
	bldr.log("tarball created")

	upstreamSrc.gitSpec = *spec
	upstreamSrc.sourceFile = sourceFile
	upstreamSrc.skipSigCheck = upstreamSrcFromManifest.Signature.SkipCheck
	pubKey := upstreamSrcFromManifest.Signature.DetachedSignature.PubKey

	if !upstreamSrc.skipSigCheck {
		if pubKey == "" {
			return nil, fmt.Errorf("%sexpected public-key for %s to verify git repo",
				bldr.errPrefix, pkg)
		}
		pubKeyPath := filepath.Join(getDetachedSigDir(), pubKey)
		if pathErr := util.CheckPath(pubKeyPath, false, false); pathErr != nil {
			return nil, fmt.Errorf("%sCannot find public-key at path %s",
				bldr.errPrefix, pubKeyPath)
		}
		upstreamSrc.pubKeyPath = pubKeyPath
	}

	return &upstreamSrc, nil
}

// verifyGitSignature verifies that the git repo commit/tag is signed.
func verifyGitSignature(executor executor.Executor, pubKeyPath string, gitSpec gitSpec, errPrefix util.ErrPrefix) error {
	tmpDir, mkdtErr := os.MkdirTemp("", "eext-keyring")
	if mkdtErr != nil {
		return fmt.Errorf("%sError '%s'creating temp dir for keyring",
			errPrefix, mkdtErr)
	}
	defer os.RemoveAll(tmpDir)

	err := os.Setenv("GNUPGHOME", tmpDir)
	if err != nil {
		return fmt.Errorf("%sunable to set ENV variable GNUPGHOME", errPrefix)
	}
	defer os.Unsetenv("GNUPGHOME")

	if err := executor.Exec("gpg", "--fingerprint"); err != nil {
		return fmt.Errorf("%sError '%s'creating keyring",
			errPrefix, err)
	}

	// Import public key

	if err := executor.Exec("gpg", "--import", pubKeyPath); err != nil {
		return fmt.Errorf("%sError '%s' importing public-key %s",
			errPrefix, err, pubKeyPath)
	}

	clonedDir := gitSpec.ClonedDir
	revision := gitSpec.Revision
	if err := executor.ExecInDir(clonedDir, "git", "show-ref", "--quiet", "--tags"); err == nil {
		// the provided ref is a tag
		return executor.ExecInDir(clonedDir, "git", "verify-tag", "-v", revision)
	}

	if err := executor.ExecInDir(clonedDir, "git", "cat-file", "-e", revision); err == nil {
		// found an object with that hash

		return executor.ExecInDir(clonedDir, "git", "verify-commit", "-v", revision)
	}
	return fmt.Errorf("%sinvalid revision %s provided, provide either a COMMIT or TAG: %s", errPrefix, revision, err)
}
