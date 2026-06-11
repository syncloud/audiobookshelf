#!/bin/bash -e

source ${SNAP_DATA}/config/abs.env

export FILE_BROWSER_ROOT=$(realpath "${FILE_BROWSER_ROOT}" 2>/dev/null || echo "${FILE_BROWSER_ROOT}")

/bin/rm -f ${SNAP_DATA}/audiobookshelf.sock

cd ${SNAP}/abs
exec ${SNAP}/node/node.sh ${SNAP}/abs/index.js
