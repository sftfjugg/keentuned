%define debug_package %{nil}
%define anolis_release 2

Name:           keentuned
Version:        2.0.1
Release:        %{?anolis_release}%{?dist}
Url:            https://gitee.com/anolis/keentuned
Summary:        KeenTune tuning tools
Vendor:         Alibaba
License:        MulanPSLv2
Source:         https://gitee.com/anolis/keentuned/repository/archive/%{name}-%{version}.tar.gz

Autoreq:        0
BuildRequires:  go >= 1.13
BuildRequires:	systemd

Requires(post): systemd
Requires(preun): systemd
Requires(postun): systemd

%description
KeenTune tuning tools rpm package

%prep
%setup -n %{name}-%{version}

%build
%make_build

%install
%make_install

%clean
[ "$RPM_BUILD_ROOT" != "/" ] && rm -rf "$RPM_BUILD_ROOT"
rm -rf $RPM_BUILD_DIR/%{name}-%{version}

%post
%systemd_post keentuned.service

%preun
%systemd_preun keentuned.service

%postun
%systemd_postun_with_restart keentuned.service

%files
%defattr(0644,root,root, 0755)
%license LICENSE
%doc README.md docs/*
%attr(0755, root, root) %{_bindir}/keentune
%attr(0755, root, root) %{_bindir}/keentuned
%attr(0666, root, root) %{_sysconfdir}/keentune/conf/keentuned.conf
%dir %{_sysconfdir}/keentune
%dir %{_sysconfdir}/keentune/conf
%{_sysconfdir}/keentune
%{_prefix}/lib/systemd/system/keentuned.service
%{_mandir}/man8/keentune.8*
%{_mandir}/man8/keentuned.8*
%{_mandir}/man5/keentuned.conf.5*
%{_mandir}/man7/keentune-benchmark.7*
%{_mandir}/man7/keentune-profile.7*
%{_mandir}/man7/keentune-detect.7*
%{_prefix}/share/bash-completion/completions/keentune.bash

%changelog
* Tue Jan 17 2023 Wenchao Liu <wb-lwc921938@alibaba-inc.com> - 2.0.1-2
- add: Update profile set for UI
- fix: Reduce the delay of default setting
- fix: Sync update keentune version in docs
- fix: other known important bug fixes

* Thu Dec 15 2022 Wenchao Liu <wb-lwc921938@alibaba-inc.com> - 2.0.0-0
- Support experience tuning of Tuned

* Thu Sep 22 2022 Wenchao Liu <wb-lwc921938@alibaba-inc.com> - 1.4.0-0
- add: Support UI Intelligent tuning and Sensitive parameter identification
- add: Support show recommendation log for expert one click tuning
- add: Support UI configure KeenTune groups' network environment
- add: Support UI perform param rollback
- add: Support UI perform param dump
- add: Support KeenTune Recovery of all environments
- add: Remove field '--config' from Intelligent tuning and Sensitive parameter identification
- fix: Remove daemon self start
- fix: Unify response code for cli
- fix: Modify daemon start show log
- other bugs fix

* Thu Jul 21 2022 Runzhe Wang <runzhe.wrz@alibaba-inc.com> - 1.3.0-1
- fix: missing of man dir

* Tue Jul 12 2022 Wenchao Liu <wb-lwc921938@alibaba-inc.com> - 1.3.0-0
- Add: Support UI expert one click tuning

* Mon Jun 20 2022 Wenchao Liu <wb-lwc921938@alibaba-inc.com> - 1.2.1-2
- fix: Clear task residue on startup

* Mon Jun 20 2022 Runzhe Wang <runzhe.wrz@alibaba-inc.com> - 1.2.1-1
- update docs

* Thu May 05 2022 happy_orange <songnannan@linux.alibaba.com> - 1.2.0-2
- add makefile
- update spec file

* Mon Apr 04 2022 Runzhe Wang <runzhe.wrz@alibaba-inc.com> - 1.2.0
- Add capabilities of target-group and bench-group
- Fix some issues
- Add 'keentune version' command

* Thu Mar 03 2022 Runzhe Wang <runzhe.wrz@alibaba-inc.com> - 1.1.0
- remove parameter fs.nr_open
- Add support for GP (in iTuned) in sensitizing algorithms
- Add support for lasso in sensitizing algorithms
- refactor tornado module: replace await by threadpool
- lazy load domain in keentune-target
- fix other bugs
- Add baseline reading before init brain
- Supporting multiple param json for tuning
- Fix rollback failure
- Clean empty dir when uninstall
- Verify input arguments of command 'param tune'
- Supporting of multiple target tuning
- Fix bug which cause keentune hanging after command 'param stop'
- Add verification of conflicting commands such as 'param dump', 'param delete' when a tuning job is runing.
- Remove version limitation of tornado
- Refactor sysctl domain to improve stability of parameter setting
- Fix some user experience issues

* Wed Dec 15 2021 Runzhe Wang <runzhe.wrz@alibaba-inc.com> - 1.0.0
- add tpce tpch benchmark files
- fix bug: can not running in alinux2 and centos7
- change modify codeup address to gitee
- manage keentuned with systemctl
- fix: show brain error in the keentuned log
- fix: profile set supports absolute and relative paths
- fix: show exact job abort log after the stop command
- add nginx_conf parameter config file
- use '%license' macro
- update license to MulanPSLv2
- Init Keentuned.
