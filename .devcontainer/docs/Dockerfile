# syntax=docker/dockerfile:1

# Copyright (c) Pedersen authors.
# 
# Use of this source code is governed by an MIT-style
# license that can be found in the LICENSE file or at
# https://opensource.org/licenses/MIT.

ARG NODE_VERSION=20

FROM node:${NODE_VERSION}-alpine

EXPOSE 3000

WORKDIR /src

RUN apk add --no-cache git

ENV NODE_ENV=dev
