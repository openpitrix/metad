# Copyright 2018 The OpenPitrix Authors. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.

# go get -u github.com/kardianos/govendor

VGO_PROXY:=http_proxy=socks5://127.0.0.1:2080 https_proxy=socks5://127.0.0.1:2080

default:

init-vendor:
	govendor init
	govendor add +external

update-vendor:
	govendor update +external
	govendor list

vgo_build:
	$(VGO_PROXY) vgo build

vgo:
	vgo fmt ./...
	vgo vet ./...
	vgo test ./...

vgo_vendor:
	mv vendor _vendor_backup
	vgo mod -vendor

graph:
	godepgraph \
		-o openpitrix.io/metad \
		-p openpitrix.io/metad/vendor \
		openpitrix.io/metad \
	| \
		dot -Tpng > import-graph.png

tools:
	go get golang.org/x/vgo
	go get github.com/kardianos/govendor
	go get github.com/kisielk/godepgraph

clean:
