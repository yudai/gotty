gotty: resource.go main.go app/*.go
	go build

resource.go: bindata/hterm.js bindata/index.html bindata/gotty.js
	go-bindata -pkg app -ignore=\\.gitkeep -o app/resource.go bindata/

bindata:
	mkdir bindata
bindata/hterm.js: bindata libapps/hterm/js/*.js
	cd libapps && \
	LIBDOT_SEARCH_PATH=`pwd` ./libdot/bin/concat.sh -i ./hterm/concat/hterm_all.concat -o ../bindata/hterm.js

bindata/index.html: bindata resources/index.html
	cp resources/index.html bindata/index.html

bindata/gotty.js: bindata resources/gotty.js
	cp resources/gotty.js bindata/gotty.js
