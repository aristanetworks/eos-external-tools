# eos-external-tools

This repo hosts the source for tools for building external packages(RPMs) required for EOS out of Abuild.
Each such package is called a lemur, and there's a git repository corresponding to each such package.

The main tool that is hosted here is called lemurbldr.
lemurbldr is a thin wrapper around Fedora mock, and is responsible to autogenerate a mock configuration.

Each lemur package has a git repository associated with it.
This repository contains the modified spec file and Arista specific patches and other sources.
It also includes some instructions for lemurbldr on how to install this package, these are checked in as
a manifest file named lemurbldr.yaml.

Example usage:
lemurbldr clone -r <repo-name> <repo-URL>
lemurbldr createSrpm -r <repo-name>
lemurbldr mock -r <repo-name> -t <target-arch>

Reach out to `eos-next@arista.com` if you have any questions.
