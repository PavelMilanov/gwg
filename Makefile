GO := go
version := dev

.PHONY: build test deb release amd arm clean

build:
	cd src && $(GO) build -trimpath -ldflags="-X main.VERSION=$(version)" -o ../gwg .

test:
	cd src && $(GO) test ./...

deb:
	dpkg-buildpackage -us -uc -b

release: amd arm

amd:
	cd src && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -trimpath -ldflags="-s -w -X main.VERSION=$(version)" -o ../gwg .
	tar -cf gwg.linux_amd64.tar gwg
	$(RM) gwg

arm:
	cd src && CGO_ENABLED=0 GOOS=linux GOARCH=arm $(GO) build -trimpath -ldflags="-s -w -X main.VERSION=$(version)" -o ../gwg .
	tar -cf gwg.linux_arm.tar gwg
	$(RM) gwg

clean:
	$(RM) gwg gwg.linux_amd64.tar gwg.linux_arm.tar
