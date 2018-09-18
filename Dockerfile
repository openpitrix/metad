# Copyright 2018 Yunify Inc. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.

FROM golang:1.10.1-alpine3.7 as builder

RUN apk add --no-cache upx git

WORKDIR /go/src/openpitrix.io/metad/
COPY . .

RUN mkdir -p /metad_bin
RUN go generate openpitrix.io/metad/pkg/version && \
CGO_ENABLED=0 GOOS=linux GOBIN=/metad_bin go install -ldflags '-w -s' -tags netgo openpitrix.io/metad

RUN find /metad_bin -type f -exec upx {} \;

FROM alpine:3.7
COPY --from=builder /metad_bin/* /usr/local/bin/

CMD ["sh"]
