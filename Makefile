GITREV:=$(shell git rev-parse HEAD)

.PHONY: all clean build

all: clean

ball: clean build-win build-darwin build-linux

build-linux:
	env GOOS=linux GOARCH=amd64 go build -o omestore-linux-amd64 -ldflags \
 	"-X github.com/omecodes/omestores/info.Revision=${GITREV}"\
	  github.com/omecodes/omestore
build-win:
	env GOOS=windows GOARCH=amd64 go build -o omestore-windows-amd64.exe -ldflags \
 	"-X github.com/omecodes/omestores/info.Revision=${GITREV}"\
	  github.com/omecodes/omestore

build-darwin:
	env GOOS=darwin GOARCH=amd64 go build -o omestore-darwin-amd64 -ldflags \
 	"-X github.com/omecodes/omestores/info.Revision=${GITREV}"\
 	  github.com/omecodes/omestore

clean:
	rm -f omestore-*