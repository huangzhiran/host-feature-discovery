#!/bin/bash
HFD_VERSION=`git describe --tags --dirty --always`
docker build . --build-arg HFD_VERSION=${HFD_VERSION} -t \
    huangzhiran/host-feature-discovery:${HFD_VERSION}
