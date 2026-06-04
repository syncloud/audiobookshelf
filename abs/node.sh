#!/bin/bash -e

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

LIBS=$(echo \
  ${DIR}/lib \
  ${DIR}/usr/lib \
  ${DIR}/usr/local/lib | tr ' ' ':')

LOADER=$(ls ${DIR}/lib/ld-musl-*.so* 2>/dev/null | head -1)

exec "${LOADER}" --library-path "${LIBS}" ${DIR}/usr/local/bin/node "$@"
