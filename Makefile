DIST ?= $(shell rpm --eval %{dist})
SPECFILE ?= golbclient.spec
RELEASE ?= $(shell grep 'Release = "' lbclient.go | cut -f 2 -d \")
VERSION ?= $(shell grep 'Version = "' lbclient.go | cut -f 2 -d \")

COPY ?= $(shell cat $(SPECFILE) | sed "s/\#REPLACE_BY_VERSION\#/$(VERSION)/" | sed "s/\#REPLACE_BY_RELEASE\#/$(RELEASE)/"  > $(SPECFILE).tmp )
 
PKG ?= $(shell $(COPY) rpm -q --specfile $(SPECFILE).tmp --queryformat "%{name}-%{version}\n" | head -n 1)

installgo:
	mkdir -p /go13
	yum -y install git gcc
	curl https://dl.google.com/go/go1.13.14.linux-amd64.tar.gz  | tar -zxC /go13
	ln -s /go13/go/bin/go /usr/bin/go13
	export GOPATH=/go13
	go13 get ./...

srpm: installgo
	echo "Creating the source rpm"
	mkdir -p SOURCES version
	go13 mod vendor
	tar zcf SOURCES/$(PKG).tgz  --exclude SOURCES --exclude .git --exclude .koji --exclude .gitlab-ci.yml --exclude go.mod --exclude go.sum --transform "s||$(PKG)/|" .
	rpmbuild -bs --define 'dist $(DIST)' --define "_topdir $(PWD)/build" --define '_sourcedir $(PWD)/SOURCES' $(SPECFILE).tmp
   
rpm: srpm
	echo "Creating the rpm"
	rpmbuild -bb --define 'dist $(DIST)' --define "_topdir $(PWD)/build" --define '_sourcedir $(PWD)/SOURCES' $(SPECFILE).tmp
