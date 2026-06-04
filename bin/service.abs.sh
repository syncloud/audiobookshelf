#!/bin/bash -e

source ${SNAP_DATA}/config/abs.env
export CONFIG_PATH METADATA_PATH

export NODE_ENV=production
export SOURCE=docker
export PORT=3333
export HOST=unix/${SNAP_DATA}/audiobookshelf.sock
export ROUTER_BASE_PATH=/audiobookshelf
export NODE_EXTRA_CA_CERTS=/var/snap/platform/current/syncloud.ca.crt

source ${SNAP}/node/runtime-env.sh

/bin/rm -f ${SNAP_DATA}/audiobookshelf.sock

cd ${SNAP}/abs
exec ${SNAP}/node/node.sh ${SNAP}/abs/index.js
