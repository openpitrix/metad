# Copyright 2018 The OpenPitrix Authors. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.

default:

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
	go get github.com/kisielk/godepgraph

clean:
