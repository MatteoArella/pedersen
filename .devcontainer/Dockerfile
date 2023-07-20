# syntax=docker/dockerfile:1

# Copyright (c) Pedersen authors.
# 
# Use of this source code is governed by an MIT-style
# license that can be found in the LICENSE file or at
# https://opensource.org/licenses/MIT.

ARG GO_VARIANT=1.17-alpine

FROM golang:${GO_VARIANT}

ARG GO_OPENSSL_VERSION_OVERRIDE=1.1.1

ENV GO_OPENSSL_VERSION_OVERRIDE=${GO_OPENSSL_VERSION_OVERRIDE}

RUN apk update && \
    apk del --no-cache openssl && \
    apk add --no-cache make build-base pkgconfig coreutils perl linux-headers

COPY scripts/openssl.sh /tmp/openssl.sh

RUN /tmp/openssl.sh --dev ${GO_OPENSSL_VERSION_OVERRIDE}
