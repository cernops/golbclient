%global provider      gitlab
%global provider_tld  cern.ch
%global project       lb-experts
%global provider_full %{provider}.%{provider_tld}/%{project}
%global repo          golbclient

%global import_path   %{provider_full}/%{repo}
%global gopath        %{_datadir}/gocode
%global debug_package %{nil}

Name: lbclient
Version: #REPLACE_BY_VERSION#
Release: #REPLACE_BY_RELEASE#%{?dist}

Summary: CERN DNS Load Balancer Client
License: ASL 2.0
URL: https://github.com/cernops/%{repo}

Source: %{name}-%{version}.tgz
BuildRequires: golang >= 1.13
BuildRequires: checkpolicy
%if 0%{?el6}%{?el7}
BuildRequires: policycoreutils-python
%else
BuildRequires: policycoreutils-python-utils
%endif
ExclusiveArch: x86_64
Requires: net-snmp

%description
%{summary}

This is a concurrent implementation of the CERN LB client.

The load balancing daemon dynamically handles the list of machines behind a
given DNS alias to allow scaling and improve availability.

The Domain Name System (DNS), the defacto standard for name resolution and
esential for the network, is an open standard based protocol which allows
the use of names instead of IP addresses on the network.

Load balancing is an advanced function that can be provided by DNS, to load
balance requests across several machines running the same service by using
the same DNS name.

The load balancing server requests each machine for its load status.
The SNMP daemon, gets the request and calls the locally installed metric
program, which delivers the load value in SNMP syntax to STDOUT. The SNMP
daemon then passes this back to the load balancing server.

The lowest loaded machine names are updated on the DNS servers via the
DynDNS mechanism.

%prep
%setup -n %{name}-%{version} -q
%if 0%{?rhel} >= 7
  cp -p config/cc7-lbclient.te config/lbclient.te
%endif
%if 0%{?rhel} == 6
  cp -p config/slc6-lbclient.te config/lbclient.te
%endif

checkmodule -M -m -o config/lbclient.mod config/lbclient.te
semodule_package -m config/lbclient.mod -o config/lbclient.pp


%build
mkdir -p src/%{provider_full}
ln -s ../../../ src/%{provider_full}/%{repo}
GOPATH=$(pwd):%{gopath} go build -o lbclient %{import_path}


%install
# main package binary
install -d -p %{buildroot}/usr/local/sbin/ %{buildroot}/usr/share/selinux/targeted/ %{buildroot}/usr/local/etc/ %{buildroot}/usr/sbin/
install -p -m0755 lbclient %{buildroot}/usr/sbin/lbclient
install -p config/lbclient.pp  %{buildroot}/usr/share/selinux/targeted/lbclient.pp
cd %{buildroot}/usr/local/sbin && ln -s ../../sbin/lbclient
echo "load constant -1" >  %{buildroot}/usr/local/etc/lbclient.conf

%preun
# remove all the files if this is the last package removed (not upgraded)
if [ "$1" == 0 ] ; then
  semodule -r lbclient
fi

%post
semodule -i /usr/share/selinux/targeted/lbclient.pp

%files
%doc LICENSE COPYING README.md
/usr/sbin/lbclient
/usr/local/sbin/lbclient
/usr/share/selinux/targeted/lbclient.pp
%config(noreplace) /usr/local/etc/lbclient.conf


%changelog
* Mon Nov 30 2020 Ignacio Reguero <ignacio.reguero@cern.ch> - 2.2.0-3
- Allow check for /eos/user and /eos/project
* Wed Oct 14 2020 Ignacio Reguero <ignacio.reguero@cern.ch> - 2.2.0-1
- New check for EOS
* Wed Aug 05 2020 Pablo Saiz <pablo.saiz@cern.ch>           - 2.1.3-2
- Move to rpmci
* Mon Jul 06 2020 Pablo Saiz <pablo.saiz@cern.ch>           - 2.1.3
- Removal of the selinux when the rpm is removed
* Tue Oct 08 2019 Pablo Saiz <pablo.saiz@cern.ch>           - 2.1.2
- Adding a default configuration to disable lbclient
* Thu Sep 05 2019 Pablo Saiz <pablo.saiz@cern.ch>           - 2.1.1
- Generic check daemons check tcp4 and tcp6
* Wed Sep 04 2019 Paulo Canilho <paulo.canilho@cern.ch>     - 2.1.0
- Fixed the [command] check returned errors to prevent failures during the [-t] puppet install
- Added a test to check the executable output format
- Implemented the [collectd_alarms] as a type of check
- Implemented a new logger; now using [logrus]
* Fri Mar 08 2019 Paulo Canilho <paulo.canilho@cern.ch>     - 2.0.8
- File logger
- Generilized the parameterized checks
* Tue Mar 05 2019 Ignacio reguero <Ignacio.reguero@cern.ch> - 2.0.7
- In collectd, we can specify the name of the anchor
* Thu Feb 28 2019 Pablo Saiz <Pablo.Saiz@cern.ch>           - 2.0.5
- Collectd support
* Thu Feb 21 2019 Pablo Saiz <Pablo.Saiz@cern.ch>           - 2.0.3
- Include the link in /usr/local/sbin/lbclient
* Wed Feb 13 2019 Pablo Saiz <pablo.saiz@cern.ch>           - 2.0.2
- Copying the old configuration file
* Tue Feb 05 2019 Paulo Canilho <Paulo.Canilho@cern.ch>     - 0.0.2
- Setting up the post install
* Fri Jun 15 2018 Pablo Saiz <Pablo.Saiz@cern.ch>           - 0.0.1
- First version of the rpm
