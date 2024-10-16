#!/usr/bin/env bash

cat "$@" | xargs dnf --assumeyes --installroot=/dest --noplugins \
--config=/etc/dnf/dnf.conf \
--setopt=cachedir=/var/cache/dnf \
--setopt=reposdir=/etc/yum.repos.d \
--setopt=varsdir=/etc/dnf \
install
