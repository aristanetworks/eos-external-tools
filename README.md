# eos-external-tools

This repo hosts the source for tools for building external packages required for EOS out of Abuild.
Each package that this repo builds has it's own repository with a name starting with `eos-external-pkg-`.
These packages are built using github actions with runners running on the Arista `infra` k8s cluster and build artifatcts are RPMs that are published to Arista's artifactory instance.

Reach out to `eos-next@arista.com` if you have any questions.
