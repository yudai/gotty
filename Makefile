OUTPUT_DIR = ./builds
GIT_COMMIT = `git rev-parse HEAD | cut -c1-7`
VERSION = 2.0.0-alpha.3
BUILD_OPTIONS = -ldflags "-X main.Version=$(VERSION) -X main.CommitID=$(GIT_COMMIT)"

all: gotty

gotty: gotty.go server/*.go webtty/*.go backend/*.go Makefile asset
	go build ${BUILD_OPTIONS}

.PHONY: asset
asset: server/static/js/gotty-bundle.js server/static/index.html server/static/favicon.png server/static/css/index.css server/static/css/xterm.css server/static/css/xterm_customize.css

.PHONY: all
all: asset gotty

.PHONY: clean
clean:
	rm -rf js/node_modules
	rm -rf server/static
	rm -f gotty
	

server/static:
	mkdir server/static

server/static/index.html: server/static resources/index.html
	cp resources/index.html server/static/index.html

server/static/favicon.png: server/static resources/favicon.png
	cp resources/favicon.png server/static/favicon.png

server/static/js: server/static
	mkdir -p server/static/js


server/static/js/gotty-bundle.js: server/static/js js/dist/gotty-bundle.js
	cp js/dist/gotty-bundle.js server/static/js/gotty-bundle.js

server/static/css: server/static
	mkdir -p server/static/css

server/static/css/index.css: server/static/css resources/index.css
	cp resources/index.css server/static/css/index.css

server/static/css/xterm_customize.css: server/static/css resources/xterm_customize.css
	cp resources/xterm_customize.css server/static/css/xterm_customize.css

server/static/css/xterm.css: server/static/css js/node_modules/xterm/css/xterm.css
	cp js/node_modules/xterm/css/xterm.css server/static/css/xterm.css

js/node_modules/xterm/css/xterm.css:
	cd js && \
	npm install

js/dist/gotty-bundle.js: js/src/* js/node_modules/webpack
	cd js && \
	`npm bin`/webpack

js/node_modules/webpack: 
	cd js && \
	npm install

tools:
	go get github.com/tools/godep
	go get github.com/mitchellh/gox
	go get github.com/tcnksm/ghr

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
