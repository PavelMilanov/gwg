GO := go
version := dev
DEB_OUTPUT_DIR ?= vagrant/share
DEB_PACKAGE = $(shell dpkg-parsechangelog -SSource)
DEB_ARCH = $(shell dpkg-architecture -qDEB_HOST_ARCH)
DEB_VERSION = $(if $(filter dev,$(version)),0~dev,$(version))
DEB_FILE = $(DEB_PACKAGE)_$(DEB_VERSION)_$(DEB_ARCH).deb

.PHONY: build test deb release amd arm clean

build:
	cd src && $(GO) build -trimpath -ldflags="-X github.com/PavelMilanov/go-wg-manager/cmd.Version=$(version)" -o ../gwg .

test:
	cd src && $(GO) test ./...

deb:
	@test "$(origin version)" = "command line" || { echo "usage: make deb version=dev" >&2; exit 2; }
	VERSION="$(version)" DEB_VERSION="$(DEB_VERSION)" dpkg-buildpackage -us -uc -b -d
	install -Dm644 "../$(DEB_FILE)" "$(DEB_OUTPUT_DIR)/$(DEB_FILE)"

release: amd arm

amd:
	cd src && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -trimpath -ldflags="-s -w -X github.com/PavelMilanov/go-wg-manager/cmd.Version=$(version)" -o ../gwg .
	tar -cf gwg.linux_amd64.tar gwg
	$(RM) gwg

arm:
	cd src && CGO_ENABLED=0 GOOS=linux GOARCH=arm $(GO) build -trimpath -ldflags="-s -w -X github.com/PavelMilanov/go-wg-manager/cmd.Version=$(version)" -o ../gwg .
	tar -cf gwg.linux_arm.tar gwg
	$(RM) gwg

clean:
	$(RM) gwg gwg.linux_amd64.tar gwg.linux_arm.tar
