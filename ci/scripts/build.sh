#!/bin/bash -eux

pushd dp-search-reindex-api
  make build
  cp build/dp-search-reindex-api Dockerfile.concourse ../build
popd
