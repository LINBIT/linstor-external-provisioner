Name: linstor-external-provisioner
Version: 0.7.5
Release: 1%{?dist}
Summary: LINSTOR flexvolume plugin
License: GPLv2+
ExclusiveOS: linux
Source: %{name}-%{version}.tar.gz
Group: Applications/System

%description
External provisioner driver implementation for Linstor volumes


%prep
%setup -q

%build

%install
mkdir -p %{buildroot}/%{_sbindir}/
cp %{_builddir}/%{name}-%{version}/%{name} %{buildroot}/%{_sbindir}/

%files
%{_sbindir}/%{name}

%changelog
* Tue Jul 31 2018 Roland Kammerer <roland.kammerer@linbit.com> 0.7.5-1
-  New upstream release
