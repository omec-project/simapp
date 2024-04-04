# Copyright 2019-present Open Networking Foundation
# Copyright 2024-present Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0
#

FROM golang:1.22.2-bookworm AS builder

LABEL maintainer="Aether SD-Core <dev@aetherproject.org>"

RUN apt-get update && \
    apt-get -y install --no-install-recommends \
    vim && \
    apt-get clean

WORKDIR $GOPATH/src/simapp
COPY . .
RUN make all

FROM alpine:3.19 AS simapp

LABEL description="Aether open source 5G Core Network" \
    version="Stage 3"

ARG DEBUG_TOOLS

# Install debug tools ~ 50MB (if DEBUG_TOOLS is set to true)
RUN if [ "$DEBUG_TOOLS" = "true" ]; then \
        apk update && apk add --no-cache -U gcompat vim strace net-tools curl netcat-openbsd bind-tools bash; \
        fi

WORKDIR /simapp/bin

# Copy executable
COPY --from=builder /go/src/simapp/bin/* .
