# Copyright 2018 The OpenPitrix Authors. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.

default:

graph:
	godepgraph \
		-o openpitrix.io/metad \
		-p openpitrix.io/metad/vendor \
		openpitrix.io/metad \
	| \
		dot -Tpng > import-graph.png

tools:
	go get github.com/kisielk/godepgraph

clean:
