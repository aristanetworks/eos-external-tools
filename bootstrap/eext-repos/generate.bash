#!/bin/bash

set -e
set -x

generate_repo_files() {
   WORKING_DIR=`pwd`
   pushd ${COLLATERALS_DIR}
   ./generate-repo-file.bash ./eext-repos-build.repo.template ./repos-build.env "${WORKING_DIR}/eext-repos-build.repo"
   ./generate-repo-file.bash ./eext-repos-devel.repo.template ./repos-devel.env "${WORKING_DIR}/eext-repos-devel.repo"
   popd
   mkdir -p /dest/etc/yum.repos.d
   chmod 755 /dest/etc/yum.repos.d
   cp -a ./eext-repos-build.repo /dest/etc/yum.repos.d
}

generate_rpm() {
   WORKING_DIR=`pwd`
   mkdir -p rpmbuild
   pushd rpmbuild
   mkdir SOURCES SPECS

   cp "${COLLATERALS_DIR}"/eext-repos.spec SPECS/
   cp "${WORKING_DIR}"/*.repo "${COLLATERALS_DIR}"/*.pem SOURCES/
   rpmbuild --define "_topdir `pwd`" \
            --define "eext_alma_version ${DNF_DISTRO_REPO_VERSION}" \
            --define "eext_alma_release ${DNF_DISTRO_REPO_RELEASE}"  \
            --define "source_date_epoch_from_changelog 1" \
            --define "use_source_date_epoch_as_buildtime 1" \
            --define "clamp_mtime_to_source_date_epoch 1" \
            --define "_buildhost eext-buildhost" \
            --define "_build_name_fmt %%{NAME}.rpm" \
            -ba ./SPECS/eext-repos.spec

   if [ ! -f "./RPMS/eext-repos-build.rpm" ]; then
      echo "Error: './RPMS/ext-repos-build.rpm' not found after rpmbuild."
   fi

   if [ ! -f "./RPMS/eext-repos-devel.rpm" ]; then
      echo "Error: './RPMS/eext-repos-devel.rpm' not found after rpmbuild."
   fi

   mkdir -p /dest/RPMS
   chmod 755 /dest/RPMS
   cp -a ./RPMS/*.rpm /dest/RPMS/
   popd
}

setup_gpg_keys() {
   mkdir -p /dest/usr/share/eext-gpg-keys
   chmod 755  /dest/usr/share/eext-gpg-keys
   cp "${COLLATERALS_DIR}"/*.pem /dest/usr/share/eext-gpg-keys/
}

usage() {
   echo "Usage: $0 <collaterals_dir>"
   exit 1
}

if [ $# -ne 1 ]; then
   usage
fi

COLLATERALS_DIR=$1
if [ ! -d "$COLLATERALS_DIR" ]; then
   echo "Error: Collaterals directory '$COLLATERALS_DIR' not found."
   exit 1
fi

TEMPLATE_FILE=$1
set -a
source "$COLLATERALS_DIR/repos-common.env"
set +a
export ARCH=$(uname -m)

generate_repo_files
generate_rpm
setup_gpg_keys
