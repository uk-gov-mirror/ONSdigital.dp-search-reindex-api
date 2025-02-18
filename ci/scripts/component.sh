#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/dp-search-reindex-api
#  This is required to tell mongo which binary to download, without this
#  mongo tries to download a binary for debian which doesn't work on the container
  export MEMONGO_DOWNLOAD_URL=https://fastdl.mongodb.org/linux/mongodb-linux-x86_64-ubuntu1804-4.0.23.tgz
  make test-component
popd
