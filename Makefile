GIT_VER := $(shell git describe --tags)
DATE := $(shell date +%Y-%m-%dT%H:%M:%S%z)
export GO111MODULE := on

notico: *.go go.*
	go build .

clean:
	rm -f notico
