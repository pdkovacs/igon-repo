#!/bin/bash

parent_dir=$(dirname $0)

set -x
cp -a "${parent_dir}"/dist "${parent_dir}"/docker/ \
&& docker build -t iconrepo "${parent_dir}"/docker
set +x
