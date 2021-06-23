# Copyright 2019-present Open Networking Foundation
#
# SPDX-License-Identifier: Apache-2.0
#

FROM golang:1.14.4-stretch AS sim

LABEL maintainer="ONF <omec-dev@opennetworking.org>"

RUN apt-get update
RUN apt-get -y install vim 
RUN cd $GOPATH/src && mkdir -p simapp
COPY . $GOPATH/src/simapp
RUN cd $GOPATH/src/simapp && go install 

FROM sim AS simapp
WORKDIR /simapp
RUN mkdir -p /simapp/bin
COPY --from=sim $GOPATH/bin/* /simapp/bin/
WORKDIR /simapp
