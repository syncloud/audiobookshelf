#!/bin/bash -e

source ${SNAP_DATA}/config/abs.env

/bin/rm -f ${SNAP_DATA}/audiobookshelf.sock

cd ${SNAP}/abs
exec ${SNAP}/node/node.sh ${SNAP}/abs/index.js
