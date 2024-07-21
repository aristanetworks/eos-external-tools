Name:           eext-release
Version:        %{eext_alma_version}
Release:        alma_%{eext_alma_release}.eext_1
Summary:        Empty AlmaLinux release for eext
License:        GPLv2

Source0: eext-repos-build.repo
Source1: eext-repos-devel.repo
Source2: alma9-b86b3716-gpg-pubkey.pem
Source3: alma9-b86b3716-gpg-pubkey.pem

%description
Dummy package to define srpm

%package -n eext-gpg-keys
Summary:        gpgkeys for curated repos
Requires:       almalinux-release = %{eext_alma_version}-%{eext_alma_release}

%description -n eext-gpg-keys
gpgkeys for curated repos

%package -n eext-repos-build
Summary:        Subset of vaulted almalinux repos to be used by eext build
Requires:       almalinux-release = %{eext_alma_version}-%{eext_alma_release}
# Remove any almalinux-repos
Obsoletes:      almalinux-repos = %{eext_alma_version}-%{eext_alma_release}
Provides:       almalinux-repos = %{eext_alma_version}-%{eext_alma_release}
# Don't allow epel repos to be configured
Conflicts:      epel-release
Requires:       eext-gpg-keys = %{version}-%{release}

%description -n eext-repos-build
Subset of vaulted almalinux repos to be used by eext build.
The vaulted penultimate dot release is used to ensure a frozen dnf repo for build reproducibility.
The "eext-alma-vault" local mirror is used. 

%package -n eext-repos-devel
Summary:        Vaulted almalinux repos and other disabled repos to be used for eext dev workflow. 
Requires:       almalinux-release = %{eext_alma_version}-%{eext_alma_release}
# Remove any almalinux-repos
Obsoletes:      almalinux-repos = %{eext_alma_version}-%{eext_alma_release}
Obsoletes:      eext-repos-build = %{version}-%{release}
Provides:       almalinux-repos = %{eext_alma_version}-%{eext_alma_release}
# Don't allow epel repos to be configured
Conflicts:      epel-release
Requires:       eext-gpg-keys = %{version}-%{release}

%description -n eext-repos-devel
Vaulted almalinux repos and other disabled repos to be used for eext dev workflow.
The "alma-vault" local mirror/cache is used to avoid polluting the "eext-alma-vault"
local mirror's cache with RPMs pulled in by developers for their environment.
This is because the local mirror's cache can be snapshoted for every release to hold
the dependency set.

%install
# create /etc/yum.repos.d
install -d -m 0755 %{buildroot}%{_sysconfdir}/yum.repos.d
install -p -m 0644 %{SOURCE0} %{buildroot}%{_sysconfdir}/yum.repos.d/
install -p -m 0644 %{SOURCE1} %{buildroot}%{_sysconfdir}/yum.repos.d/
mkdir -p %{buildroot}%{_datadir}/eext-gpg-keys
cp -a %{SOURCE2} %{buildroot}%{_datadir}/eext-gpg-keys
cp -a %{SOURCE3} %{buildroot}%{_datadir}/eext-gpg-keys

%files -n eext-gpg-keys
%{_datadir}/eext-gpg-keys

%files -n eext-repos-build
%{_sysconfdir}/yum.repos.d/eext-repos-build.repo

%files -n eext-repos-devel
%{_sysconfdir}/yum.repos.d/eext-repos-devel.repo

%changelog
* Sun Jul 21 2024 Arun Ajith S <aajith@arista.com> - 9.3-alma_1.el9.eext_1
- Creating spec file when we're based off almalinux-release-9.3-1
