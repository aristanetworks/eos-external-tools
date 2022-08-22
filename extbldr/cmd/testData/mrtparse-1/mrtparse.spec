# Arista patches begin {
Patch1000: arista-remove-default-sigpipe.patch
# } end Arista patches

Summary:        Tool for parsing routing information dump files in MRT format
Name:           mrtparse
Version:        2.0.1
Release:        %{?release:%{release}}%{!?release:eng}

License:        Apache 
Source0:        mrtparse-2.0.1.tar.gz
Url:            https://github.com/YoshiyukiYamauchi/mrtparse
BuildArch:      noarch

%description
mrtprse is a module to read and analyze the MRT format data. The MRT format data 
can be used to export routing protocol messages, state changes, and routing 
information base contents, and is standardized in RFC6396. Programs like Quagga 
/ Zebra, BIRD, OpenBGPD and PyRT can dump the MRT fotmat data.

%package -n python2-mrtparse
Summary:        Tool for parsing routing information dump files in MRT format
BuildRequires:  python2-devel
BuildRequires:  python2-setuptools
%{?python_provide:%python_provide python2-mrtparse}
Provides:       mrtparse = %{version}-%{release}
Obsoletes:      mrtparse < %{version}-%{release}

%description -n python2-mrtparse
mrtprse is a module to read and analyze the MRT format data. The MRT format data 
can be used to export routing protocol messages, state changes, and routing 
information base contents, and is standardized in RFC6396. Programs like Quagga 
/ Zebra, BIRD, OpenBGPD and PyRT can dump the MRT fotmat data.

%package -n python3-mrtparse
Summary:        Tool for parsing routing information dump files in MRT format
BuildRequires:  python3-devel
BuildRequires:  python3-setuptools
%{?python_provide:%python_provide python3-mrtparse}

%description -n python3-mrtparse
mrtprse is a module to read and analyze the MRT format data. The MRT format data 
can be used to export routing protocol messages, state changes, and routing 
information base contents, and is standardized in RFC6396. Programs like Quagga 
/ Zebra, BIRD, OpenBGPD and PyRT can dump the MRT fotmat data.

%prep
%setup -n mrtparse-2.0.1

# Arista patches begin {
%patch1000 -p1
# } end Arista patches

%build
%py2_build
%py3_build

%install
%py2_install
%py3_install

%files -n python2-mrtparse
%license LICENSE
%doc README.rst
%{python2_sitelib}/mrtparse-%{version}-py%{python2_version}.egg-info
%{python2_sitelib}/mrtparse/__init__.py*
%{python2_sitelib}/mrtparse/base.py*
%{python2_sitelib}/mrtparse/params.py*

%files -n python3-mrtparse
%license LICENSE
%doc README.rst
%{python3_sitelib}/mrtparse-%{version}-py%{python3_version}.egg-info
%{python3_sitelib}/mrtparse/__init__.py*
%{python3_sitelib}/mrtparse/base.py*
%{python3_sitelib}/mrtparse/params.py*
%{python3_sitelib}/mrtparse/__pycache__/*cpython*

