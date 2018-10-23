Name: linstor-external-provisioner
Version: 0.7.8
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
* Tue Oct 18 2018 Hayley Swimelar <hayley@hayleylaptop.us.linbit>  0.7.8-1
-  Update controller to 5.2.0
-  Add flags to override client qps

* Wed Sep 05 2018 Roland Kammerer <roland.kammerer@linbit.com> 0.7.7-1
-  New upstream release

* Tue Jul 31 2018 Roland Kammerer <roland.kammerer@linbit.com> 0.7.5-1
-  New upstream release
