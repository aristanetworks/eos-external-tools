# Bootstrapping - RPM installation

This directory contains all the files required to install necessary RPMs when
building the image.

In the files with RPM list use a path specification for local rpm, otherwise
it'll be installed from one of the repos used for bootstrapping.

* `rpms-build`

    Specifies extra rpms to be installed in the build base-image.

* `rpms-common`

    Specifies common rpms to be installed in all base-images

* `rpms-devel`

    Specifies extra rpms to be installed in the devel base-image
