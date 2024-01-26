# Copyright 2019-present Open Networking Foundation
#
# SPDX-License-Identifier: Apache-2.0
#

FROM golang:1.18.3-stretch AS sim

LABEL maintainer="ONF <omec-dev@opennetworking.org>"

RUN apt-get update && apt-get -y install vim
RUN cd $GOPATH/src && mkdir -p simapp
COPY . $GOPATH/src/simapp
RUN cd $GOPATH/src/simapp && go install 

FROM alpine:3.16 AS simapp

#RUN apk update && apk add -U libc6-compat vim strace net-tools curl netcat-openbsd bind-tools bash
RUN apk update && apk add -U gcompat vim strace net-tools curl netcat-openbsd bind-tools bash

WORKDIR /simapp
RUN mkdir -p /simapp/bin

# Copy executable
COPY --from=sim /go/bin/* /simapp/bin/
WORKDIR /simapp
