gotty: app/resource.go main.go app/*.go
	go build

resource:  app/resource.go

app/resource.go: bindata/static/js/hterm.js bindata/static/js/gotty.js  bindata/static/index.html bindata/static/favicon.png
	go-bindata -prefix bindata -pkg app -ignore=\\.gitkeep -o app/resource.go bindata/...
	gofmt -w app/resource.go

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
