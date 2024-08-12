%define libpcap_version 1.10.1

Name:           libpcap
Epoch: 14
Version:        %{libpcap_version}
Release:        1%{?dist}.Ar.1.%{?eext_release:%{eext_release}}%{!?eext_release:eng}
Summary: A system-independent interface for user-level packet capture
Group: Development/Libraries
License: BSD with advertising
URL: http://www.tcpdump.org
BuildRequires: glibc-kernheaders >= 2.2.0 git bison flex libnl3-devel gcc
Requires: libnl3
BuildRoot: %{_tmppath}/%{name}-%{version}-%{release}-root-%(%{__id_u} -n)

Source0:        libpcap-%libpcap-{libpcap_version}.tar.gz



%description
Libpcap provides a portable framework for low-level network
monitoring.  Libpcap can provide network statistics collection,
security monitoring and network debugging.  Since almost every system
vendor provides a different interface for packet capture, the libpcap
authors created this system-independent API to ease in porting and to
alleviate the need for several system-dependent packet capture modules
in each application.

Install libpcap if you need to do low-level network traffic monitoring
on your network.


%package devel
Summary: Libraries and header files for the libpcap library
Group: Development/Libraries
Requires: %{name} = %{epoch}:%{version}-%{release}

%description devel
Libpcap provides a portable framework for low-level network
monitoring.  Libpcap can provide network statistics collection,
security monitoring and network debugging.  Since almost every system
vendor provides a different interface for packet capture, the libpcap
authors created this system-independent API to ease in porting and to
alleviate the need for several system-dependent packet capture modules
in each application.

This package provides the libraries, include files, and other 
resources needed for developing libpcap applications.

%prep
%autosetup -S git -n libpcap-%{libpcap_version}

#sparc needs -fPIC 
%ifarch %{sparc}
sed -i -e 's|-fpic|-fPIC|g' configure
%endif

find . -name '*.c' -o -name '*.h' | xargs chmod 644

%build
export CFLAGS="$RPM_OPT_FLAGS -fno-strict-aliasing"
# Explicitly specify each configure flag to avoid dynamically deciding what
# features to include.  If we want a feature, ensure that we have the
# supporting libraries listed as BuildRequires/Requires above.
# Read configure.ac to figure out whether it's "--enable/--disable"
# or "--with/--without".
%configure --enable-usb --disable-netmap --without-dpdk --disable-bluetooth \
    --disable-dbus --disable-rdma --without-dag --without-septel --without-snf \
    --without-turbocap
%{?a4_configure:exit 0}
make %{?_smp_mflags}

%install
rm -rf $RPM_BUILD_ROOT
make install DESTDIR=$RPM_BUILD_ROOT
rm -f $RPM_BUILD_ROOT%{_libdir}/libpcap.a

%clean
rm -rf $RPM_BUILD_ROOT

%post -p /sbin/ldconfig

%postun -p /sbin/ldconfig


%files
%defattr(-,root,root)
%doc LICENSE README.md CHANGES CREDITS
%{_libdir}/libpcap.so.*
%{_mandir}/man7/pcap*.7*

# Listing all include files individually for the benefit of pkgdeps
%files devel
%defattr(-,root,root)
%{_bindir}/pcap-config
%{_includedir}/pcap-bpf.h
%{_includedir}/pcap-namedb.h
%{_includedir}/pcap.h
%{_includedir}/pcap/bluetooth.h
%{_includedir}/pcap/bpf.h
%{_includedir}/pcap/can_socketcan.h
%{_includedir}/pcap/compiler-tests.h
%{_includedir}/pcap/dlt.h
%{_includedir}/pcap/funcattrs.h
%{_includedir}/pcap/ipnet.h
%{_includedir}/pcap/namedb.h
%{_includedir}/pcap/nflog.h
%{_includedir}/pcap/pcap-inttypes.h
%{_includedir}/pcap/pcap.h
%{_includedir}/pcap/sll.h
%{_includedir}/pcap/socket.h
%{_includedir}/pcap/usb.h
%{_includedir}/pcap/vlan.h
%{_libdir}/libpcap.so
%{_libdir}/pkgconfig/libpcap.pc
%{_mandir}/*


%changelog
* Fri Apr 22 2011 Miroslav Lichvar <mlichvar@redhat.com> 14:1.1.1-3
- ignore /sys/net/dev files on ENODEV (#693943)
- drop ppp patch
- compile with -fno-strict-aliasing

* Tue Feb 08 2011 Fedora Release Engineering <rel-eng@lists.fedoraproject.org> - 14:1.1.1-2
- Rebuilt for https://fedoraproject.org/wiki/Fedora_15_Mass_Rebuild

* Tue Apr 06 2010 Miroslav Lichvar <mlichvar@redhat.com> 14:1.1.1-1
- update to 1.1.1

* Wed Dec 16 2009 Miroslav Lichvar <mlichvar@redhat.com> 14:1.0.0-5.20091201git117cb5
- update to snapshot 20091201git117cb5

* Sat Oct 17 2009 Dennis Gilmore <dennis@ausil.us> 14:1.0.0-4.20090922gite154e2
- use -fPIC on sparc arches

* Wed Sep 23 2009 Miroslav Lichvar <mlichvar@redhat.com> 14:1.0.0-3.20090922gite154e2
- update to snapshot 20090922gite154e2
- drop old soname

* Fri Jul 24 2009 Fedora Release Engineering <rel-eng@lists.fedoraproject.org> - 14:1.0.0-2.20090716git6de2de
- Rebuilt for https://fedoraproject.org/wiki/Fedora_12_Mass_Rebuild

* Wed Jul 22 2009 Miroslav Lichvar <mlichvar@redhat.com> 14:1.0.0-1.20090716git6de2de
- update to 1.0.0, git snapshot 20090716git6de2de

* Wed Feb 25 2009 Fedora Release Engineering <rel-eng@lists.fedoraproject.org> - 14:0.9.8-4
- Rebuilt for https://fedoraproject.org/wiki/Fedora_11_Mass_Rebuild

* Fri Jun 27 2008 Miroslav Lichvar <mlichvar@redhat.com> 14:0.9.8-3
- use CFLAGS when linking (#445682)

* Tue Feb 19 2008 Fedora Release Engineering <rel-eng@fedoraproject.org> - 14:0.9.8-2
- Autorebuild for GCC 4.3

* Wed Oct 24 2007 Miroslav Lichvar <mlichvar@redhat.com> 14:0.9.8-1
- update to 0.9.8

* Wed Aug 22 2007 Miroslav Lichvar <mlichvar@redhat.com> 14:0.9.7-3
- update license tag

* Wed Jul 25 2007 Jesse Keating <jkeating@redhat.com> - 14:0.9.7-2
- Rebuild for RH #249435

* Tue Jul 24 2007 Miroslav Lichvar <mlichvar@redhat.com> 14:0.9.7-1
- update to 0.9.7

* Tue Jun 19 2007 Miroslav Lichvar <mlichvar@redhat.com> 14:0.9.6-1
- update to 0.9.6

* Tue Nov 28 2006 Miroslav Lichvar <mlichvar@redhat.com> 14:0.9.5-1
- split from tcpdump package (#193657)
- update to 0.9.5
- don't package static library
- maintain soname