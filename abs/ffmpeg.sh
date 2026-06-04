#!/bin/sh -e

DIR=$( cd "$( dirname "$0" )" && pwd )

LIBS=$(echo \
  ${DIR}/lib \
  ${DIR}/usr/lib \
  ${DIR}/usr/local/lib | tr ' ' ':')

LOADER=$(ls ${DIR}/lib/ld-musl-*.so* 2>/dev/null | head -1)

exec "${LOADER}" --library-path "${LIBS}" ${DIR}/usr/bin/ffmpeg "$@"
