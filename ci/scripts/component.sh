#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/dp-search-reindex-api
  make test-component
popd
