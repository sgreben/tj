VERSION = 5.0.2

PACKAGES := $(shell go list -f {{.Dir}} ./...)
GOFILES  := $(addsuffix /*.go,$(PACKAGES))
GOFILES  := $(wildcard $(GOFILES))

.PHONY: clean release

# go get -u github.com/github/hub
release: zip
	hub release delete $(VERSION) || true
	hub release create $(VERSION) -m "$(VERSION)" -a release/tj_$(VERSION)_osx_x86_64.zip -a release/tj_$(VERSION)_windows_x86_64.zip -a release/tj_$(VERSION)_linux_x86_64.zip

zip: release/tj_$(VERSION)_osx_x86_64.zip release/tj_$(VERSION)_windows_x86_64.zip release/tj_$(VERSION)_linux_x86_64.zip

binaries: binaries/osx_x86_64/tj binaries/windows_x86_64/tj.exe binaries/linux_x86_64/tj

clean: 
	rm -rf binaries/
	rm -rf release/

release/tj_$(VERSION)_osx_x86_64.zip: binaries/osx_x86_64/tj
	mkdir -p release
	cd binaries/osx_x86_64 && zip -r -D ../../release/tj_$(VERSION)_osx_x86_64.zip tj
	
binaries/osx_x86_64/tj: $(GOFILES)
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.config.version=$(VERSION)" -o binaries/osx_x86_64/tj ./cmd/tj

release/tj_$(VERSION)_windows_x86_64.zip: binaries/windows_x86_64/tj.exe
	mkdir -p release
	cd binaries/windows_x86_64 && zip -r -D ../../release/tj_$(VERSION)_windows_x86_64.zip tj.exe
	
binaries/windows_x86_64/tj.exe: $(GOFILES)
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.config.version=$(VERSION)" -o binaries/windows_x86_64/tj.exe ./cmd/tj

release/tj_$(VERSION)_linux_x86_64.zip: binaries/linux_x86_64/tj
	mkdir -p release
	cd binaries/linux_x86_64 && zip -r -D ../../release/tj_$(VERSION)_linux_x86_64.zip tj
	
binaries/linux_x86_64/tj: $(GOFILES)
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.config.version=$(VERSION)" -o binaries/linux_x86_64/tj ./cmd/tj