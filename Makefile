VERSION = 3.0.1

PACKAGES := $(shell go list -f {{.Dir}} ./...)
GOFILES  := $(addsuffix /*.go,$(PACKAGES))
GOFILES  := $(wildcard $(GOFILES))

.PHONY: clean

zip: release/ts_$(VERSION)_osx_x86_64.zip release/ts_$(VERSION)_windows_x86_64.zip release/ts_$(VERSION)_linux_x86_64.zip

binaries: binaries/osx_x86_64/ts binaries/windows_x86_64/ts.exe binaries/linux_x86_64/ts

clean: 
	rm -rf binaries/
	rm -rf release/

release/ts_$(VERSION)_osx_x86_64.zip: binaries/osx_x86_64/ts
	mkdir -p release
	cd binaries/osx_x86_64 && zip -r -D ../../release/ts_$(VERSION)_osx_x86_64.zip ts
	
binaries/osx_x86_64/ts: $(GOFILES)
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.config.version=$(VERSION)" -o binaries/osx_x86_64/ts ./cmd/ts

release/ts_$(VERSION)_windows_x86_64.zip: binaries/windows_x86_64/ts.exe
	mkdir -p release
	cd binaries/windows_x86_64 && zip -r -D ../../release/ts_$(VERSION)_windows_x86_64.zip ts.exe
	
binaries/windows_x86_64/ts.exe: $(GOFILES)
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.config.version=$(VERSION)" -o binaries/windows_x86_64/ts.exe ./cmd/ts

release/ts_$(VERSION)_linux_x86_64.zip: binaries/linux_x86_64/ts
	mkdir -p release
	cd binaries/linux_x86_64 && zip -r -D ../../release/ts_$(VERSION)_linux_x86_64.zip ts
	
binaries/linux_x86_64/ts: $(GOFILES)
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.config.version=$(VERSION)" -o binaries/linux_x86_64/ts ./cmd/ts