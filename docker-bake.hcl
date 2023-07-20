// Copyright (c) Pedersen authors.
// 
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

variable "GO_VERSION" {
  # default ARG value set in Dockerfile
  default = null
}

variable "OPENSSL_VERSION" {
  default = null
}

variable "GO_BUILDTAGS" {
  default = null
}

variable "GO_BUILDFLAGS" {
  default = null
}

variable "GO_LINKMODE" {
  default = null
}

variable "GO_STRIP" {
  default = null
}

variable "GO_TESTFLAGS" {
  default = null
}

# Defines the output folder to override the default behavior.
# See Makefile for details, this is generally only useful for
# the packaging scripts and care should be taken to not break
# them.
variable "DESTDIR" {
  default = ""
}

variable "REPO_NAME" {
  default = null
}

variable "REPO_URL" {
  default = null
}

variable "ORGANIZATION_NAME" {
  default = null
}

variable "DOCS_URL" {
  default = null
}

variable "DOCS_EDIT_URL" {
  default = null
}

variable "VERSION" {
  default = null
}

function "outdir" {
  params = [defaultdir]
  result = DESTDIR != "" ? DESTDIR : "${defaultdir}"
}

target "_common" {
  args = {
    GO_VERSION = GO_VERSION
    OPENSSL_VERSION = OPENSSL_VERSION
    GO_BUILDTAGS = GO_BUILDTAGS
    GO_BUILDFLAGS = GO_BUILDFLAGS
    BUILDKIT_CONTEXT_KEEP_GIT_DIR = 1
  }
}

group "default" {
  targets = ["binary"]
}

group "validate" {
  targets = ["lint", "license-validate"]
}

target "lint" {
  inherits = ["_common"]
  target = "lint"
  output = ["type=cacheonly"]
}

target "license-validate" {
  target = "license-validate"
  output = ["type=cacheonly"]
}

target "test" {
  inherits = ["_common"]
  target = "test"
  args = {
    GO_TESTFLAGS = GO_TESTFLAGS
  }
  output = ["type=cacheonly"]
  platforms = ["local"]
}

target "test-cross" {
  inherits = ["test"]
  platforms = [
    "linux/amd64",
    "linux/arm/v6",
    "linux/arm/v7",
    "linux/arm64",
    "linux/ppc64le",
    "linux/riscv64",
    "linux/s390x",
  ]
}

target "binary" {
  inherits = ["_common"]
  target = "binary"
  args = {
    OPENSSL_VERSION = OPENSSL_VERSION
    GO_LINKMODE = GO_LINKMODE
    GO_STRIP = GO_STRIP
  }
  output = [outdir("./bin/build")]
  platforms = ["local"]
}

target "binary-cross" {
  inherits = ["binary"]
  platforms = [
    "darwin/amd64",
    "darwin/arm64",
    "linux/amd64",
    "linux/arm/v6",
    "linux/arm/v7",
    "linux/arm64",
    "linux/ppc64le",
    "linux/riscv64",
    "linux/s390x",
    "windows/amd64",
  ]
}

target "release" {
  inherits = ["binary-cross"]
  target = "release"
  output = [outdir("./bin/release")]
}

target "docs" {
  target = "docs-release"
  args = {
    ORGANIZATION_NAME = ORGANIZATION_NAME
    REPO_NAME = REPO_NAME
    REPO_URL = REPO_URL
    DOCS_URL = DOCS_URL
    DOCS_EDIT_URL = DOCS_EDIT_URL
    VERSION = VERSION
  }
  secret = [
    "type=env,id=ALGOLIA_APP_ID",
    "type=env,id=ALGOLIA_SEARCH_API_KEY",
    "type=env,id=ALGOLIA_INDEX_NAME",
  ]
  output = [outdir("./bin/docs")]
}

target "docker-metadata-action" {}

target "image-cross" {
  inherits = ["docker-metadata-action", "release"]
  target = "image"
  output = ["type=image"]
}
