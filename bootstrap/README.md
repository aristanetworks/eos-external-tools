# Bootstrapping a Redhat based distro image for the eext tool

This document explains how a base image for eext buids or development workflow is built. The base image should contain the utilties called by the eext static go binary.
The main dependencies are rpm, rpmbuild and mock.
We support two base images, one for running eext builds in Abuilds/bsy builds, and another for running eext builds in an user's interactive development environment. We call them base-image-build and base-image-devel.


## Bootstrap image
We need a RPM based rootfs to boostrap the eext base image which is today based of AlmaLinux.
We construct this rootfs image form a tarball published by some RPM based distro.
We need a tarball per target-arch, ie x86_64 amd aarch64. Note that i686 RPMs are built with the x86_64 eext base image itself.
This bootstrap tarball need not be AlmaLinux, it can be any RPM based distro. We're using a CentOS Stream tarball for this purpose, because they publish it to the mirror with checksums.

### Updating/Uploading the bootstrap tarball

1. Download a CentOS-Stream-Container-Base tarball with a timestamp in its name, like `20230704`, from:
   - [https://cloud.centos.org/centos/9-stream/x86_64/images/](https://cloud.centos.org/centos/9-stream/x86_64/images/)
   - [https://cloud.centos.org/centos/9-stream/aarch64/images/](https://cloud.centos.org/centos/9-stream/aarch64/images/)
2. Upload them to artifactory in the subpath `eext-sources/bootstrap/CentOS-Stream/`
```
  curl -H "Authorization: Bearer ${ARTIFACTORY_TOKEN}" -X PUT https://artifactory.infra.corp.arista.io/artifactory/eext-sources/bootstrap/CentOS-Stream/ -T <TARBALL_PATH>
```
3. Update the `extract/CHECKSUM` file in the local repo for the new entries from the `CHECKSUM` files:
   - [https://cloud.centos.org/centos/9-stream/x86_64/images/CHECKSUM](https://cloud.centos.org/centos/9-stream/x86_64/images/CHECKSUM)
   - [https://cloud.centos.org/centos/9-stream/aarch64/images/CHECKSUM](https://cloud.centos.org/centos/9-stream/aarch64/images/CHECKSUM)
4. Update the `bootstrap_filename_version` variable in `extract/extract.bash`.

## Base Image Build

### Repo configuration
To build an AlmaLinux base-image we run `dnf --installroot` inside the bootstrap image. We maintain our own curated dnf configuration on the `eext-repos` directory.
The configuration points to the second most recent dot release of the main AlmaLinux that eext is tracking. The idea is that this vaulted and frozen, giving us reproducible builds.
Note that we make them point to a local artifactory remote-repo/mirror of the upstream AlmaLinux repo. This makes sure we don't have unnecessary internet traffic, and we have a copy of the dependencies maintained in the remote repo's cache.

We maintain two such configuration file templates in our source for each variant of the base image:
1 `eext-repos/eext-repos-build.repo.template`
2.`eext-repos/eext-repos-devel.repo.template`

The build specific configuration just includes the basic repos needed for satisfying the the eext dependencies like `rpm` and `mock`.
The development configuration configures a superset of the build configuration, because the developer might need to install more tools like editors etc.
We try to track the major/minor versions of the distros similarly between the two. The `eext-repos/repos-common.env` file  holds the versions, actual repo names, URLs etc as environment variables.

We want the build specific repos to point to the `eext-alma-vault` artifactory remote repo, while we point the development repos to the the global `alma-vault` remote repository.
This makes sure that the dependencies in the developer's environment don't pollute `eext-alma-vault` cache.
The base-image specific configuration exists in the `eext-repos/repos-build.env` and `eext-repos/repos-devel.env` files.

The `eext-repos/generate.bash` script loads the `.env` files and builds the actual repo configuration files from the templates.

Note that we need repo configuration in two stages:
1. For the boostrap container for base-image generation
2. Repo configuration in the base image itself.

Note that the distinction in the repo configuration exists only for `2`. The bootstrap container is always configured with the build variant.
The developer just uses these repos to further install further packages in his bus or docker container with dnf.

### Base Image Contents
Once the bootstrap image and repo configuration is ready, we run the `install-rpms/install-rpms.bash` script to build the base image.
The set of rpms to be installed is specified in the text files: `install-rpms/rpms-common`, `install-rpms/rpms-build` and `install-rpms/rpms-devel`.
The `install-rpms` script takes two arguments, `--common-rpms-file rpms-common` and `--extra-rpms-file (rpms-build | rpms-devel)`.

