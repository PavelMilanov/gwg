.DEFAULT_GOAL := build

GO := go
component ?= gwu
DEB_OUTPUT_DIR ?= vagrant/share
DEB_ARCH = $(shell dpkg-architecture -qDEB_HOST_ARCH)
SOURCE_DATE_EPOCH ?= $(shell git log -1 --format=%ct 2>/dev/null || date +%s)
VERSION_PATTERN := ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$$

ifneq ($(words $(component)),1)
$(error component must be gwu or all)
endif
ifeq ($(component),gwu)
else ifeq ($(component),all)
else
$(error unknown component '$(component)'; use gwu or all)
endif

ifneq ($(origin version),undefined)
$(error version override is not supported; edit src/internal/gwucli/VERSION)
endif

.PHONY: build build-gwu test deb deb-gwu release amd arm clean check-version-gwu

build: build-gwu

check-version-gwu:
	@test "$$(wc -l < src/internal/gwucli/VERSION)" -eq 1 && \
		LC_ALL=C grep -Eq '$(VERSION_PATTERN)' src/internal/gwucli/VERSION || \
		{ echo "invalid version in src/internal/gwucli/VERSION" >&2; exit 1; }

build-gwu: check-version-gwu
	cd src && $(GO) build -trimpath -o ../gwu ./cmd/gwu

test:
	cd src && $(GO) test ./...

deb: deb-gwu

deb-gwu: check-version-gwu
	@set -eu; \
	mkdir -p .build/deb/gwu "$(DEB_OUTPUT_DIR)"; \
	stage=$$(mktemp -d "$(CURDIR)/.build/deb/gwu/build.XXXXXX"); \
	mkdir -p "$$stage/source/debian"; \
	cp -a Makefile readme.md LICENSE src docs "$$stage/source/"; \
	cp -a debian/common/. "$$stage/source/debian/"; \
	cp -a debian/gwu/. "$$stage/source/debian/"; \
	package_version=$$(cat "$$stage/source/src/internal/gwucli/VERSION"); \
	package_version=$${package_version#v}; \
	build_date=$$(LC_ALL=C date -d '@$(SOURCE_DATE_EPOCH)' -R); \
	printf 'gwu (%s) unstable; urgency=medium\n\n  * Build gwu from its VERSION file.\n\n -- Pavel Milanov <milanov71@icloud.com>  %s\n' \
		"$$package_version" "$$build_date" > "$$stage/source/debian/changelog"; \
	(cd "$$stage/source" && dpkg-buildpackage -us -uc -b -d); \
	install -m644 "$$stage/gwu_$${package_version}_$(DEB_ARCH).deb" "$(DEB_OUTPUT_DIR)/gwu_$${package_version}_$(DEB_ARCH).deb"

release: amd arm

amd: architecture = amd64
arm: architecture = arm

amd arm: check-version-gwu
	@set -eu; \
	binary_version=$$(cat src/internal/gwucli/VERSION); \
	output="$(CURDIR)/.build/release/gwu/$$binary_version/linux_$(architecture)"; \
	mkdir -p "$$output"; \
	(cd src && CGO_ENABLED=0 GOOS=linux GOARCH=$(architecture) $(GO) build -trimpath -ldflags='-s -w' -o "$$output/gwu" ./cmd/gwu); \
	tar -cf "gwu_$${binary_version}_linux_$(architecture).tar" -C "$$output" gwu

clean:
	$(RM) gwu
