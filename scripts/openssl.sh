#!/usr/bin/env bash

set -euxo pipefail

BASEDIR=$(dirname $(dirname $(realpath "$0")))
ME=$(basename "$0")

usage() {
    cat <<EOF
Usage: $ME [OPTIONS]... OPENSSL-VERSION

Download and install OpenSSL with version OPENSSL-VERSION.

Options:
    -prefix | --prefix PREFIX      prefix for OpenSSL
    -destdir | --destdir DESTDIR   destdir for OpenSSL
    -dev | --dev                   install OpenSSL headers
    -h                             show this help and exit
EOF
}

dev=false
target=""
prefix="/usr/local"
destdir=""

if [ ! -z "${TARGETPLATFORM+x}" ]; then
    case "$TARGETPLATFORM" in
    "darwin/amd64") 
        target="darwin64-x86_64" ;;
    "darwin/arm64") 
        target="darwin64-arm64" ;;
    "linux/amd64") 
        target="linux-x86_64" ;;
    "linux/arm/v6") 
        target="linux-generic32" ;;
    "linux/arm/v7") 
        target="linux-generic32" ;;
    "linux/arm64") 
        target="linux-generic64" ;;
    "linux/ppc64le") 
        target="linux-ppc64le" ;;
    "linux/riscv64") 
        target="linux64-riscv64" ;;
    "linux/s390x") 
        target="linux64-s390x" ;;
    "windows/amd64") 
        target="mingw64" ;;
    *)
        target="" ;;
    esac
fi

while true; do
    case "$1" in
    -h)
        usage; exit 0 ;;
    -dev | --dev)
        dev=true; shift ;;
    -prefix | --prefix)
        if [ ! -z "$2" ]; then
            prefix="$2";
        fi
        shift 2 ;;
    -destdir | --destdir)
        destdir="$2"; shift 2 ;;
    *)
        version="$1"; break ;;
    esac
done

case "$version" in
    "1.0.2")
        tag="OpenSSL_1_0_2u"
        sha256="82fa58e3f273c53128c6fe7e3635ec8cda1319a10ce1ad50a987c3df0deeef05"
        config="shared"
        make="build_libs"
        ;;
    "1.1.0")
        tag="OpenSSL_1_1_0l"
        sha256="e2acf0cf58d9bff2b42f2dc0aee79340c8ffe2c5e45d3ca4533dd5d4f5775b1d"
        config="shared"
        make="build_libs"
        ;;
    "1.1.1")
        tag="OpenSSL_1_1_1m"
        sha256="36ae24ad7cf0a824d0b76ac08861262e47ec541e5d0f20e6d94bab90b2dab360"
        config="shared"
        make="build_libs"
        ;;
    "3.0.1")
        tag="openssl-3.0.1";
        sha256="2a9dcf05531e8be96c296259e817edc41619017a4bf3e229b4618a70103251d5"
        config="shared"
        make="build_libs"
        ;;
    *)
        echo >&2 "error: unsupported OpenSSL version '$version'"
        exit 1 ;;
esac

openssl_install_from_sources() {
    local version="$1"
    local tag="$2"
    local sha256="$3"
    local config="$4"
    local make="$5"
    local sourcedir="$6"
    local destdir="$7"

    mkdir -p $sourcedir
    cd $sourcedir
    wget -O "$tag.tar.gz" "https://github.com/openssl/openssl/archive/refs/tags/$tag.tar.gz"
    echo "$sha256 $tag.tar.gz" | sha256sum -c -
    rm -rf "openssl-$tag"
    tar -xzf "$tag.tar.gz"

    rm -rf "openssl-$version"
    mv "openssl-$tag" "openssl-$version"

    cd "openssl-$version"
    ./Configure $config
    make DESTDIR="$destdir" -j$(nproc) "$make"
}

openssl_install_alias_libs() {
    local version="$1"
    local sourcedir="$2"
    local destdir="$3"

    libdir=$(PKG_CONFIG_PATH="$sourcedir/openssl-$version" pkg-config --variable=libdir libcrypto)
    libdir="$destdir$libdir"

    mkdir -p $libdir

    ls -l $sourcedir/openssl-$version/libcrypto*

    find "$sourcedir/openssl-$version" -name "libcrypto*.a" -type f -print0 | while read -d $'\0' lib; do
        cp -Lv $lib "$libdir"
    done

    find "$sourcedir/openssl-$version" \( -name "libcrypto*.so" -o -name "libcrypto*.dylib" -o \
    -name "libcrypto*.dll" \) -print0 | while read -d $'\0' lib; do
        name=$(basename "$lib")
        cp -Lv $lib "$libdir/$name.${version}"
    done
}

openssl_install_dev() {
    local version="$1"
    local sourcedir="$2"
    local destdir="$3"

    prefix=$(PKG_CONFIG_PATH="$sourcedir/openssl-$version" pkg-config --variable=prefix libcrypto)
    includedir=$(PKG_CONFIG_PATH="$sourcedir/openssl-$version" pkg-config --variable=includedir libcrypto)
    includedir="$destdir$includedir"

    mkdir -p $includedir/openssl "$destdir${prefix}/lib/pkgconfig"

    cp -Lr "$sourcedir/openssl-$version/include/openssl" "$includedir"
    cp "$sourcedir/openssl-$version/libcrypto.pc" "$destdir${prefix}/lib/pkgconfig"
}

sourcedir=/usr/local/src

if [ -n "$prefix" ]; then
    config="$config --prefix=$prefix"
fi
if [ -n "$target" ]; then
    config="$target $config"
fi

openssl_install_from_sources "$version" "$tag" "$sha256" "$config" "$make" "$sourcedir" "$destdir"
openssl_install_alias_libs "$version" "$sourcedir" "$destdir"

if $dev; then
    openssl_install_dev "$version" "$sourcedir" "$destdir"
fi
