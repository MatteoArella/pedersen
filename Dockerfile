# syntax=docker/dockerfile:1

# Copyright (c) Pedersen authors.
#
# Use of this source code is governed by an MIT-style
# license that can be found in the LICENSE file or at
# https://opensource.org/licenses/MIT.

ARG GO_VERSION=1.20.5
ARG XX_VERSION=1.2.1
ARG GOLANGCI_LINT_VERSION=v1.53.2
ARG ADDLICENSE_VERSION=v1.0.0
ARG NODE_VERSION=20
ARG LICENSE_FILES=".*\(Dockerfile\|\.go\|\.hcl\)"

# xx is a helper for cross-compilation
FROM --platform=${BUILDPLATFORM} tonistiigi/xx:${XX_VERSION} AS xx

# osxcross contains the MacOSX cross toolchain for xx
FROM tonistiigi/xx:sdk-extras AS osxcross

FROM golangci/golangci-lint:${GOLANGCI_LINT_VERSION}-alpine AS golangci-lint
FROM ghcr.io/google/addlicense:${ADDLICENSE_VERSION} AS addlicense

FROM --platform=${BUILDPLATFORM} golang:${GO_VERSION}-alpine AS base
COPY --from=xx / /
RUN apk add --no-cache bash perl make pkgconfig clang lld llvm git file
WORKDIR /src
ENV CGO_ENABLED=1

FROM --platform=${BUILDPLATFORM} base AS base-linux
ARG TARGETPLATFORM
RUN xx-apk add gcc musl-dev linux-headers

FROM --platform=${BUILDPLATFORM} base-linux AS base-darwin

FROM --platform=${BUILDPLATFORM} base AS base-windows
RUN apk add --no-cache mingw-w64-gcc
ARG TARGETPLATFORM
RUN mv /usr/bin/$(xx-info)-windres /usr/bin/$(xx-info)-windres.orig
RUN xx-clang --setup-target-triple && \
    ln -s /usr/bin/$(xx-info)-windres /usr/bin/windres

FROM base-${TARGETOS} AS build-base
COPY go.* .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

FROM build-base AS build-openssl
ARG OPENSSL_VERSION=3.0.1
ARG TARGETPLATFORM
ENV GO_OPENSSL_VERSION_OVERRIDE=${OPENSSL_VERSION}
COPY scripts/openssl.sh /bin/openssl.sh
RUN --mount=type=bind,from=osxcross,src=/xx-sdk,target=/xx-sdk,rw <<EOT
    set -ex
    [ "$(xx-info os)" = "windows" ] && prefix="/usr/$(xx-info)" || prefix="/usr"
    if [ "$(xx-info arch)" = "ppc64le" ]; then
        export XX_CC_PREFER_LINKER=ld
    fi
    CC=$(xx-clang --print-target-triple)-clang CXX=$(xx-clang --print-target-triple)-clang \
    AR=$(xx-clang --print-target-triple)-ar RANLIB=$(xx-clang --print-target-triple)-ranlib \
    STRIP=$(xx-clang --print-target-triple)-strip OBJDUMP=$(xx-clang --print-target-triple)-objdump \
    /bin/openssl.sh --dev --prefix "$prefix" --destdir "/build" "${OPENSSL_VERSION}"
EOT

FROM build-openssl AS build
ARG GO_BUILDTAGS
ARG GO_BUILDFLAGS
ARG GO_LDFLAGS
# GO_LINKMODE defines if static or dynamic binary should be produced
ARG GO_LINKMODE=static
ARG GO_STRIP=1
ARG TARGETPLATFORM
RUN --mount=type=bind,target=. \
    --mount=type=cache,target=/root/.cache \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=bind,from=osxcross,src=/xx-sdk,target=/xx-sdk,rw <<EOT
    set -ex
    cp -rv /build/* $(xx-info sysroot)
    if [ "$(xx-info os)" = "windows" ]; then
        prefix="$(xx-info sysroot)usr/$(xx-info)"
        export CGO_LDFLAGS="-lwsock32 -lws2_32"
    else
        prefix="$(xx-info sysroot)usr"
    fi
    if [ "$(xx-info arch)" = "ppc64le" ]; then
        export XX_CC_PREFER_LINKER=ld
    fi
    if [ -n "$GO_STRIP" ]; then
        GO_LDFLAGS="${GO_LDFLAGS} -s -w"
    fi
    if [ "$GO_LINKMODE" = "static" ] && [ "$(xx-info os)" != "darwin" ]; then
        GO_LDFLAGS="${GO_LDFLAGS} -linkmode external -extldflags=-static"
    fi
    export PKG_CONFIG_PATH=$prefix/lib/pkgconfig
    pkg=github.com/matteoarella/pedersen
    version=$(git describe --match 'v[0-9]*' --dirty='.m' --always --tags)
    GO_LDFLAGS="${GO_LDFLAGS} -X ${pkg}/internal.Version=${version}"

    xx-go --wrap
    go build $([ ! -n "$GO_BUILDTAGS" ] || echo "-tags ${GO_BUILDTAGS}") ${GO_BUILDFLAGS} \
        -trimpath -ldflags "${GO_LDFLAGS}" -o /out/pedersen ./cmd/pedersen
    xx-verify $([ "$GO_LINKMODE" != "static" ] || echo "--static") /out/pedersen
EOT

FROM build-base AS lint
ARG GO_BUILDTAGS
RUN --mount=type=bind,target=. \
    --mount=type=cache,target=/root/.cache \
    --mount=from=golangci-lint,source=/usr/bin/golangci-lint,target=/usr/bin/golangci-lint \
    golangci-lint run $([ ! -n "$GO_BUILDTAGS" ] || echo "--build-tags ${GO_BUILDTAGS}") ./...

FROM --platform=${BUILDPLATFORM} build-openssl AS test
ARG GO_BUILDTAGS
ARG GO_TESTFLAGS
ARG TARGETPLATFORM
RUN --mount=type=bind,target=. \
    --mount=type=cache,target=/root/.cache \
    --mount=type=cache,target=/go/pkg/mod <<EOT
    set -ex
    cp -rv /build/* $(xx-info sysroot)
    if [ "$(xx-info os)" = "windows" ]; then
        prefix="$(xx-info sysroot)usr/$(xx-info)"
        export CGO_LDFLAGS="-lwsock32 -lws2_32"
    else
        prefix="$(xx-info sysroot)usr"
    fi
    if [ "$(xx-info arch)" = "ppc64le" ]; then
        export XX_CC_PREFER_LINKER=ld
    fi
    export PKG_CONFIG_PATH=$prefix/lib/pkgconfig
    export QEMU_LD_PREFIX=$(xx-info sysroot)
    xx-go --wrap
    if [ "$(xx-info os)" = "linux" ]; then
        libdir=$($(go env PKG_CONFIG) --variable=libdir libcrypto)
        echo "/lib:/usr/local/lib:/usr/lib:$libdir" > /etc/ld-musl-$(xx-info march).path
    fi
    go test $([ ! -n "$GO_BUILDTAGS" ] || echo "-tags ${GO_BUILDTAGS}") ${GO_TESTFLAGS} -v ./...
EOT

FROM base AS license-validate
ARG LICENSE_FILES
RUN --mount=type=bind,target=. \
    --mount=from=addlicense,source=/app/addlicense,target=/usr/bin/addlicense \
    find . -regex "${LICENSE_FILES}" | xargs addlicense -check -c 'Pedersen authors' -l mit -v

FROM --platform=${BUILDPLATFORM} node:${NODE_VERSION}-alpine AS docs-build
ARG ORGANIZATION_NAME=matteoarella
ARG REPO_NAME=pedersen
ARG REPO_URL=https://github.com/matteoarella/pedersen
ARG DOCS_URL=https://matteoarella.github.io
ARG DOCS_EDIT_URL=https://github.com/matteoarella/pedersen/tree/master/docs/
ARG VERSION

ENV ORGANIZATION_NAME=${ORGANIZATION_NAME}
ENV REPO_NAME=${REPO_NAME}
ENV REPO_URL=${REPO_URL}
ENV DOCS_URL=${DOCS_URL}
ENV DOCS_EDIT_URL=${DOCS_EDIT_URL}
ENV NODE_ENV=production

WORKDIR /docs
COPY ./ .
RUN apk update && apk add --no-cache git
RUN --mount=type=secret,id=ALGOLIA_APP_ID \
    --mount=type=secret,id=ALGOLIA_SEARCH_API_KEY \
    --mount=type=secret,id=ALGOLIA_INDEX_NAME <<EOT
    set -e
    export ALGOLIA_APP_ID=$(cat /run/secrets/ALGOLIA_APP_ID)
    export ALGOLIA_SEARCH_API_KEY=$(cat /run/secrets/ALGOLIA_SEARCH_API_KEY)
    export ALGOLIA_INDEX_NAME=$(cat /run/secrets/ALGOLIA_INDEX_NAME)
    if [ -n "${VERSION}" ]; then
        export VERSION=${VERSION}
    else
        export VERSION=$(git describe --match "v[0-9]*" --dirty='.m' --tags --always)
    fi

    cd docs
    yarn install --frozen-lockfile && yarn build
EOT

FROM scratch AS docs-release
COPY --from=docs-build /docs/docs/build/ /

FROM scratch AS binary-unix
COPY --link --from=build /out/pedersen /
FROM binary-unix AS binary-darwin
FROM binary-unix AS binary-linux
FROM scratch AS binary-windows
COPY --link --from=build /out/pedersen /pedersen.exe
FROM binary-$TARGETOS AS binary
# enable scanning for this stage
ARG BUILDKIT_SBOM_SCAN_STAGE=true

FROM --platform=$BUILDPLATFORM alpine AS releaser
WORKDIR /work
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
RUN --mount=from=binary <<EOT
    set -ex
    mkdir -p /out
    # TODO: should just use standard arch
    TARGETARCH=$([ "$TARGETARCH" = "amd64" ] && echo "x86_64" || echo "$TARGETARCH")
    TARGETARCH=$([ "$TARGETARCH" = "arm64" ] && echo "aarch64" || echo "$TARGETARCH")
    cp pedersen* "/out/pedersen-${TARGETOS}-${TARGETARCH}${TARGETVARIANT}$(ls pedersen* | sed -e 's/^pedersen//')"
EOT

FROM scratch AS release
COPY --from=releaser /out/ /

FROM scratch as image
COPY --from=build /out/pedersen /pedersen
ENTRYPOINT [ "/pedersen" ]
