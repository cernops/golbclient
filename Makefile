DIST ?= $(shell rpm --eval %{dist})
SPECFILE ?= golbclient.spec

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
#	go mod vendor
	tar cvf SOURCES/$PKG.tg  --exclude SOURCES --exclude .git --exclude .koji --exclude .gitlab-ci.yml --exclude go.mod --exclude go.sum --transform "s||$PKG/|" .
	gzip -c SOURCES/$PKG.tg > SOURCES/$PKG.tgz
	rm -rf SOURCES/$PKG.tg
	rpmbuild -bs --define 'dist $(DIST)' --define "_topdir $(PWD)/build" --define '_sourcedir $(PWD)' $(SPECFILE)

rpm: srpm
	echo "Creating the rpm"
	rpmbuild -bb --define 'dist $(DIST)' --define "_topdir $(PWD)/build" --define '_sourcedir $(PWD)' $(SPECFILE)
