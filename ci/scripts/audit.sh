#!/bin/bash -eux

export cwd=$(pwd)

pushd $cwd/dp-search-reindex-api
  make audit
popd