TODAY:=$(shell date -u +%Y-%m-%dT%H:%M:%S)
GITREV:=$(shell git rev-parse HEAD)
VERSION="0.0.1"
LICENSE="Apache-2.0 License"
OME=https://rootome.com
CLIENT=ome

.PHONY: all clean build main dev

all: clean

ball: clean build-win build-darwin build-linux

build-linux:
	env GOOS=linux GOARCH=amd64 go build -o omestore-linux-amd64 -ldflags\
 	"-X github.com/omecodes/omestores/store.Version=${VERSION}\
 	 -X github.com/omecodes/omestores/store.BuildDate=${TODAY}\
 	 -X github.com/omecodes/omestores/store.Ome=${OME}\
 	 -X github.com/omecodes/omestores/store.Revision=${GITREV}\
 	 -X github.com/omecodes/omestores/store.License=${LICENSE}" .

build-win:
	env GOOS=linux GOARCH=amd64 go build -o omestore-windows-amd64.exe -ldflags\
 	"-X github.com/omecodes/omestores/store.Version=${VERSION}\
 	 -X github.com/omecodes/omestores/store.BuildDate=${TODAY}\
 	 -X github.com/omecodes/omestores/store.Ome=${OME}\
 	 -X github.com/omecodes/omestores/store.Revision=${GITREV}\
 	 -X github.com/omecodes/omestores/store.License=${LICENSE}" .

build-darwin:
	env GOOS=darwin GOARCH=amd64 go build -o omestore-darwin-amd64 -ldflags\
 	"-X github.com/omecodes/omestores/store.Version=${VERSION}\
 	 -X github.com/omecodes/omestores/store.BuildDate=${TODAY}\
 	 -X github.com/omecodes/omestores/store.Ome=${OME}\
 	 -X github.com/omecodes/omestores/store.Revision=${GITREV}\
 	 -X github.com/omecodes/omestores/store.License=${LICENSE}" .

clean:
	rm -f cells omestore-*