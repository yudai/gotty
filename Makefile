gotty: app/resource.go main.go app/*.go
	go build

resource:  app/resource.go

app/resource.go: bindata/static/hterm.js bindata/static/gotty.js  bindata/static/index.html
	go-bindata -prefix bindata -pkg app -ignore=\\.gitkeep -o app/resource.go bindata/...
	gofmt -w app/resource.go

bindata:
	mkdir bindata

bindata/static: bindata
	mkdir bindata/static

bindata/static/hterm.js: bindata/static libapps/hterm/js/*.js
	cd libapps && \
	LIBDOT_SEARCH_PATH=`pwd` ./libdot/bin/concat.sh -i ./hterm/concat/hterm_all.concat -o ../bindata/static/hterm.js

bindata/static/gotty.js: bindata/static resources/gotty.js
	cp resources/gotty.js bindata/static/gotty.js

bindata/static/index.html: bindata/static resources/index.html
	cp resources/index.html bindata/static/index.html
