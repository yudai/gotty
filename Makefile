OUTPUT_DIR = ./builds
GIT_COMMIT = `git rev-parse HEAD | cut -c1-7`
VERSION = 1.1.0
BUILD_OPTIONS = -ldflags "-X main.Version=$(VERSION) -X main.CommitID=$(GIT_COMMIT)"

gotty: main.go server/*.go webtty/*.go backend/*.go Makefile asset
	go build ${BUILD_OPTIONS}

docker: 
	docker build . -t gotty-bash:$(VERSION)

.PHONY: asset
asset: bindata/static/js/gotty.js bindata/static/index.html bindata/static/favicon.png bindata/static/css/index.css bindata/static/css/xterm.css bindata/static/css/xterm_customize.css bindata/static/manifest.json bindata/static/icon_192.png server/asset.go

server/asset.go:
	go-bindata -prefix bindata -pkg server -ignore=\\.gitkeep -o server/asset.go bindata/...
	gofmt -w server/asset.go

.PHONY: all
all: asset gotty docker

bindata:
	mkdir bindata

bindata/static: bindata
	mkdir bindata/static

bindata/static/index.html: bindata/static resources/index.html
	cp resources/index.html bindata/static/index.html

bindata/static/manifest.json: bindata/static resources/manifest.json
	cp resources/manifest.json bindata/static/manifest.json

bindata/static/favicon.png: bindata/static resources/favicon.png
	cp resources/favicon.png bindata/static/favicon.png

bindata/static/icon_192.png: bindata/static resources/icon_192.png
	cp resources/icon_192.png bindata/static/icon_192.png

bindata/static/js: bindata/static
	mkdir -p bindata/static/js

bindata/static/css: bindata/static
	mkdir -p bindata/static/css

bindata/static/css/index.css: bindata/static/css resources/index.css
	cp resources/index.css bindata/static/css/index.css

bindata/static/css/xterm_customize.css: bindata/static/css resources/xterm_customize.css
	cp resources/xterm_customize.css bindata/static/css/xterm_customize.css

bindata/static/css/xterm.css: bindata/static/css js/node_modules/xterm/dist/xterm.css
	cp js/node_modules/xterm/dist/xterm.css bindata/static/css/xterm.css

js/node_modules/xterm/dist/xterm.css:
	cd js && \
	npm install

bindata/static/js/gotty.js: js/src/* js/node_modules/webpack
	cd js && \
	npx webpack

js/node_modules/webpack:
	cd js && \
	npm install

README.md: README.md.in
	git log --pretty=format:' * %aN' | \
		grep -v 'S.*ren L. Hansen' | \
		grep -v 'Iwasaki Yudai' | \
		sort -u > contributors.txt.tmp
	./gotty --help | sed '1,/GLOBAL OPTIONS/ d' > options.txt.tmp
	sed -f README.md.sed < $< > $@
	rm contributors.txt.tmp options.txt.tmp

tools:
	go get github.com/tools/godep
	go get github.com/mitchellh/gox
	go get github.com/tcnksm/ghr
	go get github.com/jteeuwen/go-bindata/...

test:
	if [ `go fmt $(go list ./... | grep -v /vendor/) | wc -l` -gt 0 ]; then echo "go fmt error"; exit 1; fi

cross_compile:
	GOARM=5 gox -os="darwin linux freebsd netbsd openbsd solaris" -arch="386 amd64 arm" -osarch="!darwin/386" -osarch="!darwin/arm" -output "${OUTPUT_DIR}/pkg/{{.OS}}_{{.Arch}}/{{.Dir}}"

targz:
	mkdir -p ${OUTPUT_DIR}/dist
	cd ${OUTPUT_DIR}/pkg/; for osarch in *; do (cd $$osarch; tar zcvf ../../dist/gotty_${VERSION}_$$osarch.tar.gz ./*); done;

shasums:
	cd ${OUTPUT_DIR}/dist; sha256sum * > ./SHA256SUMS

release-artifacts: asset gotty cross_compile targz shasums

release:
	ghr -draft ${VERSION} ${OUTPUT_DIR}/dist # -c ${GIT_COMMIT} --delete --prerelease -u sorenisanerd -r gotty ${VERSION}

clean:
	rm -fr gotty builds bindata server/asset.go js/dist
