# this make only for development
# make amd64 version=

version=

amd64:
	@GOOS=linux GOARCH=amd64 go install -ldflags="-X 'main.VERSION=${version}'"
	@cp /Users/pavel_milanov/go/bin/linux_amd64/go-wg-manager gwg
	@tar --totals -cvf gwg.linux_amd64.tar gwg
	@rm gwg

arm:
	@GOOS=linux GOARCH=arm go install -ldflags="-X 'main.VERSION=${version}'"
	@cp /Users/pavel_milanov/go/bin/linux_arm/go-wg-manager gwg
	@tar --totals -cvf gwg.linux_arm.tar gwg
	@rm gwg
