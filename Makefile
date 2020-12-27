GITREV:=$(shell git rev-parse HEAD)

.PHONY: all clean build

all: clean

ball: clean build-win build-darwin build-linux

build-linux:
	env GOOS=linux GOARCH=amd64 go build -o store-linux-amd64 -ldflags \
 	"-X github.com/omecodes/stores/info.Revision=${GITREV}"\
	  github.com/omecodes/store
build-win:
	env GOOS=windows GOARCH=amd64 go build -o store-windows-amd64.exe -ldflags \
 	"-X github.com/omecodes/stores/info.Revision=${GITREV}"\
	  github.com/omecodes/store

build-darwin:
	env GOOS=darwin GOARCH=amd64 go build -o store-darwin-amd64 -ldflags \
 	"-X github.com/omecodes/stores/info.Revision=${GITREV}"\
 	  github.com/omecodes/store

clean:
	rm -f store-*