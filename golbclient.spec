%global provider	gitlab
%global provider_tld	cern.ch
%global project		lb-experts
%global provider_full %{provider}.%{provider_tld}/%{project}
%global repo		golbclient

%global import_path	%{provider_full}/%{repo}
%global gopath		%{_datadir}/gocode
%global debug_package	%{nil}

Name:		%{repo}
Version:	0.0
Release:	1

Summary:	CERN DNS Load Balancer Client
License:	ASL 2.0
URL:		https://%{import_path}
# Source:		https://%{import_path}/archive/%{commit}/%{repo}-%{shortcommit}.tar.gz
Source:		%{name}-%{version}.tgz
BuildRequires:	golang >= 1.5
ExclusiveArch:	x86_64 

%description
%{summary}

This is a concurrent implementation of the CERN LB client.

The load balancing daemon dynamically handles the list of machines behind a given DNS alias to allow scaling and improve availability.

The Domain Name System (DNS), the defacto standard for name resolution and esential for the network, is an open standard based protocol which allows the use of names instead of IP addresses on the network.
Load balancing is an advanced function that can be provided by DNS, to load balance requests across several machines running the same service by using the same DNS name.

The load balancing server requests each machine for its load status.
The SNMP daemon, gets the request and calls the locally installed metric program, which delivers the load value in SNMP syntax to STDOUT. The SNMP daemon then passes this back to the load balancing server.
The lowest loaded machine names are updated on the DNS servers via the DynDNS mechanism.


%prep
%setup -n %{name}-%{version} -q
%if 0%{?rhel} >= 7
  cp -p config/cc7-lbclient.te config/lbclient.te
%endif
%if 0%{?rhel} == 6
  cp -p config/slc6-lbclient.te config/lbclient.te
%endif

checkmodule -M -m -o config/lbclient.mod config/lbclient.te
semodule_package -m config/lbclient.mod -o /usr/share/selinux/targeted/lbclient.pp
semodule -i /usr/share/selinux/targeted/lbclient.pp


%build
mkdir -p src/%{provider_full}
ln -s ../../../ src/%{provider_full}/%{repo}
GOPATH=$(pwd):%{gopath} go build -o lbclient %{import_path}



%install
# main package binary
install -d -p %{buildroot}%{_bindir}
install -p -m0755 lbclient %{buildroot}%{_bindir}/lbclient


%check
GOPATH=$(pwd)/:%{gopath} go test %{provider_full}/%{repo}


%files
%doc LICENSE COPYING README.md
/usr/bin/lbclient


%changelog
* Fri Jun 15 2018 Pablo Saiz <Pablo.Saiz@cern.ch>           - 0.0.1
- First version of the rpm
