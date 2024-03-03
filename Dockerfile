# Copyright 2019-present Open Networking Foundation
# Copyright 2024-present Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0
#

FROM golang:1.22.0-bookworm AS sim

LABEL maintainer="ONF <omec-dev@opennetworking.org>"

RUN apt-get update && \
    apt-get -y install --no-install-recommends \
    vim

WORKDIR $GOPATH/src/simapp
COPY . .
RUN CGO_ENABLED=0 go install

FROM alpine:3.19 AS simapp

# Install debug tools ~ 50MB (if DEBUG_TOOLS is set to true)
RUN if [ "$DEBUG_TOOLS" = "true" ]; then \
        apk update && apk add --no-cache -U gcompat vim strace net-tools curl netcat-openbsd bind-tools bash; \
        fi

WORKDIR /simapp/bin

# Copy executable
COPY --from=sim /go/bin/* .
