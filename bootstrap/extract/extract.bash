#!/bin/bash

set -e
set -x

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

arch=$(arch)
bootstrap_file_repodir_path="eext-sources/bootstrap/CentOS-Stream"
bootstrap_filename_base="CentOS-Stream-Container-Base-9"
bootstrap_filename_version="20240715.0"
bootstrap_filename_extension="tar.xz"
bootstrap_filename="${bootstrap_filename_base}-${bootstrap_filename_version}.${arch}.${bootstrap_filename_extension}"

# URL of tarball with OS image
bootstrap_url="${DNF_HOST}/${bootstrap_file_repodir_path}/${bootstrap_filename}"

# Download the tarball into the mutable working dir
wget ${bootstrap_url}

# Validate downloaded tarball
grep "${bootstrap_filename}" "${COLLATERALS_DIR}/CHECKSUM" | sha256sum -wc

# Extract tarball and setup rootfs
# This is a nested tarball, the real rootfs is in layer.tar
# Extract the first level tarball inside the extract subdirectory
# within the working directory and and then extract the
# second level layer.tar directly to /dest
mkdir extract
tar --strip-components=1 -C ./extract -xf ./${bootstrap_filename}
tar -xf ./extract/layer.tar -C /dest

# Now modify the extracted file system to remove unwanted

# Note that we'll layer on our own curated yum repos and gpg keys into the bootstrap
# image instead of using the one from the bootstrap image
rm -rf /dest/etc/yum.repos.d
rm -rf /dest/usr/share/distribution-gpg-keys

