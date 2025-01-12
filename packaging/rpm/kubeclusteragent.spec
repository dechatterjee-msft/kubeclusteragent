Name:           kubeclusteragent
Version:
Release:        1%{?dist}
Summary:        kubeclusteragent service

License:        Apache 2.0
URL:            example.com
Source0:        kubeclusteragent-%{version}.tar.gz

Requires:       bash
%{?systemd_requires}
#BuildRequires:  systemd-rpm-macros

%description
Packages the kubeclusteragent binary as a systemd service

%prep
%autosetup

%install
rm -rf $RPM_BUILD_ROOT
mkdir -p $RPM_BUILD_ROOT/%{_bindir}
mkdir -p $RPM_BUILD_ROOT/%{_unitdir}
cp kubeclusteragent $RPM_BUILD_ROOT/%{_bindir}
cp kubeclusteragent.service $RPM_BUILD_ROOT/%{_unitdir}

%post
%systemd_post kubeclusteragent.service

%preun
%systemd_preun kubeclusteragent.service

%postun
%systemd_postun_with_restart kubeclusteragent.service

%files
%{_bindir}/kubeclusteragent
%{_unitdir}/kubeclusteragent.service

%changelog
- Initial RPM packaging
