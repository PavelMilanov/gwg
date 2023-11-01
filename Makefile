# this make only for development
amd64:
	@GOOS=linix GOARCH=amd go install
	@cp /Users/pavel_milanov/go/bin/linux_amd64/go-wg-manager gwg
	@tar --totals -cvf gwg.linux_amd64.tar gwg
	@rm gwg

arm:
	@GOOS=linix GOARCH=arm go install
	@cp /Users/pavel_milanov/go/bin/linux_arm/go-wg-manager gwg
	@tar --totals -cvf gwg.linux_arm.tar gwg
	@rm gwg
