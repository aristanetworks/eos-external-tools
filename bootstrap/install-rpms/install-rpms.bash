#!/bin/bash

set -e
set -x

dnf_install() {
   dnf --assumeyes --installroot=/dest --noplugins \
       --config=/etc/dnf/dnf.conf \
       --setopt=cachedir=/var/cache/dnf \
       --setopt=reposdir=/etc/yum.repos.d \
       --setopt=varsdir=/etc/dnf \
       install "$@"
}

usage() {
    echo "Usage: $0 --common-rpms-file FILE --extra-rpms-file FILE"
    exit 1
}

# Parse command-line options
while [[ "$#" -gt 0 ]]; do
   case $1 in
      --common-rpms-file)
         common_rpms_file="$2"
         shift 2
         ;;
      --extra-rpms-file)
         extra_rpms_file="$2"
         shift 2
         ;;
      *)
         usage
         ;;
   esac
done

if [[ -z "$common_rpms_file" && -z "$extra_rpms_file" ]]; then
   echo "Error: At least one of the options must be specified."
   usage
fi

rpms=()
for file in "$common_rpms_file" "$extra_rpms_file"; do
   if [[ -n "$file" && ! -f "$file" ]]; then
       echo "Error: File '$file' does not exist."
       exit 1
   fi

   mapfile -t tmp_array < <(awk '!/^#/' "$file")
   rpms+=("${tmp_array[@]}")
done

if [[ ${#rpms[@]} -eq 0 ]]; then
   echo "Error: No RPMs specified"
   exit 1
fi

dnf_install "${rpms[@]}"

