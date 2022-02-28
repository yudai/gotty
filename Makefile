OUTPUT_DIR = ./builds
GIT_COMMIT = `git rev-parse HEAD | cut -c1-7`
VERSION = 3.0.0-alpha.1
BUILD_OPTIONS = -ldflags "-X main.Version=$(VERSION) -X main.CommitID=$(GIT_COMMIT)"

gotty: main.go server/*.go webtty/*.go backend/*.go Makefile
	go build ${BUILD_OPTIONS}

server/asset.go: bindata/static/js/gotty-bundle.js bindata/static/js/gotty-bundle.js.map bindata/static/js/gotty-bundle.js.LICENSE.txt bindata/static/index.html bindata/static/favicon.png bindata/static/css/index.css bindata/static/css/xterm.css bindata/static/css/xterm_customize.css
	go-bindata -prefix bindata -pkg server -ignore=\\.gitkeep -o server/asset.go bindata/...
	gofmt -w server/asset.go

.PHONY: all
all: server/asset.go gotty

bindata:
	mkdir -p bindata

bindata/static: bindata
	mkdir -p bindata/static

bindata/static/index.html: bindata/static resources/index.html
	cp resources/index.html bindata/static/index.html

bindata/static/favicon.png: bindata/static resources/favicon.png
	cp resources/favicon.png bindata/static/favicon.png

bindata/static/js: bindata/static
	mkdir -p bindata/static/js


bindata/static/js/gotty-bundle.js: bindata/static/js js/dist/gotty-bundle.js
	cp js/dist/gotty-bundle.js bindata/static/js/gotty-bundle.js

bindata/static/js/gotty-bundle.js.LICENSE.txt: js/dist/gotty-bundle.js.LICENSE.txt
	cp $< $@
bindata/static/js/gotty-bundle.js.map: js/dist/gotty-bundle.js.map
	cp $< $@

bindata/static/css: bindata/static
	mkdir -p bindata/static/css

bindata/static/css/index.css: bindata/static/css resources/index.css
	cp resources/index.css bindata/static/css/index.css

bindata/static/css/xterm_customize.css: bindata/static/css resources/xterm_customize.css
	cp resources/xterm_customize.css bindata/static/css/xterm_customize.css

bindata/static/css/xterm.css: bindata/static/css js/node_modules/xterm/css/xterm.css
	cp js/node_modules/xterm/css/xterm.css bindata/static/css/xterm.css

js/node_modules/xterm/css/xterm.css:
	cd js && \
		npm install

js/dist/gotty-bundle.js: js/src/* js/node_modules/webpack
	cd js && \
		npx webpack

js/dist/gotty-bundle.js.LICENSE.txt: js/dist/gotty-bundle.js
js/dist/gotty-bundle.js.map: js/dist/gotty-bundle.js

js/node_modules/webpack:
	cd js && \
		npm install

tools:
	go install github.com/go-bindata/go-bindata/v3/go-bindata@latest # for static asset management
	# TODO convert to `go install`
	go get github.com/mitchellh/gox           # for crosscompiling
	go get github.com/tcnksm/ghr              # for making gihub releases

test:
	if [ `go fmt $(go list ./... | grep -v /vendor/) | wc -l` -gt 0 ]; then echo "go fmt error"; exit 1; fi

cross_compile:
	GOARM=5 gox -os="darwin linux freebsd netbsd openbsd" -arch="386 amd64 arm" -osarch="!darwin/arm" -output "${OUTPUT_DIR}/pkg/{{.OS}}_{{.Arch}}/{{.Dir}}"

targz:
	mkdir -p ${OUTPUT_DIR}/dist
	cd ${OUTPUT_DIR}/pkg/; for osarch in *; do (cd $$osarch; tar zcvf ../../dist/gotty_${VERSION}_$$osarch.tar.gz ./*); done;

shasums:
	cd ${OUTPUT_DIR}/dist; sha256sum * > ./SHA256SUMS

release:
	ghr -c ${GIT_COMMIT} --delete --prerelease -u yudai -r gotty pre-release ${OUTPUT_DIR}/dist

.PHONY: update_go_modules
update_go_modules:
	go get -u

.PHONY: clean
clean:
	rm -rf bindata/ js/node_modules/*
