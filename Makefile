# Copyright 2018 The OpenPitrix Authors. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.

# go get -u github.com/kardianos/govendor

VGO_PROXY:=http_proxy=socks5://127.0.0.1:2080 https_proxy=socks5://127.0.0.1:2080

default:
	./metad \
		-backend=etcdv3 \
		-nodes=http://127.0.0.1:2379 \
		-log_level=debug \
		-listen=:8080 \
		-xff

init-etcd:
	ETCDCTL_API=3 etcdctl get --prefix ""

	ETCDCTL_API=3 etcdctl put abc abc-value
	ETCDCTL_API=3 etcdctl put abc/aaa abc/aaa-value

init-metad:
	curl -X PUT -H "Content-Type: application/json" http://127.0.0.1:9611/v1/data \
		--data-binary "@./testdata/simple.json"
	curl -H "Content-Type: application/json" -X PUT http://127.0.0.1:9611/v1/mapping \
		--data-binary "@./testdata/mapping.json"


metad-get:
	curl -H "Accept: application/json" http://127.0.0.1:9611/v1/mapping
	curl -H "Accept: application/json" -H "X-Forwarded-For: 192.168.1.1" http://127.0.0.1:8080/self/node

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

build-image-%: ## build docker image
	@if [ "$*" = "latest" ];then \
	docker build -t openpitrix/metad:latest .; \
	elif [ "`echo "$*" | grep -E "^v[0-9]+\.[0-9]+\.[0-9]+"`" != "" ];then \
	docker build -t openpitrix/metad:$* .; \
	fi

push-image-%: ## push docker image
	@if [ "$*" = "latest" ];then \
	docker push openpitrix/metad:latest; \
	elif [ "`echo "$*" | grep -E "^v[0-9]+\.[0-9]+\.[0-9]+"`" != "" ];then \
	docker push openpitrix/metad:$*; \
	fi

clean:
