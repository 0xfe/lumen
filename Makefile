# Lumen Makefile

DIST = `pwd`/dist
SRC=lumen.go cli/*.go store/*.go main/main.go
BUILDSRC=main/main.go

default: all

clean:
	mkdir -p dist
	rm -rf $(DIST)/*

test: $(SRC)
	go test -v ./...

update: Gopkg.toml $(SRC)
	dep ensure -update

all: $(DIST)/lumen.linux.amd64 $(DIST)/lumen.linux.arm $(DIST)/lumen.linux.arm64 $(DIST)/lumen.macos $(DIST)/lumen.windows

.PHONY: clean all default

# Linux builds
$(DIST)/lumen.linux.amd64: $(SRC)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-extldflags "-static"' -o $@ $(BUILDSRC)

$(DIST)/lumen.linux.arm: $(SRC)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 go build -a -ldflags '-extldflags "-static"' -o $@ $(BUILDSRC)

$(DIST)/lumen.linux.arm64:: $(SRC)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -ldflags '-extldflags "-static"' -o $@ $(BUILDSRC)

# MacOS
$(DIST)/lumen.macos: $(SRC)
	CGO_ENABLED=0 GOOS=darwin go build -a -ldflags '-extldflags "-static"' -o $@ $(BUILDSRC)

# Windows
$(DIST)/lumen.windows: $(SRC)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -a -ldflags '-extldflags "-static"' -o $@ $(BUILDSRC)
