commit f50be4fe9b4f01028b89f8c5aafa3f19feef0545 (HEAD -> aajith-add-epel-relase)
Author: Arun Ajith S <aajith@arista.com>
Date:   Fri Sep 1 10:46:28 2023 +0000

    dnfconfig: Remove epel9 bundle
    
    The epel9 repo-bundle pointed to a moving dnf repository which is bad
    for reproducible eext builds.
    
    Instead of epel9, use a new repo-bundle epel9-subset which is a
    frozen/versioned subset of epel9 managed locally. It only has a few epel
    packages like quilt, hiredis etc required by existing eext third-party
    repos as on the date of this change.

commit 6d01f12a0bb8aef44bc48a0b7fba87ec94459916
Author: Arun Ajith S <aajith@arista.com>
Date:   Fri Sep 1 10:45:26 2023 +0000

    mock_cfg: Sort packages in chroot setup

commit 7fbc29a0483a981c7ddb1ae5032ad871a296041c
Author: Arun Ajith S <aajith@arista.com>
Date:   Fri Sep 1 10:40:30 2023 +0000

    base-image: Add epel-release into the base image
    
    This makes sure that /etc/pki/rpm-gpg/RPM-GPG-KEY-EPEL-9 is installed in
    the image to be used later by mock for gpgcheck.

commit 9b9101a508995e0d1bc073e719ff80ea2d4ccbc3
Author: Arun Ajith S <aajith@arista.com>
Date:   Fri Sep 1 02:51:38 2023 -0700

    bootstrap: Rename externaldeps to epel9

commit 87c2bd6e394ec9265563570f1fa7a9100f90b4c9 (origin/main, origin/aajith-new-features, origin/HEAD, main, aajith-new-features)
Author: Arun Ajith S <aajith@arista.com>
Date:   Sun Aug 20 02:15:01 2023 -0700

    cmd/testData: Fix yamllint issues

commit 507d1e249bd044306e197ae1a6009d9e568c3ad9
Author: Arun Ajith S <aajith@arista.com>
Date:   Sun Aug 20 01:57:55 2023 -0700

    Don't use container image based mock chroot
    
    mock-5.0 and above uses a podman container for bootstrap chroot.
    See: https://rpm-software-management.github.io/mock/Feature-container-for-bootstrap
    
    Avoid this to make sure we don't another container layer for the barney
    builds.

commit 96972b22a668c891fd8e080a51f7eb66e0e72c03
Author: Arun Ajith S <aajith@arista.com>
Date:   Sun Aug 20 08:38:31 2023 +0000

    Make eext base image build reproducible
    
    1. Use alma-vault as the base repo for already frozen release 9.1
    2. Exclude podman from the repos so that mock doesn't accidentally use
    podman based builds. Note that this is weak dependency for mock.
    3. Since epel keeps getting updates, don't use epel, instead use a local
    repo created manually which a subset of epel packages needed for eext.

commit 00285e29082a13a78ade9adc7380f5df36e23821
Author: Arun Ajith S <aajith@arista.com>
Date:   Wed Aug 16 03:44:06 2023 -0700

    Enable build for unmodified-srpms
    
    Add new package types `unmodified-srpm` and `standalone`.
    
    In unmodified-srpm, we expect no spec file or sources in the git
    repository. We'll just build the upstream srpm directly.
    
    For standalone srpms, there might be no upstream sources with all
    sources specified locally.
    
    Fixes: BUG838063

commit f93539ce72efef198032a6a96941a31747203441
Author: Arun Ajith S <aajith@arista.com>
Date:   Mon Aug 14 04:30:42 2023 -0700

    eext: Implement chained builds
    
    Add a new attribute `local-deps` in the build spec of the manifest.
    Add a new config DepsDir which is the directory containing local RPM
    dependencies, default it to /RPMS so that barney build output of
    dependencies can be pulled directly in to the build floor.
    
    When building packages with local-deps, all the dependencies in DepsDir
    are copied over to a sub-path in the working directory, then createrepo
    is run there to generate the yum metadata, and the auto-generated mock
    configuration has a new `local-deps` repo setup to point to this
    directory with the dependency RPMs and yum metadata with a `file://`
    URL.
    
    Enhanced mock_cfg_test to verify that the new repo is generated when the
    manifest specifies a chained build. Also added a separate test to make
    sure the deps are copied over and that the yum metadata is generated.
    
    Fixes: BUG837043

commit 99f73df50d0d3cecf4ab358169d37da3b55561af
Author: Arun Ajith S <aajith@arista.com>
