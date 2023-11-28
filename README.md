# eos-external-tools

This repo hosts the source for tools for building external packages(RPMs) required for EOS out of Abuild.
There's a git repository corresponding to each such eos-external(eext) package. The tool to build these
package is hosted here and is called eext.

eext is a thin wrapper around Fedora mock, and is responsible to autogenerate a mock configuration.

This external package's git repository contains the modified spec file and
Arista specific patches and other sources.
It also includes some instructions for eext on how to build this package,
these are checked in as a manifest file named `eext.yaml` in the root of the repo.

The tool expects all third party repos to be cloned under `SrcDir` configuration,
it can be overridden with `EEXT_SRCDIR` environment variable.
If no --repo argument is specified SrcDir is ignored, and it is assumed that
that there is only one source repo and it is cloned in the current directory.

The built SRPMs and RPMs are made available in subdirectories of `DestDir` configuration,
it can be overridden with `EEXT_DESTDIR` environment variable.
The tool uses `WorkingDir` configuration to keep intermediate results,
it can be overridden with `EEXT_WORKINGDIR` environment variable.


Example usage:
```
eext create-srpm [-r <repo-name>]
eext mock [-r <repo-name>] -t <target-arch>
```

