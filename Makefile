OUTPUT_DIR = ./builds
GIT_COMMIT = `git rev-parse HEAD`

gotty: server/asset.go main.go server/*.go webtty/*.go backend/*.go
	godep go build

asset:  server/asset.go

server/asset.go: bindata/static/js/hterm.js bindata/static/js/gotty.js  bindata/static/index.html bindata/static/favicon.png
	go-bindata -prefix bindata -pkg server -ignore=\\.gitkeep -o server/asset.go bindata/...
	gofmt -w server/asset.go

bindata:
	mkdir bindata

bindata/static: bindata
	mkdir bindata/static

bindata/static/index.html: bindata/static resources/index.html
	cp resources/index.html bindata/static/index.html

bindata/static/favicon.png: bindata/static resources/favicon.png
	cp resources/favicon.png bindata/static/favicon.png

bindata/static/js: bindata/static
	mkdir -p bindata/static/js

bindata/static/js/hterm.js: bindata/static/js libapps/hterm/js/*.js
	cd libapps && \
	LIBDOT_SEARCH_PATH=`pwd` ./libdot/bin/concat.sh -i ./hterm/concat/hterm_all.concat -o ../bindata/static/js/hterm.js

bindata/static/js/gotty.js: bindata/static/js resources/gotty.js
	cp resources/gotty.js bindata/static/js/gotty.js

tools:
	go get github.com/tools/godep
	go get github.com/mitchellh/gox
	go get github.com/tcnksm/ghr
	go get github.com/jteeuwen/go-bindata/...

test:
	if [ `go fmt $(go list ./... | grep -v /vendor/) | wc -l` -gt 0 ]; then echo "go fmt error"; exit 1; fi

cross_compile:
	GOARM=5 gox -os="darwin linux freebsd netbsd openbsd" -arch="386 amd64 arm" -osarch="!darwin/arm" -output "${OUTPUT_DIR}/pkg/{{.OS}}_{{.Arch}}/{{.Dir}}"

targz:
	mkdir -p ${OUTPUT_DIR}/dist
	cd ${OUTPUT_DIR}/pkg/; for osarch in *; do (cd $$osarch; tar zcvf ../../dist/gotty_$$osarch.tar.gz ./*); done;

shasums:
	cd ${OUTPUT_DIR}/dist; sha256sum * > ./SHA256SUMS

release:
	ghr -c ${GIT_COMMIT} --delete --prerelease -u yudai -r gotty pre-release ${OUTPUT_DIR}/dist
